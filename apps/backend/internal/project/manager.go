// Package project manages AIStudio projects on the filesystem.
//
// All projects are real filesystem directories.
// All files are real files.
// All code is real code.
// Projects can be executed independently of AIStudio.
//
// The database only stores metadata (recent projects, indexing).
// The actual project content lives on the filesystem.
//
// Workflow.json is the Single Source of Truth for each project.
// It is always stored at <project-root>/.aistudio/workflow.json
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aistudio/backend/internal/workflow"
	"github.com/google/uuid"
)

// ============================================================================
// Types
// ============================================================================

// Project represents a project on the filesystem.
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	RootPath    string    `json:"rootPath"`     // Absolute path to project root
	WorkflowID  string    `json:"workflowId"`   // Linked workflow ID
	Target      string    `json:"target"`       // python, matlab, ros2, etc.
	Status      string    `json:"status"`       // active, archived, deleted
	FileCount   int       `json:"fileCount"`
	SizeBytes   int64     `json:"sizeBytes"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ProjectInfo contains summary information about a project.
type ProjectInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	RootPath    string    `json:"rootPath"`
	Target      string    `json:"target"`
	Status      string    `json:"status"`
	FileCount   int       `json:"fileCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateOptions specifies options for creating a project.
type CreateOptions struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Target      string `json:"target"`
	Template    string `json:"template,omitempty"` // Template to use
	WorkflowID  string `json:"workflowId,omitempty"`
}

// ============================================================================
// Manager
// ============================================================================

// Manager manages project lifecycle on the filesystem.
type Manager struct {
	mu             sync.RWMutex
	projectsDir    string
	index          map[string]*Project
}

// NewManager creates a new Project Manager.
func NewManager(projectsDir string) *Manager {
	// Ensure projects directory exists
	os.MkdirAll(projectsDir, 0755)

	return &Manager{
		projectsDir: projectsDir,
		index:       make(map[string]*Project),
	}
}

// Create creates a new project directory on the filesystem.
func (m *Manager) Create(opts CreateOptions) (*Project, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	id := uuid.New().String()
	safeName := sanitizeName(opts.Name)
	rootPath := filepath.Join(m.projectsDir, safeName)

	// Ensure unique directory name
	if _, err := os.Stat(rootPath); err == nil {
		rootPath = filepath.Join(m.projectsDir, fmt.Sprintf("%s-%s", safeName, id[:8]))
	}

	// Create project directory structure
	dirs := []string{
		rootPath,
		filepath.Join(rootPath, ".aistudio"),
		filepath.Join(rootPath, "src"),
		filepath.Join(rootPath, "data"),
		filepath.Join(rootPath, "models"),
		filepath.Join(rootPath, "outputs"),
		filepath.Join(rootPath, "tests"),
		filepath.Join(rootPath, "config"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create project metadata file
	now := time.Now()
	project := &Project{
		ID:          id,
		Name:        opts.Name,
		Description: opts.Description,
		RootPath:    rootPath,
		WorkflowID:  opts.WorkflowID,
		Target:      opts.Target,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Save project metadata
	if err := m.saveProjectJSON(project); err != nil {
		return nil, fmt.Errorf("failed to save project metadata: %w", err)
	}

	// Auto-create default workflow.json (Single Source of Truth)
	wfPath := m.GetWorkflowPath(id)
	defaultWf := workflow.Workflow{
		SchemaVersion: workflow.CurrentSchemaVersion,
		ID:            id,
		Name:          opts.Name,
		Version:       1,
		Target:        workflow.Target(opts.Target),
		Nodes:         make([]workflow.Node, 0),
		Edges:         make([]workflow.Edge, 0),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := workflow.Save(&defaultWf, wfPath); err != nil {
		return nil, fmt.Errorf("failed to create workflow.json: %w", err)
	}

	// Update index
	m.mu.Lock()
	m.index[id] = project
	m.mu.Unlock()

	return project, nil
}

// Open opens an existing project from the filesystem.
func (m *Manager) Open(path string) (*Project, error) {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("project path does not exist: %s", path)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("project path is not a directory: %s", path)
	}

	// Try to load project metadata
	project, err := m.loadProjectJSON(path)
	if err != nil {
		// Create metadata for existing directory
		project = &Project{
			ID:        uuid.New().String(),
			Name:      filepath.Base(path),
			RootPath:  path,
			Status:    "active",
			CreatedAt: info.ModTime(),
			UpdatedAt: time.Now(),
		}
		m.saveProjectJSON(project)
	}

	// Update file count and size
	m.updateProjectStats(project)

	// Ensure workflow.json exists
	if err := m.ensureWorkflowJSON(project); err != nil {
		return nil, fmt.Errorf("failed to ensure workflow.json: %w", err)
	}

	// Update index
	m.mu.Lock()
	m.index[project.ID] = project
	m.mu.Unlock()

	return project, nil
}

// Get returns a project by ID.
func (m *Manager) Get(id string) (*Project, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.index[id]
	if !ok {
		return nil, false
	}

	// Refresh stats
	m.updateProjectStats(p)
	return p, true
}

// List returns all indexed projects.
func (m *Manager) List() []*ProjectInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	projects := make([]*ProjectInfo, 0, len(m.index))
	for _, p := range m.index {
		if p.Status != "deleted" {
			projects = append(projects, &ProjectInfo{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				RootPath:    p.RootPath,
				Target:      p.Target,
				Status:      p.Status,
				FileCount:   p.FileCount,
				CreatedAt:   p.CreatedAt,
				UpdatedAt:   p.UpdatedAt,
			})
		}
	}

	// Sort by updated time, newest first
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].UpdatedAt.After(projects[j].UpdatedAt)
	})

	return projects
}

// Delete removes a project from the filesystem.
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	project, ok := m.index[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("project not found: %s", id)
	}

	// Mark as deleted (don't actually delete files for safety)
	project.Status = "deleted"
	project.UpdatedAt = time.Now()
	m.mu.Unlock()

	// Update metadata file
	return m.saveProjectJSON(project)
}

// DeletePermanently permanently removes a project from the filesystem.
func (m *Manager) DeletePermanently(id string) error {
	m.mu.Lock()
	project, ok := m.index[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("project not found: %s", id)
	}
	delete(m.index, id)
	m.mu.Unlock()

	return os.RemoveAll(project.RootPath)
}

// GetWorkflowPath returns the path to the workflow.json file for a project.
// Workflow.json is stored at the project root as the Single Source of Truth.
func (m *Manager) GetWorkflowPath(projectID string) string {
	m.mu.RLock()
	project, ok := m.index[projectID]
	m.mu.RUnlock()

	if !ok {
		return ""
	}
	return filepath.Join(project.RootPath, "workflow.json")
}

// ensureWorkflowJSON ensures a workflow.json exists for the project.
func (m *Manager) ensureWorkflowJSON(project *Project) error {
	wfPath := filepath.Join(project.RootPath, "workflow.json")
	if _, err := os.Stat(wfPath); err == nil {
		return nil
	}

	now := time.Now()
	defaultWf := workflow.Workflow{
		SchemaVersion: workflow.CurrentSchemaVersion,
		ID:            project.ID,
		Name:          project.Name,
		Version:       1,
		Target:        workflow.Target(project.Target),
		Nodes:         make([]workflow.Node, 0),
		Edges:         make([]workflow.Edge, 0),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	return workflow.Save(&defaultWf, wfPath)
}

// GetProjectDir returns the root directory for a project.
func (m *Manager) GetProjectDir(projectID string) string {
	m.mu.RLock()
	project, ok := m.index[projectID]
	m.mu.RUnlock()

	if !ok {
		return ""
	}
	return project.RootPath
}

// Scan scans the projects directory and indexes all projects.
func (m *Manager) Scan() error {
	entries, err := os.ReadDir(m.projectsDir)
	if err != nil {
		return fmt.Errorf("failed to scan projects directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			path := filepath.Join(m.projectsDir, entry.Name())
			if _, err := m.Open(path); err != nil {
				// Skip directories that can't be opened as projects
				continue
			}
		}
	}

	return nil
}

// ============================================================================
// Private
// ============================================================================

func (m *Manager) saveProjectJSON(project *Project) error {
	path := filepath.Join(project.RootPath, ".aistudio", "project.json")
	data := fmt.Sprintf(`{
  "id": "%s",
  "name": "%s",
  "description": "%s",
  "target": "%s",
  "status": "%s",
  "workflowId": "%s",
  "createdAt": "%s",
  "updatedAt": "%s"
}`,
		project.ID,
		escapeJSON(project.Name),
		escapeJSON(project.Description),
		project.Target,
		project.Status,
		project.WorkflowID,
		project.CreatedAt.Format(time.RFC3339),
		project.UpdatedAt.Format(time.RFC3339),
	)
	return os.WriteFile(path, []byte(data), 0644)
}

func (m *Manager) loadProjectJSON(path string) (*Project, error) {
	projPath := filepath.Join(path, ".aistudio", "project.json")
	data, err := os.ReadFile(projPath)
	if err != nil {
		return nil, err
	}

	// Simple JSON parsing (avoid import cycle)
	project := &Project{
		RootPath: path,
		Status:   "active",
	}

	content := string(data)
	project.ID = extractJSONField(content, "id")
	project.Name = extractJSONField(content, "name")
	project.Description = extractJSONField(content, "description")
	project.Target = extractJSONField(content, "target")
	project.WorkflowID = extractJSONField(content, "workflowId")

	if status := extractJSONField(content, "status"); status != "" {
		project.Status = status
	}

	return project, nil
}

func (m *Manager) updateProjectStats(project *Project) {
	fileCount := 0
	var totalSize int64

	filepath.Walk(project.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			fileCount++
			totalSize += info.Size()
		}
		return nil
	})

	project.FileCount = fileCount
	project.SizeBytes = totalSize
}

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
	return string(result)
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

func extractJSONField(json, field string) string {
	prefix := fmt.Sprintf(`"%s": "`, field)
	idx := strings.Index(json, prefix)
	if idx < 0 {
		return ""
	}
	start := idx + len(prefix)
	end := strings.Index(json[start:], "\"")
	if end < 0 {
		return ""
	}
	return json[start : start+end]
}