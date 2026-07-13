package workflow

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

const (
	BufferSize = 4096
)

var parseBufPool = sync.Pool{
	New: func() any {
		b := make([]byte, BufferSize)
		return &b
	},
}

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

// ParseFile parses a workflow from a JSON file using streaming decoder.
func ParseFile(path string) (*Workflow, error) {
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

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("workflow file open error: %w", err)
	}
	defer f.Close()

	var wf Workflow
	decoder := json.NewDecoder(bufio.NewReaderSize(f, BufferSize))
	if err := decoder.Decode(&wf); err != nil {
		return nil, fmt.Errorf("workflow parse error: %w", err)
	}

	if err := validateSchema(&wf); err != nil {
		return nil, fmt.Errorf("workflow validation error: %w", err)
	}

	return &wf, nil
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

// Save writes a workflow to a JSON file using atomic write with buffered I/O.
func Save(wf *Workflow, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("workflow directory create error: %w", err)
	}

	data, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		return fmt.Errorf("workflow marshal error: %w", err)
	}

	tmpPath := filepath.Join(dir, "."+uuid.New().String()+".tmp")
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("workflow file create error: %w", err)
	}
	buf := bufio.NewWriterSize(tmpFile, BufferSize)
	if _, err := buf.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("workflow file write error: %w", err)
	}
	if err := buf.Flush(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("workflow file flush error: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("workflow file sync error: %w", err)
	}
	tmpFile.Close()

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("workflow file rename error: %w", err)
	}

	return nil
}

// SaveDirect writes a workflow to a JSON file directly (no atomic write) with buffered I/O.
func SaveDirect(wf *Workflow, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("workflow directory create error: %w", err)
	}

	data, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		return fmt.Errorf("workflow marshal error: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("workflow file create error: %w", err)
	}
	defer f.Close()
	buf := bufio.NewWriterSize(f, BufferSize)
	if _, err := buf.Write(data); err != nil {
		return fmt.Errorf("workflow file write error: %w", err)
	}
	if err := buf.Flush(); err != nil {
		return fmt.Errorf("workflow file flush error: %w", err)
	}
	return nil
}

// ToJSON marshals a workflow to indented JSON.
func ToJSON(wf *Workflow) ([]byte, error) {
	return json.MarshalIndent(wf, "", "  ")
}

// Clone creates a deep copy of a workflow using marshal/unmarshal roundtrip.
func Clone(wf *Workflow) (*Workflow, error) {
	data, err := json.Marshal(wf)
	if err != nil {
		return nil, fmt.Errorf("workflow clone marshal error: %w", err)
	}
	var clone Workflow
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil, fmt.Errorf("workflow clone unmarshal error: %w", err)
	}
	return &clone, nil
}

// LoadFromFile reads and parses a workflow from a JSON file on disk.
func LoadFromFile(path string) (*Workflow, error) {
	return ParseFile(path)
}

// SaveToFile writes a workflow to a JSON file on disk with atomic write.
func SaveToFile(wf *Workflow, path string) error {
	return Save(wf, path)
}

// ValidateWorkflowFile reads a workflow file and validates its contents.
func ValidateWorkflowFile(path string) error {
	wf, err := LoadFromFile(path)
	if err != nil {
		return err
	}

	result := ValidateWorkflow(wf)
	if !result.Valid {
		return fmt.Errorf("workflow validation failed: %v", result.Errors)
	}

	return nil
}
