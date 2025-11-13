package fission

import "k8s.io/client-go/dynamic"

// Service provides operations for managing Fission functions
type Service struct {
	client dynamic.Interface
}

// NewService creates a new Fission service
func NewService(client dynamic.Interface) *Service {
	return &Service{
		client: client,
	}
}
