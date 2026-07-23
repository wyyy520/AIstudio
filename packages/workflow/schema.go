package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// SchemaVersion represents a workflow schema version string.
type SchemaVersion string

// CurrentSchemaVersion is the current workflow schema version.
const CurrentSchemaVersion = "2.0.0"

// SchemaMigrator handles workflow schema version migration.
type SchemaMigrator struct {
	migrations map[string]func(wf *Workflow) error
}

// NewSchemaMigrator creates a new SchemaMigrator with built-in migrations.
func NewSchemaMigrator() *SchemaMigrator {
	m := &SchemaMigrator{
		migrations: make(map[string]func(wf *Workflow) error),
	}
	m.registerDefaults()
	return m
}

// Register adds a migration from one version to another.
func (m *SchemaMigrator) Register(fromVersion string, fn func(wf *Workflow) error) {
	m.migrations[fromVersion] = fn
}

// Migrate migrates a workflow from its current version to the target version.
// Returns true if migration was performed, false if already at target.
func (m *SchemaMigrator) Migrate(wf *Workflow, targetVersion string) (bool, error) {
	if wf.SchemaVersion == targetVersion {
		return false, nil
	}

	current := wf.SchemaVersion
	maxIter := 20
	for i := 0; i < maxIter; i++ {
		if current == targetVersion {
			return true, nil
		}
		fn, ok := m.migrations[current]
		if !ok {
			return false, &MigrationError{
				FromVersion: current,
				ToVersion:   targetVersion,
				Message:     "no migration path available",
			}
		}
		if err := fn(wf); err != nil {
			return false, &MigrationError{
				FromVersion: current,
				ToVersion:   targetVersion,
				Message:     err.Error(),
			}
		}
		current = wf.SchemaVersion
	}
	return false, &MigrationError{
		FromVersion: current,
		ToVersion:   targetVersion,
		Message:     "migration exceeded max iterations (possible cycle)",
	}
}

// registerDefaults registers built-in schema migrations.
func (m *SchemaMigrator) registerDefaults() {
	m.Register("1.0.0", func(wf *Workflow) error {
		wf.SchemaVersion = "2.0.0"
		return nil
	})
}

// MigrationError represents a workflow schema migration error.
type MigrationError struct {
	FromVersion string `json:"from_version"`
	ToVersion   string `json:"to_version"`
	Message     string `json:"message"`
}

func (e *MigrationError) Error() string {
	return "workflow migration " + e.FromVersion + " -> " + e.ToVersion + ": " + e.Message
}

// WorkflowManager handles file-based workflow CRUD operations.
type WorkflowManager struct {
	migrator *SchemaMigrator
}

// NewWorkflowManager creates a new WorkflowManager.
func NewWorkflowManager() *WorkflowManager {
	return &WorkflowManager{
		migrator: NewSchemaMigrator(),
	}
}

// Read reads and parses a workflow from a JSON file.
// Automatically migrates the workflow to the latest schema version.
func (m *WorkflowManager) Read(path string) (*Workflow, error) {
	wf, err := LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	migrated, err := m.migrator.Migrate(wf, CurrentSchemaVersion)
	if err != nil {
		return nil, err
	}
	if migrated {
		if err := SaveToFile(wf, path); err != nil {
			return nil, err
		}
	}

	return wf, nil
}

// Write saves a workflow to a JSON file atomically.
func (m *WorkflowManager) Write(wf *Workflow, path string) error {
	return SaveToFile(wf, path)
}

// Exists checks if a workflow file exists at the given path.
func (m *WorkflowManager) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Delete removes a workflow file from disk.
func (m *WorkflowManager) Delete(path string) error {
	return os.Remove(path)
}

// CreateDefault creates a new default workflow with the given parameters.
func (m *WorkflowManager) CreateDefault(projectID, name, target string) *Workflow {
	now := time.Now()
	return &Workflow{
		SchemaVersion: CurrentSchemaVersion,
		ID:            projectID,
		Name:          name,
		Version:       1,
		Target:        Target(target),
		Nodes:         make([]Node, 0),
		Edges:         make([]Edge, 0),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// GenerateJSONSchema generates a JSON Schema (Draft-07) that validates workflow.json files.
func (wf *Workflow) GenerateJSONSchema() map[string]any {
	validNodeTypes := make([]string, len(ValidNodeTypes()))
	for i, nt := range ValidNodeTypes() {
		validNodeTypes[i] = string(nt)
	}
	validDataTypes := make([]string, len(ValidDataTypes()))
	for i, dt := range ValidDataTypes() {
		validDataTypes[i] = string(dt)
	}
	validTargets := make([]string, len(ValidTargets()))
	for i, t := range ValidTargets() {
		validTargets[i] = string(t)
	}

	return map[string]any{
		"$schema":              "http://json-schema.org/draft-07/schema#",
		"$id":                  "https://aistudio.ai/schemas/workflow_v2.json",
		"title":                "AIStudio Workflow",
		"description":          "JSON Schema for AIStudio workflow.json files (version 2.0.0)",
		"type":                 "object",
		"required":             []string{"schema_version", "id", "name", "target", "nodes", "edges"},
		"additionalProperties": false,
		"properties": map[string]any{
			"schema_version": map[string]any{
				"type":    "string",
				"enum":    []string{"2.0.0"},
				"default": "2.0.0",
			},
			"id": map[string]any{
				"type":        "string",
				"description": "Unique workflow identifier",
				"minLength":   1,
			},
			"name": map[string]any{
				"type":        "string",
				"description": "Human-readable workflow name",
				"minLength":   1,
			},
			"description": map[string]any{
				"type": "string",
			},
			"version": map[string]any{
				"type":    "integer",
				"minimum": 1,
				"default": 1,
			},
			"author": map[string]any{
				"type": "string",
			},
			"tags": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
			"metadata": map[string]any{
				"type": "object",
			},
			"variables": map[string]any{
				"type": "object",
			},
			"target": map[string]any{
					"type": "string",
					"enum": validTargets,
				},
				"domain": map[string]any{
					"type": "string",
				},
				"viewport": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"x":    map[string]any{"type": "number"},
						"y":    map[string]any{"type": "number"},
						"zoom": map[string]any{"type": "number"},
					},
				},
				"plugins": map[string]any{
					"type": "object",
					"additionalProperties": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id":      map[string]any{"type": "string"},
							"name":    map[string]any{"type": "string"},
							"version": map[string]any{"type": "string"},
							"enabled": map[string]any{"type": "boolean"},
						},
					},
				},
			"nodes": map[string]any{
				"type":     "array",
				"items":    wf.generateNodeSchema(validNodeTypes, validDataTypes),
				"minItems": 1,
			},
			"edges": map[string]any{
				"type":  "array",
				"items": wf.generateEdgeSchema(),
			},
			"created_at": map[string]any{
				"type":   "string",
				"format": "date-time",
			},
			"updated_at": map[string]any{
				"type":   "string",
				"format": "date-time",
			},
		},
	}
}

// generateNodeSchema returns the JSON Schema for a workflow node.
func (wf *Workflow) generateNodeSchema(validNodeTypes, validDataTypes []string) map[string]any {
	portSchema := map[string]any{
		"type":     "object",
		"required": []string{"id", "name", "type"},
		"properties": map[string]any{
			"id":          map[string]any{"type": "string", "minLength": 1},
			"name":        map[string]any{"type": "string", "minLength": 1},
			"type":        map[string]any{"type": "string", "enum": validDataTypes},
			"description": map[string]any{"type": "string"},
			"required":    map[string]any{"type": "boolean", "default": false},
		},
	}

	return map[string]any{
		"type":     "object",
		"required": []string{"id", "type", "name", "position"},
		"properties": map[string]any{
			"id":          map[string]any{"type": "string", "minLength": 1},
			"type":        map[string]any{"type": "string", "enum": validNodeTypes},
			"name":        map[string]any{"type": "string", "minLength": 1},
			"description": map[string]any{"type": "string"},
			"position": map[string]any{
				"type":     "object",
				"required": []string{"x", "y"},
				"properties": map[string]any{
					"x": map[string]any{"type": "number"},
					"y": map[string]any{"type": "number"},
				},
			},
			"size": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"width":  map[string]any{"type": "number"},
					"height": map[string]any{"type": "number"},
				},
			},
			"config": map[string]any{
				"type": "object",
			},
			"inputs": map[string]any{
				"type":  "array",
				"items": portSchema,
			},
			"outputs": map[string]any{
				"type":  "array",
				"items": portSchema,
			},
			"constraints": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"min_inputs":      map[string]any{"type": "integer"},
						"max_inputs":      map[string]any{"type": "integer"},
						"min_outputs":     map[string]any{"type": "integer"},
						"max_outputs":     map[string]any{"type": "integer"},
						"required_config": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
						"allowed_types":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					},
				},
				"status":     map[string]any{"type": "string"},
				"enabled":    map[string]any{"type": "boolean"},
				"plugin":     map[string]any{"type": "string"},
				"domain":     map[string]any{"type": "string"},
				"metadata":   map[string]any{"type": "object"},
				"created_at": map[string]any{"type": "string", "format": "date-time"},
				"updated_at": map[string]any{"type": "string", "format": "date-time"},
			},
		}
	}

// generateEdgeSchema returns the JSON Schema for a workflow edge.
func (wf *Workflow) generateEdgeSchema() map[string]any {
	endpointSchema := map[string]any{
		"type":     "object",
		"required": []string{"node_id", "port_id"},
		"properties": map[string]any{
			"node_id": map[string]any{"type": "string", "minLength": 1},
			"port_id": map[string]any{"type": "string", "minLength": 1},
		},
	}

	conditionSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"expression":  map[string]any{"type": "string"},
			"true_label":  map[string]any{"type": "string"},
			"false_label": map[string]any{"type": "string"},
		},
	}

	return map[string]any{
		"type":     "object",
		"required": []string{"id", "source", "target"},
		"properties": map[string]any{
			"id":        map[string]any{"type": "string", "minLength": 1},
			"source":    endpointSchema,
			"target":    endpointSchema,
			"label":     map[string]any{"type": "string"},
			"type":      map[string]any{"type": "string", "enum": []string{"data", "control", "condition"}},
			"condition": conditionSchema,
			"metadata":  map[string]any{"type": "object"},
		},
	}
}

// SaveSchema saves the JSON Schema to a file.
func SaveSchema(wf *Workflow, path string) error {
	schema := wf.GenerateJSONSchema()
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON schema: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
