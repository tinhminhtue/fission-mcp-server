package fission

import (
	"k8s.io/client-go/dynamic"
)

// Service provides operations for managing Fission functions
type Service struct {
	client    dynamic.Interface
	routerURL string
}

// NewService creates a new Fission service
func NewService(client dynamic.Interface, routerURL string) *Service {
	if routerURL == "" {
		routerURL = "http://router.fission.svc.cluster.local"
	}
	return &Service{
		client:    client,
		routerURL: routerURL,
	}
}
