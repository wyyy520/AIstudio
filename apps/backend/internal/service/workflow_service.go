package service

import (
	"fmt"
	"time"

	"github.com/aistudio/backend/internal/database/models"
	"github.com/aistudio/backend/internal/project"
	"github.com/aistudio/backend/internal/workflow"
	"gorm.io/gorm"
)

// WorkflowService handles workflow business logic.
// Workflow data lives in workflow.json on disk (Single Source of Truth).
// Database stores only metadata for indexing/search.
type WorkflowService struct {
	db      *gorm.DB
	pm      *project.Manager
	mgr     *workflow.WorkflowManager
	engine  *workflow.Engine
}

// NewWorkflowService creates a new WorkflowService.
func NewWorkflowService(db *gorm.DB, pm *project.Manager) *WorkflowService {
	engine := workflow.NewEngineWithBuiltIns()
	return &WorkflowService{
		db:     db,
		pm:     pm,
		mgr:    workflow.NewWorkflowManager(),
		engine: engine,
	}
}

// List returns all workflow metadata entries, optionally filtered by project ID.
func (s *WorkflowService) List(projectID string) ([]models.Workflow, error) {
	var workflows []models.Workflow
	query := s.db.Model(&models.Workflow{}).Order("updated_at DESC")
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	if err := query.Find(&workflows).Error; err != nil {
		return nil, fmt.Errorf("list workflows: %w", err)
	}
	return workflows, nil
}

// Get returns a single workflow metadata entry by ID.
func (s *WorkflowService) Get(id string) (*models.Workflow, error) {
	var wf models.Workflow
	if err := s.db.First(&wf, id).Error; err != nil {
		return nil, fmt.Errorf("workflow not found: %s", id)
	}
	return &wf, nil
}

// ReadWorkflowFile reads and parses a workflow.json from the given path.
func (s *WorkflowService) ReadWorkflowFile(path string) (*workflow.Workflow, error) {
	return s.mgr.Read(path)
}

// ReadByProject reads the workflow.json for the given project.
func (s *WorkflowService) ReadByProject(projectID string) (*workflow.Workflow, error) {
	wfPath := s.pm.GetWorkflowPath(projectID)
	if wfPath == "" {
		return nil, fmt.Errorf("project not found: %s", projectID)
	}
	return s.mgr.Read(wfPath)
}

// WriteByProject saves a workflow to the project's workflow.json.
func (s *WorkflowService) WriteByProject(projectID string, wf *workflow.Workflow) error {
	wfPath := s.pm.GetWorkflowPath(projectID)
	if wfPath == "" {
		return fmt.Errorf("project not found: %s", projectID)
	}
	wf.UpdatedAt = time.Now()
	return s.mgr.Write(wf, wfPath)
}

// Create creates a new workflow metadata entry and a default workflow.json.
func (s *WorkflowService) Create(projectID, name string) (*models.Workflow, error) {
	projectDir := s.pm.GetProjectDir(projectID)
	if projectDir == "" {
		return nil, fmt.Errorf("project not found: %s", projectID)
	}

	wfPath := s.pm.GetWorkflowPath(projectID)

	now := time.Now()
	meta := models.Workflow{
		Name:          name,
		SchemaVersion: workflow.CurrentSchemaVersion,
		Version:       1,
		Path:          wfPath,
		Status:        "active",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.db.Create(&meta).Error; err != nil {
		return nil, fmt.Errorf("create workflow metadata: %w", err)
	}

	return &meta, nil
}

// UpdateMetadata updates the workflow metadata (not the workflow.json content).
func (s *WorkflowService) UpdateMetadata(projectID, name string) error {
	var workflow models.Workflow
	if err := s.db.Where("project_id = ?", projectID).First(&workflow).Error; err != nil {
		return fmt.Errorf("workflow metadata not found for project: %s", projectID)
	}

	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if name != "" {
		updates["name"] = name
	}
	return s.db.Model(&workflow).Updates(updates).Error
}

// Delete removes a workflow metadata entry and the workflow.json file.
func (s *WorkflowService) Delete(id string) error {
	var meta models.Workflow
	if err := s.db.First(&meta, id).Error; err != nil {
		return fmt.Errorf("workflow not found: %s", id)
	}

	// Remove workflow.json from disk (best effort)
	if meta.Path != "" {
		s.mgr.Delete(meta.Path)
	}

	if err := s.db.Delete(&meta).Error; err != nil {
		return fmt.Errorf("delete workflow: %w", err)
	}
	return nil
}

// Run executes a workflow by its database ID.
func (s *WorkflowService) Run(id string) (*workflow.ExecutionResult, error) {
	meta, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	wf, err := s.mgr.Read(meta.Path)
	if err != nil {
		return nil, fmt.Errorf("read workflow file: %w", err)
	}

	return s.engine.RunWorkflow(nil, wf)
}

// ListNodeTypes returns all registered workflow node types.
func (s *WorkflowService) ListNodeTypes() []workflow.NodeDefinition {
	return s.engine.Registry().List()
}

// Engine returns the workflow engine instance.
func (s *WorkflowService) Engine() *workflow.Engine {
	return s.engine
}