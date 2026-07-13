package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// ============================================================================
// Workflow Parser
// ============================================================================

// Parse parses a workflow from JSON bytes.
func Parse(data []byte) (*Workflow, error) {
	var wf Workflow
	if err := json.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("workflow parse error: %w", err)
	}

	if err := validateSchema(&wf); err != nil {
		return nil, fmt.Errorf("workflow validation error: %w", err)
	}

	return &wf, nil
}

// ParseFile parses a workflow from a JSON file.
// Validates that the file exists, is valid JSON, and matches the schema.
func ParseFile(path string) (*Workflow, error) {
	// Validate file exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("workflow file not found: %s: %w", path, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("workflow path is a directory, not a file: %s", path)
	}
	if info.Size() == 0 {
		return nil, fmt.Errorf("workflow file is empty: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("workflow file read error: %w", err)
	}

	if !json.Valid(data) {
		return nil, fmt.Errorf("workflow file contains invalid JSON: %s", path)
	}

	return Parse(data)
}

// MustParse parses a workflow from JSON bytes, panicking on error.
func MustParse(data []byte) *Workflow {
	wf, err := Parse(data)
	if err != nil {
		panic(err)
	}
	return wf
}

// MustParseFile parses a workflow from a JSON file, panicking on error.
func MustParseFile(path string) *Workflow {
	wf, err := ParseFile(path)
	if err != nil {
		panic(err)
	}
	return wf
}

// Save saves a workflow to a JSON file using atomic write.
// Writes to a temporary file first, then renames to prevent corruption.
func Save(wf *Workflow, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("workflow directory create error: %w", err)
	}

	data, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		return fmt.Errorf("workflow marshal error: %w", err)
	}

	// Atomic write: write to temp file, then rename
	tmpPath := filepath.Join(dir, "."+uuid.New().String()+".tmp")
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("workflow file write error: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("workflow file rename error: %w", err)
	}

	return nil
}

// SaveDirect writes a workflow to a JSON file directly (no atomic write).
// Use Save() instead for atomicity; this is for internal use.
func SaveDirect(wf *Workflow, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("workflow directory create error: %w", err)
	}

	data, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		return fmt.Errorf("workflow marshal error: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("workflow file write error: %w", err)
	}

	return nil
}

// ToJSON marshals a workflow to indented JSON.
func ToJSON(wf *Workflow) ([]byte, error) {
	return json.MarshalIndent(wf, "", "  ")
}

// Clone creates a deep copy of a workflow.
func Clone(wf *Workflow) (*Workflow, error) {
	data, err := json.Marshal(wf)
	if err != nil {
		return nil, fmt.Errorf("workflow clone marshal error: %w", err)
	}
	return Parse(data)
}

// validateSchema validates the workflow schema and required fields.
func validateSchema(wf *Workflow) error {
	if wf.ID == "" {
		return fmt.Errorf("workflow ID is required")
	}
	if wf.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	if wf.SchemaVersion == "" {
		wf.SchemaVersion = CurrentSchemaVersion
	}
	if wf.Target == "" {
		return fmt.Errorf("workflow target is required")
	}

	// Validate target
	valid := false
	for _, t := range ValidTargets() {
		if wf.Target == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid target: %s", wf.Target)
	}

	// Validate nodes — allow empty for new workflows
	if len(wf.Nodes) == 0 {
		return nil
	}

	nodeIDs := make(map[string]bool)
	for _, node := range wf.Nodes {
		if node.ID == "" {
			return fmt.Errorf("node ID is required")
		}
		if nodeIDs[node.ID] {
			return fmt.Errorf("duplicate node ID: %s", node.ID)
		}
		nodeIDs[node.ID] = true
	}

	// Validate edges
	for _, edge := range wf.Edges {
		if edge.ID == "" {
			return fmt.Errorf("edge ID is required")
		}
		if !nodeIDs[edge.Source.NodeID] {
			return fmt.Errorf("edge %s: source node %s not found", edge.ID, edge.Source.NodeID)
		}
		if !nodeIDs[edge.Target.NodeID] {
			return fmt.Errorf("edge %s: target node %s not found", edge.ID, edge.Target.NodeID)
		}
	}

	return nil
}