package lifecycle

import (
	"context"
	"fmt"
)

// Service represents a service that can be started and stopped
type Service interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// Manager manages the lifecycle of multiple services
type Manager struct {
	services []Service
	names    []string
}

// New creates a new lifecycle manager
func New() *Manager {
	return &Manager{
		services: make([]Service, 0),
		names:    make([]string, 0),
	}
}

// Register registers a service with the lifecycle manager
func (m *Manager) Register(name string, service Service) {
	m.services = append(m.services, service)
	m.names = append(m.names, name)
}

// StartAll starts all registered services in order
func (m *Manager) StartAll(ctx context.Context) error {
	for i, service := range m.services {
		if err := service.Start(ctx); err != nil {
			return fmt.Errorf("failed to start %s: %w", m.names[i], err)
		}
	}
	return nil
}

// ShutdownAll shuts down all registered services in reverse order
func (m *Manager) ShutdownAll(ctx context.Context) error {
	var lastErr error
	for i := len(m.services) - 1; i >= 0; i-- {
		if err := m.services[i].Shutdown(ctx); err != nil {
			lastErr = fmt.Errorf("failed to shutdown %s: %w", m.names[i], err)
		}
	}
	return lastErr
}
