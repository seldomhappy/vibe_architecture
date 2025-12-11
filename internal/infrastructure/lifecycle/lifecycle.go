package lifecycle

import (
	"context"
	"fmt"
)

// Component represents a lifecycle-managed component
type Component interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Manager manages component lifecycle
type Manager struct {
	components []Component
}

// NewManager creates a new lifecycle manager
func NewManager() *Manager {
	return &Manager{
		components: make([]Component, 0),
	}
}

// Register registers a component
func (m *Manager) Register(component Component) {
	m.components = append(m.components, component)
}

// Start starts all components
func (m *Manager) Start(ctx context.Context) error {
	for i, component := range m.components {
		if err := component.Start(ctx); err != nil {
			return fmt.Errorf("failed to start component %d: %w", i, err)
		}
	}
	return nil
}

// Stop stops all components in reverse order
func (m *Manager) Stop(ctx context.Context) error {
	for i := len(m.components) - 1; i >= 0; i-- {
		if err := m.components[i].Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop component %d: %w", i, err)
		}
	}
	return nil
}
