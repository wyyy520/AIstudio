package project

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aistudio/packages/workflow"
	"github.com/google/uuid"
)

type Manager struct {
	mu          sync.RWMutex
	projectsDir string
	index       map[string]*Project
}

func NewManager(projectsDir string) *Manager {
	os.MkdirAll(projectsDir, 0755)
	return &Manager{
		projectsDir: projectsDir,
		index:       make(map[string]*Project),
	}
}

func (m *Manager) Create(name, target, dir string) (*Project, error) {
	if name == "" {
		return nil, fmt.Errorf("project name is required")
	}
	if target == "" {
		return nil, fmt.Errorf("project target is required")
	}
	if dir == "" {
		dir = m.projectsDir
	}

	id := uuid.New().String()
	safeName := sanitizeName(name)
	rootPath := filepath.Join(dir, safeName)

	if _, err := os.Stat(rootPath); err == nil {
		rootPath = filepath.Join(dir, fmt.Sprintf("%s-%s", safeName, id[:8]))
	}

	dirs := []string{
		rootPath,
		filepath.Join(rootPath, "logs"),
		filepath.Join(rootPath, "models"),
		filepath.Join(rootPath, "cache"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	wfPath := filepath.Join(rootPath, "workflow.json")
	wfMgr := workflow.NewWorkflowManager()
	wf := wfMgr.CreateDefault(id, name, target)
	if err := workflow.Save(wf, wfPath); err != nil {
		return nil, fmt.Errorf("failed to create workflow.json: %w", err)
	}

	if err := os.WriteFile(filepath.Join(rootPath, ".gitignore"), []byte(DefaultGitIgnore()), 0644); err != nil {
		return nil, fmt.Errorf("failed to create .gitignore: %w", err)
	}

	if err := os.WriteFile(filepath.Join(rootPath, "README.md"), []byte(DefaultReadme(name, "")), 0644); err != nil {
		return nil, fmt.Errorf("failed to create README.md: %w", err)
	}

	if err := os.WriteFile(filepath.Join(rootPath, "settings.json"), []byte(DefaultSettings()), 0644); err != nil {
		return nil, fmt.Errorf("failed to create settings.json: %w", err)
	}

	now := time.Now()
	project := &Project{
		ID:         id,
		Name:       name,
		RootPath:   rootPath,
		WorkflowID: id,
		Target:     target,
		Status:     ProjectStatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := m.saveMetadata(project); err != nil {
		return nil, fmt.Errorf("failed to save project metadata: %w", err)
	}

	m.mu.Lock()
	m.index[id] = project
	m.mu.Unlock()

	return project, nil
}

func (m *Manager) Open(path string) (*Project, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("project path does not exist: %s", path)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("project path is not a directory: %s", path)
	}

	project, err := m.loadMetadata(path)
	if err != nil {
		project = &Project{
			ID:        uuid.New().String(),
			Name:      filepath.Base(path),
			RootPath:  path,
			Status:    ProjectStatusActive,
			CreatedAt: info.ModTime(),
			UpdatedAt: time.Now(),
		}
		m.saveMetadata(project)
	}

	m.updateStats(project)

	wfPath := filepath.Join(path, "workflow.json")
	if _, err := os.Stat(wfPath); os.IsNotExist(err) {
		wfMgr := workflow.NewWorkflowManager()
		wf := wfMgr.CreateDefault(project.ID, project.Name, project.Target)
		if project.Target == "" {
			wf.Target = workflow.TargetPython
		}
		if err := workflow.Save(wf, wfPath); err != nil {
			return nil, fmt.Errorf("failed to create workflow.json: %w", err)
		}
	}

	m.mu.Lock()
	m.index[project.ID] = project
	m.mu.Unlock()

	return project, nil
}

func (m *Manager) Close(project *Project) error {
	project.UpdatedAt = time.Now()
	if err := m.saveMetadata(project); err != nil {
		return fmt.Errorf("failed to save project metadata on close: %w", err)
	}
	m.mu.Lock()
	delete(m.index, project.ID)
	m.mu.Unlock()
	return nil
}

func (m *Manager) Delete(project *Project) error {
	if err := os.RemoveAll(project.RootPath); err != nil {
		return fmt.Errorf("failed to remove project directory: %w", err)
	}
	m.mu.Lock()
	delete(m.index, project.ID)
	m.mu.Unlock()
	return nil
}

func (m *Manager) Import(path string) (*Project, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("import path does not exist: %s", path)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("import path is not a directory: %s", path)
	}

	wfPath := filepath.Join(path, "workflow.json")
	var target string
	if _, err := os.Stat(wfPath); err == nil {
		wf, parseErr := workflow.ParseFile(wfPath)
		if parseErr == nil {
			target = string(wf.Target)
		}
	} else {
		wfMgr := workflow.NewWorkflowManager()
		id := uuid.New().String()
		wf := wfMgr.CreateDefault(id, filepath.Base(path), string(workflow.TargetPython))
		if err := workflow.Save(wf, wfPath); err != nil {
			return nil, fmt.Errorf("failed to generate workflow.json: %w", err)
		}
		target = string(workflow.TargetPython)
	}

	subDirs := []string{"logs", "models", "cache"}
	for _, d := range subDirs {
		os.MkdirAll(filepath.Join(path, d), 0755)
	}

	now := time.Now()
	project := &Project{
		ID:        uuid.New().String(),
		Name:      filepath.Base(path),
		RootPath:  path,
		Target:    target,
		Status:    ProjectStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	m.updateStats(project)

	if err := m.saveMetadata(project); err != nil {
		return nil, fmt.Errorf("failed to save project metadata: %w", err)
	}

	m.mu.Lock()
	m.index[project.ID] = project
	m.mu.Unlock()

	return project, nil
}

func (m *Manager) Export(project *Project, format string) (string, error) {
	switch strings.ToLower(format) {
	case "zip":
		return m.exportZip(project)
	case "tar", "tar.gz", "tgz":
		return m.exportTar(project)
	default:
		return "", fmt.Errorf("unsupported export format: %s (supported: zip, tar)", format)
	}
}

func (m *Manager) List() ([]*Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	projects := make([]*Project, 0, len(m.index))
	for _, p := range m.index {
		if p.Status != ProjectStatusDeleted {
			projects = append(projects, p)
		}
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].UpdatedAt.After(projects[j].UpdatedAt)
	})

	return projects, nil
}

func (m *Manager) Get(id string) (*Project, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.index[id]
	if ok {
		m.updateStats(p)
	}
	return p, ok
}

func (m *Manager) GetByPath(path string) *Project {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.index {
		if p.RootPath == path {
			return p
		}
	}
	return nil
}

func (m *Manager) Scan() error {
	entries, err := os.ReadDir(m.projectsDir)
	if err != nil {
		return fmt.Errorf("failed to scan projects directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			path := filepath.Join(m.projectsDir, entry.Name())
			if _, openErr := m.Open(path); openErr != nil {
				continue
			}
		}
	}
	return nil
}

func (m *Manager) ProjectCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.index)
}

func (m *Manager) ProjectsDir() string {
	return m.projectsDir
}

func (m *Manager) saveMetadata(project *Project) error {
	metaPath := filepath.Join(project.RootPath, ".aistudio", "project.json")
	os.MkdirAll(filepath.Dir(metaPath), 0755)

	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal project metadata: %w", err)
	}
	return os.WriteFile(metaPath, data, 0644)
}

func (m *Manager) loadMetadata(path string) (*Project, error) {
	metaPath := filepath.Join(path, ".aistudio", "project.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}
	var project Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, err
	}
	project.RootPath = path
	return &project, nil
}

func (m *Manager) updateStats(project *Project) {
	fileCount := 0
	var totalSize int64
	filepath.Walk(project.RootPath, func(p string, info os.FileInfo, err error) error {
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

func (m *Manager) exportZip(project *Project) (string, error) {
	exportPath := project.RootPath + ".zip"
	f, err := os.Create(exportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create zip file: %w", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	baseName := filepath.Base(project.RootPath)
	err = filepath.Walk(project.RootPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(project.RootPath, p)
		if relErr != nil {
			return nil
		}
		if rel == "." {
			return nil
		}
		zipPath := filepath.Join(baseName, rel)
		if info.IsDir() {
			zipPath += "/"
			_, headerErr := w.Create(zipPath)
			return headerErr
		}
		header, headerErr := zip.FileInfoHeader(info)
		if headerErr != nil {
			return headerErr
		}
		header.Name = zipPath
		header.Method = zip.Deflate
		writer, writerErr := w.CreateHeader(header)
		if writerErr != nil {
			return writerErr
		}
		src, openErr := os.Open(p)
		if openErr != nil {
			return openErr
		}
		defer src.Close()
		_, copyErr := io.Copy(writer, src)
		return copyErr
	})
	if err != nil {
		return "", fmt.Errorf("failed to add files to zip: %w", err)
	}

	return exportPath, nil
}

func (m *Manager) exportTar(project *Project) (string, error) {
	exportPath := project.RootPath + ".tar.gz"
	f, err := os.Create(exportPath)
	if err != nil {
		return "", fmt.Errorf("failed to create tar file: %w", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	baseName := filepath.Base(project.RootPath)
	err = filepath.Walk(project.RootPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(project.RootPath, p)
		if relErr != nil {
			return nil
		}
		if rel == "." {
			return nil
		}
		tarPath := filepath.Join(baseName, rel)
		header, headerErr := tar.FileInfoHeader(info, "")
		if headerErr != nil {
			return headerErr
		}
		header.Name = tarPath
		if info.IsDir() {
			header.Name += "/"
		}
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if !info.IsDir() {
			src, openErr := os.Open(p)
			if openErr != nil {
				return openErr
			}
			defer src.Close()
			_, copyErr := io.Copy(tw, src)
			return copyErr
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to add files to tar: %w", err)
	}

	return exportPath, nil
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