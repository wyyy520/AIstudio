// Package project manages AIStudio projects on the filesystem.
//
// All projects are real filesystem directories conforming to the AIStudio
// project layout defined in the system design document (§7.4):
//
//	Project/
//	├── workflow.json    // Workflow IR — single source of truth
//	├── project.json     // Project metadata
//	├── settings.json    // Project-level settings
//	├── datasets/        // Dataset files
//	├── generated/       // Auto-generated engineering projects
//	├── outputs/         // Run outputs (logs, artifacts, weights)
//	├── logs/            // Runtime / compile logs
//	├── cache/           // Compiler / template cache
//	├── templates/       // Template cache
//	├── plugins/         // Per-project plugins
//	├── assets/          // Images, docs, resources
//	├── scripts/         // User-defined helper scripts
//	└── temp/            // Temporary files
package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Project Types
// ============================================================================

// Project represents a project on the filesystem.
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	RootPath    string    `json:"rootPath"`              // Absolute path to project root
	Target      string    `json:"target,omitempty"`      // python, matlab, stm32, ansys, etc.
	Status      string    `json:"status"`                // active, archived
	Version     int       `json:"version"`               // Project version (incremented on save)
	FileCount   int       `json:"fileCount"`             // Recursive file count (approx)
	SizeBytes   int64     `json:"sizeBytes"`              // Total directory size (approx)
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	LastOpened  time.Time `json:"lastOpened,omitempty"`  // Last time project was opened
}

// ProjectSummary is the API-facing lightweight project representation.
type ProjectSummary struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Target      string    `json:"target,omitempty"`
	Status      string    `json:"status"`
	RootPath    string    `json:"rootPath"`
	Version     int       `json:"version"`
	FileCount   int       `json:"fileCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// workflowInfo is the minimal metadata we track inside workflow.json.
type workflowInfo struct {
	SchemaVersion string    `json:"schema_version"`
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Version       int       `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateOptions specifies options for creating a new project.
type CreateOptions struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Target      string `json:"target,omitempty"`
	RootDir     string `json:"rootDir,omitempty"` // Parent directory; empty = default projects dir
}

// ============================================================================
// Manager — filesystem-based project lifecycle manager
// ============================================================================

// Manager manages project lifecycle on the real filesystem.
type Manager struct {
	mu          sync.RWMutex
	projectsDir string          // Default parent directory for new projects
	index       map[string]*Project // id → project (in-memory index)
	recentFile  string          // Path to recent-projects JSON file
	recentMax   int
}

// NewManager creates a new Project Manager.
// projectsDir is the default parent directory where new projects are created.
func NewManager(projectsDir string) *Manager {
	absDir, _ := filepath.Abs(projectsDir)
	os.MkdirAll(absDir, 0755)

	m := &Manager{
		projectsDir: absDir,
		index:       make(map[string]*Project),
		recentFile:  filepath.Join(absDir, ".recent.json"),
		recentMax:   20,
	}

	// Load recent projects on startup
	m.loadRecent()

	return m
}

// ============================================================================
// Create
// ============================================================================

// Create creates a new project directory on the real filesystem.
func (m *Manager) Create(opts CreateOptions) (*Project, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	id := uuid.New().String()

	// Determine root path
	parentDir := m.projectsDir
	if opts.RootDir != "" {
		parentDir = opts.RootDir
	}
	safeName := sanitizeName(opts.Name)
	rootPath := filepath.Join(parentDir, safeName)

	// Ensure unique directory name
	if _, err := os.Stat(rootPath); err == nil {
		rootPath = filepath.Join(parentDir, fmt.Sprintf("%s_%s", safeName, id[:8]))
	}

	// Create project directory structure per silu.md §7.4
	dirs := []string{
		rootPath,
		filepath.Join(rootPath, "datasets"),
		filepath.Join(rootPath, "generated"),
		filepath.Join(rootPath, "outputs"),
		filepath.Join(rootPath, "logs"),
		filepath.Join(rootPath, "cache"),
		filepath.Join(rootPath, "templates"),
		filepath.Join(rootPath, "plugins"),
		filepath.Join(rootPath, "scripts"),
		filepath.Join(rootPath, "temp"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create dir %s: %w", dir, err)
		}
	}

	now := time.Now()
	project := &Project{
		ID:          id,
		Name:        opts.Name,
		Description: opts.Description,
		RootPath:    rootPath,
		Target:      opts.Target,
		Status:      "active",
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastOpened:  now,
	}

	// Write project.json at root
	if err := m.writeProjectJSON(project); err != nil {
		return nil, fmt.Errorf("write project.json: %w", err)
	}

	// Write workflow.json at root — single source of truth
	if err := m.writeWorkflowJSON(project); err != nil {
		return nil, fmt.Errorf("write workflow.json: %w", err)
	}

	// Update index
	m.mu.Lock()
	m.index[id] = project
	m.saveRecentLocked()
	m.mu.Unlock()

	return project, nil
}

// ============================================================================
// Open (real folder)
// ============================================================================

// Open opens any real filesystem directory as an AIStudio project.
// If the directory already contains project.json it is restored;
// otherwise metadata is auto-created.
func (m *Manager) Open(path string) (*Project, error) {
	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	// Security: prevent path traversal outside allowed directories
	if !m.isPathAllowed(absPath) {
		return nil, fmt.Errorf("access denied: path outside allowed directories")
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %s", absPath)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", absPath)
	}

	// Check for existing project metadata
	project, err := m.readProjectJSON(absPath)
	if err != nil {
		// Create fresh metadata for an existing directory
		now := time.Now()
		project = &Project{
			ID:         uuid.New().String(),
			Name:       filepath.Base(absPath),
			RootPath:   absPath,
			Status:     "active",
			Version:    1,
			CreatedAt:  now,
			UpdatedAt:  now,
			LastOpened: now,
		}
		if err := m.writeProjectJSON(project); err != nil {
			return nil, fmt.Errorf("create project.json: %w", err)
		}
	}

	project.LastOpened = time.Now()
	project.Status = "active"

	// Ensure required subdirectories exist
	m.ensureDirs(project)

	// Ensure workflow.json exists
	m.ensureWorkflowJSON(project)

	// Update stats
	m.updateProjectStats(project)

	// Save updated metadata
	m.writeProjectJSON(project)

	// Update index
	m.mu.Lock()
	m.index[project.ID] = project
	m.saveRecentLocked()
	m.mu.Unlock()

	return project, nil
}

// OpenByID re-opens a known project by its ID.
func (m *Manager) OpenByID(id string) (*Project, error) {
	m.mu.RLock()
	p, ok := m.index[id]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("project not found: %s", id)
	}

	return m.Open(p.RootPath)
}

// ============================================================================
// Read
// ============================================================================

// Get returns a project by ID from the in-memory index.
func (m *Manager) Get(id string) (*Project, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.index[id]
	if !ok {
		return nil, false
	}
	return p, true
}

// List returns all indexed projects as summaries, sorted by last-opened (newest first).
func (m *Manager) List() []*ProjectSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*ProjectSummary, 0, len(m.index))
	for _, p := range m.index {
		if p.Status == "deleted" {
			continue
		}
		list = append(list, &ProjectSummary{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Target:      p.Target,
			Status:      p.Status,
			RootPath:    p.RootPath,
			Version:     p.Version,
			FileCount:   p.FileCount,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].UpdatedAt.After(list[j].UpdatedAt)
	})

	return list
}

// Recent returns the most recently opened projects (max n).
func (m *Manager) Recent(n int) []*ProjectSummary {
	if n <= 0 {
		n = 10
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect, sort by LastOpened desc, limit
	all := make([]*Project, 0, len(m.index))
	for _, p := range m.index {
		if p.Status == "deleted" {
			continue
		}
		all = append(all, p)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].LastOpened.After(all[j].LastOpened)
	})

	if len(all) > n {
		all = all[:n]
	}

	result := make([]*ProjectSummary, len(all))
	for i, p := range all {
		result[i] = &ProjectSummary{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Target:      p.Target,
			Status:      p.Status,
			RootPath:    p.RootPath,
			Version:     p.Version,
			FileCount:   p.FileCount,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
	}
	return result
}

// ============================================================================
// Update / Save
// ============================================================================

// Update updates a project's mutable fields and persists to disk.
func (m *Manager) Update(id string, updates map[string]interface{}) (*Project, error) {
	m.mu.Lock()
	project, ok := m.index[id]
	if !ok {
		m.mu.Unlock()
		return nil, fmt.Errorf("project not found: %s", id)
	}

	if v, ok := updates["name"]; ok {
		if name, ok := v.(string); ok && name != "" {
			project.Name = name
		}
	}
	if v, ok := updates["description"]; ok {
		if desc, ok := v.(string); ok {
			project.Description = desc
		}
	}
	if v, ok := updates["target"]; ok {
		if t, ok := v.(string); ok {
			project.Target = t
		}
	}
	if v, ok := updates["status"]; ok {
		if s, ok := v.(string); ok {
			project.Status = s
		}
	}

	project.UpdatedAt = time.Now()
	project.Version++

	// Persist
	if err := m.writeProjectJSON(project); err != nil {
		m.mu.Unlock()
		return nil, fmt.Errorf("save project.json: %w", err)
	}

	m.saveRecentLocked()
	m.mu.Unlock()

	cp := *project
	return &cp, nil
}

// SaveWorkflow persists the workflow.json for a project.
// content must be valid JSON-serializable workflow data.
func (m *Manager) SaveWorkflow(projectID string, content interface{}) error {
	m.mu.RLock()
	project, ok := m.index[projectID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("project not found: %s", projectID)
	}

	data, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal workflow: %w", err)
	}

	wfPath := filepath.Join(project.RootPath, "workflow.json")
	if err := os.WriteFile(wfPath, data, 0644); err != nil {
		return fmt.Errorf("write workflow.json: %w", err)
	}

	m.mu.Lock()
	project.UpdatedAt = time.Now()
	project.Version++
	m.mu.Unlock()

	return nil
}

// ReadWorkflow reads and unmarshals the workflow.json for a project.
func (m *Manager) ReadWorkflow(projectID string, target interface{}) error {
	m.mu.RLock()
	project, ok := m.index[projectID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("project not found: %s", projectID)
	}

	wfPath := filepath.Join(project.RootPath, "workflow.json")
	data, err := os.ReadFile(wfPath)
	if err != nil {
		return fmt.Errorf("read workflow.json: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("unmarshal workflow: %w", err)
	}

	return nil
}

// ============================================================================
// Delete
// ============================================================================

// Delete marks a project as deleted (soft delete — safe).
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	project, ok := m.index[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("project not found: %s", id)
	}

	project.Status = "deleted"
	project.UpdatedAt = time.Now()
	m.mu.Unlock()

	return m.writeProjectJSON(project)
}

// DeletePermanently removes a project from disk entirely.
func (m *Manager) DeletePermanently(id string) error {
	m.mu.Lock()
	project, ok := m.index[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("project not found: %s", id)
	}
	delete(m.index, id)
	m.saveRecentLocked()
	m.mu.Unlock()

	return os.RemoveAll(project.RootPath)
}

// ============================================================================
// Scan — re-index all projects in the default projects directory
// ============================================================================

// Scan scans the default projects directory and re-indexes all projects.
func (m *Manager) Scan() error {
	entries, err := os.ReadDir(m.projectsDir)
	if err != nil {
		return fmt.Errorf("scan projects dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(m.projectsDir, entry.Name())
		// Skip if already indexed
		m.mu.RLock()
		alreadyIndexed := false
		for _, p := range m.index {
			if p.RootPath == path {
				alreadyIndexed = true
				break
			}
		}
		m.mu.RUnlock()
		if alreadyIndexed {
			continue
		}
		// Try to open as project
		if _, err := m.Open(path); err != nil {
			continue
		}
	}

	return nil
}

// ============================================================================
// Utility methods
// ============================================================================

// GetProjectDir returns the root path for a project by ID.
func (m *Manager) GetProjectDir(projectID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.index[projectID]
	if !ok {
		return ""
	}
	return p.RootPath
}

// GetWorkflowPath returns the filesystem path to a project's workflow.json.
func (m *Manager) GetWorkflowPath(projectID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.index[projectID]
	if !ok {
		return ""
	}
	return filepath.Join(p.RootPath, "workflow.json")
}

// DefaultProjectsDir returns the configured default projects directory.
func (m *Manager) DefaultProjectsDir() string {
	return m.projectsDir
}

// ============================================================================
// Internal — project.json I/O
// ============================================================================

func (m *Manager) writeProjectJSON(p *Project) error {
	path := filepath.Join(p.RootPath, "project.json")
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (m *Manager) readProjectJSON(rootPath string) (*Project, error) {
	path := filepath.Join(rootPath, "project.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var p Project
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	p.RootPath = rootPath
	return &p, nil
}

// ============================================================================
// Internal — workflow.json I/O
// ============================================================================

func (m *Manager) writeWorkflowJSON(p *Project) error {
	now := time.Now()
	wf := map[string]interface{}{
		"schema_version": "2.0.0",
		"id":             p.ID,
		"name":           p.Name,
		"description":    p.Description,
		"version":        1,
		"target":         p.Target,
		"domain":         p.Target,
		"nodes":          []interface{}{},
		"edges":          []interface{}{},
		"viewport": map[string]interface{}{
			"x":    0,
			"y":    0,
			"zoom": 1,
		},
		"created_at": now.Format(time.RFC3339),
		"updated_at": now.Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		return err
	}

	wfPath := filepath.Join(p.RootPath, "workflow.json")
	return os.WriteFile(wfPath, data, 0644)
}

func (m *Manager) ensureWorkflowJSON(p *Project) {
	wfPath := filepath.Join(p.RootPath, "workflow.json")
	if _, err := os.Stat(wfPath); err == nil {
		return
	}
	m.writeWorkflowJSON(p)
}

// ============================================================================
// Internal — recent projects
// ============================================================================

type recentEntry struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	RootPath  string    `json:"rootPath"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (m *Manager) loadRecent() {
	data, err := os.ReadFile(m.recentFile)
	if err != nil {
		return
	}
	var entries []recentEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return
	}
	for _, e := range entries {
		// Try to re-open each recent project
		if _, err := os.Stat(e.RootPath); err != nil {
			continue
		}
		if _, err := m.Open(e.RootPath); err != nil {
			continue
		}
	}
}

func (m *Manager) saveRecentLocked() {
	entries := make([]recentEntry, 0, len(m.index))
	for _, p := range m.index {
		if p.Status == "deleted" {
			continue
		}
		entries = append(entries, recentEntry{
			ID:        p.ID,
			Name:      p.Name,
			RootPath:  p.RootPath,
			UpdatedAt: p.UpdatedAt,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].UpdatedAt.After(entries[j].UpdatedAt)
	})

	if len(entries) > m.recentMax {
		entries = entries[:m.recentMax]
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(m.recentFile, data, 0644)
}

// ============================================================================
// Internal — helpers
// ============================================================================

func (m *Manager) ensureDirs(p *Project) {
	dirs := []string{
		filepath.Join(p.RootPath, "datasets"),
		filepath.Join(p.RootPath, "generated"),
		filepath.Join(p.RootPath, "outputs"),
		filepath.Join(p.RootPath, "logs"),
		filepath.Join(p.RootPath, "cache"),
		filepath.Join(p.RootPath, "temp"),
	}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}
}

func (m *Manager) updateProjectStats(p *Project) {
	fileCount := 0
	var totalSize int64

	filepath.Walk(p.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		fileCount++
		totalSize += info.Size()
		return nil
	})

	p.FileCount = fileCount
	p.SizeBytes = totalSize
}


// isPathAllowed checks if a path is within allowed directories.
func (m *Manager) isPathAllowed(absPath string) bool {
	if strings.HasPrefix(absPath, m.projectsDir) {
		return true
	}
	return false
}

func sanitizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	var result []rune
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '.' {
			result = append(result, r)
		}
	}
	return string(result)
}
