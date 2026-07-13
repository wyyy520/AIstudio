package compiler

import (
	"fmt"
	"sync"

	"github.com/aistudio/packages/workflow"
)

// GeneratorRegistry manages all registered generators.
type GeneratorRegistry struct {
	mu         sync.RWMutex
	generators map[workflow.Target]Generator
}

// NewGeneratorRegistry creates a new GeneratorRegistry.
func NewGeneratorRegistry() *GeneratorRegistry {
	return &GeneratorRegistry{
		generators: make(map[workflow.Target]Generator),
	}
}

// Register registers a generator for a target.
func (r *GeneratorRegistry) Register(g Generator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	target := g.ID()
	if _, exists := r.generators[target]; exists {
		return fmt.Errorf("generator already registered for target: %s", target)
	}

	r.generators[target] = g
	return nil
}

// MustRegister registers a generator, panicking on duplicate.
func (r *GeneratorRegistry) MustRegister(g Generator) {
	if err := r.Register(g); err != nil {
		panic(err)
	}
}

// Get returns the generator for the given target.
func (r *GeneratorRegistry) Get(target workflow.Target) (Generator, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	g, ok := r.generators[target]
	return g, ok
}

// List returns all registered generators.
func (r *GeneratorRegistry) List() []Generator {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Generator, 0, len(r.generators))
	for _, g := range r.generators {
		result = append(result, g)
	}
	return result
}

// Unregister removes a generator for the given target.
func (r *GeneratorRegistry) Unregister(target workflow.Target) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.generators, target)
}

// Count returns the number of registered generators.
func (r *GeneratorRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.generators)
}

// HasTarget checks if a generator is registered for the target.
func (r *GeneratorRegistry) HasTarget(target workflow.Target) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.generators[target]
	return ok
}