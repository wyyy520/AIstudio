// Package template provides the centralized Template Engine for AIStudio.
//
// Per silu.md Chapter 4, the Template Engine is responsible for:
//   - Template discovery (scanning directories for template.json)
//   - Template copying (entire directory trees)
//   - Variable substitution ({{var}} -> values)
//   - Template versioning and validation
//   - Plugin-based template extension
//
// Generators call this engine instead of embedding their own templates.
package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Template describes a complete engineering project template.
// Each template is a directory containing:
//   - template.json metadata file
//   - .tpl files (Go template syntax with {{var}} placeholders)
//   - regular files (copied as-is)
type Template struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	Domain        string            `json:"domain"`
	Description   string            `json:"description,omitempty"`
	Author        string            `json:"author,omitempty"`
	Path          string            `json:"path"`
	Variables     []TemplateVar     `json:"variables,omitempty"`
	Requires      []string          `json:"requires,omitempty"`
	DefaultConfig map[string]any    `json:"default_config,omitempty"`
	EntryPoints   []string          `json:"entry_points,omitempty"`
}

// TemplateVar describes a variable exposed by a template.
type TemplateVar struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Default     any    `json:"default,omitempty"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
}

// Engine is the central template management and rendering system.
type Engine struct {
	mu        sync.RWMutex
	templates map[string]*Template
	roots     []string
}

// NewEngine creates a new Template Engine.
func NewEngine(templateRoots ...string) *Engine {
	return &Engine{
		templates: make(map[string]*Template),
		roots:     templateRoots,
	}
}

// Discover scans registered root directories and registers all found templates.
func (e *Engine) Discover(targetDir string) error {
	dirs := e.roots
	if targetDir != "" {
		dirs = []string{targetDir}
	}

	for _, root := range dirs {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // skip unreadable
			}
			if info.IsDir() {
				return nil
			}
			if filepath.Base(path) == "template.json" {
				tmpl, err := loadTemplateFromDir(filepath.Dir(path))
				if err != nil {
					return nil // skip invalid
				}
				e.mu.Lock()
				e.templates[tmpl.ID] = tmpl
				e.mu.Unlock()
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("discover templates in %s: %w", root, err)
		}
	}
	return nil
}

// Get returns a template by ID.
func (e *Engine) Get(id string) (*Template, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	t, ok := e.templates[id]
	return t, ok
}

// List returns all registered templates, optionally filtered by domain.
func (e *Engine) List(domain string) []*Template {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []*Template
	for _, t := range e.templates {
		if domain == "" || t.Domain == domain {
			result = append(result, t)
		}
	}
	return result
}

// Register adds a template from a parsed Template object.
func (e *Engine) Register(tmpl *Template) error {
	if tmpl.ID == "" {
		return fmt.Errorf("template ID is required")
	}
	if tmpl.Path == "" {
		return fmt.Errorf("template path is required")
	}
	if _, err := os.Stat(tmpl.Path); err != nil {
		return fmt.Errorf("template path does not exist: %s", tmpl.Path)
	}

	e.mu.Lock()
	e.templates[tmpl.ID] = tmpl
	e.mu.Unlock()
	return nil
}

// Validate checks if a template is complete and valid.
func (e *Engine) Validate(id string) error {
	tmpl, ok := e.Get(id)
	if !ok {
		return fmt.Errorf("template %q not found", id)
	}

	info, err := os.Stat(tmpl.Path)
	if err != nil {
		return fmt.Errorf("template directory %s: %w", tmpl.Path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("template path %s is not a directory", tmpl.Path)
	}

	// Check that at least one file exists in the template directory
	entries, err := os.ReadDir(tmpl.Path)
	if err != nil {
		return fmt.Errorf("read template directory: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("template %q is empty", id)
	}

	return nil
}

// loadTemplateFromDir reads a template entry from a directory containing template.json.
func loadTemplateFromDir(dir string) (*Template, error) {
	metaPath := filepath.Join(dir, "template.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("read template.json: %w", err)
	}

	var tmpl Template
	if err := jsonUnmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("parse template.json: %w", err)
	}

	if tmpl.ID == "" {
		tmpl.ID = filepath.Base(dir)
	}
	tmpl.Path = dir
	return &tmpl, nil
}

// Count returns the number of registered templates.
func (e *Engine) Count() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.templates)
}

// sanitizeName cleans a string for use as a directory name.
func sanitizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	var result []rune
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result = append(result, r)
		}
	}
	if len(result) == 0 {
		return "unnamed"
	}
	return string(result)
}
