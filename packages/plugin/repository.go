package plugin

import (
	"fmt"
	"time"
)

// Repository defines the interface for plugin persistence.
type Repository interface {
	Save(p *Plugin) error
	Update(p *Plugin) error
	FindByName(name string) (*Plugin, error)
	FindByID(id string) (*Plugin, error)
	FindAll() ([]*Plugin, error)
	Delete(name string) error
	AutoMigrate() error
}

// InMemoryRepository is an in-memory implementation of Repository.
type InMemoryRepository struct {
	plugins map[string]*Plugin
}

// NewInMemoryRepository creates a new in-memory plugin repository.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		plugins: make(map[string]*Plugin),
	}
}

func (r *InMemoryRepository) Save(p *Plugin) error {
	if p.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if _, exists := r.plugins[p.Name]; exists {
		return fmt.Errorf("plugin already exists: %s", p.Name)
	}
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	r.plugins[p.Name] = p
	return nil
}

func (r *InMemoryRepository) Update(p *Plugin) error {
	if p.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	existing, exists := r.plugins[p.Name]
	if !exists {
		return fmt.Errorf("plugin not found: %s", p.Name)
	}
	p.CreatedAt = existing.CreatedAt
	p.UpdatedAt = time.Now()
	r.plugins[p.Name] = p
	return nil
}

func (r *InMemoryRepository) FindByName(name string) (*Plugin, error) {
	p, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", name)
	}
	return p, nil
}

func (r *InMemoryRepository) FindByID(id string) (*Plugin, error) {
	for _, p := range r.plugins {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, fmt.Errorf("plugin not found by id: %s", id)
}

func (r *InMemoryRepository) FindAll() ([]*Plugin, error) {
	result := make([]*Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		result = append(result, p)
	}
	return result, nil
}

func (r *InMemoryRepository) Delete(name string) error {
	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin not found: %s", name)
	}
	delete(r.plugins, name)
	return nil
}

func (r *InMemoryRepository) AutoMigrate() error {
	return nil
}
