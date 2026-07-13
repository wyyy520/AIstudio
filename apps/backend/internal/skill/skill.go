// Package skill provides Workflow Template management.
//
// A Skill is NOT a prompt.
// A Skill is a Workflow Template — a reusable, parameterized workflow definition.
//
// Skills are used to quickly generate complete workflows for common AI tasks.
// Skills never generate code. They only produce workflow.json.
//
// Example:
//   YOLO Detection Skill → generates a YOLO training workflow
//   Transformer Classification Skill → generates a transformer classification workflow
//   Traffic Simulation Skill → generates a traffic simulation workflow
package skill

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aistudio/backend/internal/workflow"
)

// ============================================================================
// Skill — Workflow Template
// ============================================================================

// Skill is a reusable workflow template.
// Skills generate workflow.json, not code.
type Skill interface {
	// ID returns the unique skill identifier.
	ID() string

	// Name returns the human-readable skill name.
	Name() string

	// Description returns a description of what this skill does.
	Description() string

	// Category returns the skill category (e.g., "vision", "nlp", "traffic").
	Category() string

	// Version returns the skill version.
	Version() string

	// Parameters returns the parameter definitions for this skill.
	Parameters() []Parameter

	// Apply generates a workflow from the skill with the given parameters.
	// This is the only way a Skill produces output — it returns a Workflow.
	Apply(params map[string]interface{}) (*workflow.Workflow, error)

	// Template returns the raw workflow template.
	Template() *Template
}

// Parameter defines a skill parameter.
type Parameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // string, number, boolean, enum
	Label       string      `json:"label"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Options     []string    `json:"options,omitempty"` // For enum type
	Min         *float64    `json:"min,omitempty"`
	Max         *float64    `json:"max,omitempty"`
}

// Template is the raw workflow template with variable placeholders.
type Template struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Category     string            `json:"category"`
	Version      string            `json:"version"`
	Parameters   []Parameter       `json:"parameters"`
	WorkflowJSON json.RawMessage   `json:"workflow"` // Template with ${variable} placeholders
}

// ============================================================================
// Skill Manager
// ============================================================================

// Manager manages skill templates.
type Manager struct {
	registry map[string]Skill
}

// NewManager creates a new Skill Manager.
func NewManager() *Manager {
	return &Manager{
		registry: make(map[string]Skill),
	}
}

// Register registers a skill.
func (m *Manager) Register(skill Skill) error {
	id := skill.ID()
	if _, exists := m.registry[id]; exists {
		return fmt.Errorf("skill already registered: %s", id)
	}
	m.registry[id] = skill
	return nil
}

// MustRegister registers a skill, panicking on error.
func (m *Manager) MustRegister(skill Skill) {
	if err := m.Register(skill); err != nil {
		panic(err)
	}
}

// Get returns a skill by ID.
func (m *Manager) Get(id string) (Skill, bool) {
	skill, ok := m.registry[id]
	return skill, ok
}

// List returns all registered skills.
func (m *Manager) List() []Skill {
	skills := make([]Skill, 0, len(m.registry))
	for _, s := range m.registry {
		skills = append(skills, s)
	}
	return skills
}

// ListByCategory returns skills filtered by category.
func (m *Manager) ListByCategory(category string) []Skill {
	var result []Skill
	for _, s := range m.registry {
		if s.Category() == category {
			result = append(result, s)
		}
	}
	return result
}

// Apply applies a skill with the given parameters.
func (m *Manager) Apply(skillID string, params map[string]interface{}) (*workflow.Workflow, error) {
	skill, ok := m.registry[skillID]
	if !ok {
		return nil, fmt.Errorf("skill not found: %s", skillID)
	}
	return skill.Apply(params)
}

// LoadFromFile loads a skill template from a JSON file.
func (m *Manager) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read skill file: %w", err)
	}
	return m.LoadFromJSON(data)
}

// LoadFromJSON loads a skill template from JSON bytes.
func (m *Manager) LoadFromJSON(data []byte) error {
	var tmpl Template
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return fmt.Errorf("failed to parse skill template: %w", err)
	}
	return m.Register(&templateSkill{template: &tmpl})
}

// LoadFromDir loads all skill templates from a directory.
func (m *Manager) LoadFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read skill directory: %w", err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".json" {
			path := filepath.Join(dir, entry.Name())
			if err := m.LoadFromFile(path); err != nil {
				return fmt.Errorf("failed to load skill %s: %w", entry.Name(), err)
			}
		}
	}
	return nil
}

// Unregister removes a skill by ID.
func (m *Manager) Unregister(id string) {
	delete(m.registry, id)
}

// Count returns the number of registered skills.
func (m *Manager) Count() int {
	return len(m.registry)
}

// ============================================================================
// Template Skill Implementation
// ============================================================================

// templateSkill is a Skill backed by a JSON template with variable substitution.
type templateSkill struct {
	template *Template
}

func (s *templateSkill) ID() string             { return s.template.ID }
func (s *templateSkill) Name() string            { return s.template.Name }
func (s *templateSkill) Description() string     { return s.template.Description }
func (s *templateSkill) Category() string        { return s.template.Category }
func (s *templateSkill) Version() string         { return s.template.Version }
func (s *templateSkill) Parameters() []Parameter { return s.template.Parameters }
func (s *templateSkill) Template() *Template     { return s.template }

func (s *templateSkill) Apply(params map[string]interface{}) (*workflow.Workflow, error) {
	// Validate required parameters
	for _, p := range s.template.Parameters {
		if p.Required {
			if _, exists := params[p.Name]; !exists {
				return nil, fmt.Errorf("required parameter '%s' is missing", p.Name)
			}
		}
	}

	// Apply defaults for missing parameters
	mergedParams := make(map[string]interface{})
	for _, p := range s.template.Parameters {
		if val, exists := params[p.Name]; exists {
			mergedParams[p.Name] = val
		} else if p.Default != nil {
			mergedParams[p.Name] = p.Default
		}
	}

	// Perform variable substitution on the workflow JSON
	workflowJSON := string(s.template.WorkflowJSON)
	substituted := substituteVariables(workflowJSON, mergedParams)

	// Parse the substituted workflow
	var wf workflow.Workflow
	if err := json.Unmarshal([]byte(substituted), &wf); err != nil {
		return nil, fmt.Errorf("failed to parse generated workflow: %w", err)
	}

	return &wf, nil
}

// substituteVariables replaces ${variable} placeholders with values.
func substituteVariables(template string, params map[string]interface{}) string {
	result := template
	for key, val := range params {
		placeholder := fmt.Sprintf("${%s}", key)
		var strVal string
		switch v := val.(type) {
		case string:
			strVal = v
		case float64:
			strVal = fmt.Sprintf("%v", v)
		case int:
			strVal = fmt.Sprintf("%d", v)
		case bool:
			strVal = fmt.Sprintf("%t", v)
		default:
			strVal = fmt.Sprintf("%v", v)
		}
		result = replaceAll(result, placeholder, strVal)
	}
	return result
}

func replaceAll(s, old, new string) string {
	for i := 0; i < 100; i++ {
		prev := s
		s = replaceOne(s, old, new)
		if s == prev {
			break
		}
	}
	return s
}

func replaceOne(s, old, new string) string {
	idx := strings.Index(s, old)
	if idx < 0 {
		return s
	}
	return s[:idx] + new + s[idx+len(old):]
}