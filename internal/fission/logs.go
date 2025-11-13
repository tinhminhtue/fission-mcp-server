package fission

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetFunctionLogs retrieves logs for a Fission function
func (s *Service) GetFunctionLogs(ctx context.Context, name, namespace string, options LogOptions) (string, error) {
	if namespace == "" {
		namespace = DefaultNamespace
	}

	// Get Kubernetes config
	config, err := s.getKubernetesConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get Kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Find pods running this function
	// Fission functions run in pods with labels like functionName
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("functionName=%s", name),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		// Try alternative label selector
		pods, err = clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("function=%s", name),
		})
		if err != nil {
			return "", fmt.Errorf("failed to list pods: %w", err)
		}
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no pods found for function '%s'", name)
	}

	// Collect logs from all pods
	var allLogs []string
	for _, pod := range pods.Items {
		// Get logs for this pod
		podLogs, err := s.getPodLogs(ctx, clientset, pod.Namespace, pod.Name, options)
		if err != nil {
			// Log error but continue with other pods
			allLogs = append(allLogs, fmt.Sprintf("Error getting logs for pod %s: %v", pod.Name, err))
			continue
		}
		allLogs = append(allLogs, fmt.Sprintf("=== Pod: %s/%s ===\n%s", pod.Namespace, pod.Name, podLogs))
	}

	return strings.Join(allLogs, "\n\n"), nil
}

// getPodLogs retrieves logs for a specific pod
func (s *Service) getPodLogs(ctx context.Context, clientset *kubernetes.Clientset, namespace, podName string, options LogOptions) (string, error) {
	opts := &corev1.PodLogOptions{}

	// Set tail option
	if options.Tail > 0 {
		tailLines := int64(options.Tail)
		opts.TailLines = &tailLines
	}

	// Set since option
	if options.Since != "" {
		// Parse duration (e.g., "5m", "1h")
		duration, err := time.ParseDuration(options.Since)
		if err == nil {
			sinceTime := metav1.NewTime(time.Now().Add(-duration))
			opts.SinceTime = &sinceTime
		}
	}

	// Get logs
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, opts)
	logs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to stream logs: %w", err)
	}
	defer logs.Close()

	// Read logs
	logBytes, err := io.ReadAll(logs)
	if err != nil {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}

	return string(logBytes), nil
}

// getKubernetesConfig gets Kubernetes config from kubeconfig file or in-cluster config
func (s *Service) getKubernetesConfig() (*rest.Config, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Fallback to kubeconfig file
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	if kubeconfigPath := os.Getenv("KUBECONFIG"); kubeconfigPath != "" {
		kubeconfig = kubeconfigPath
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %w", err)
	}

	return config, nil
}

