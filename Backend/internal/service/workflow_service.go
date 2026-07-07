package service

import (
	"context"
	"fmt"

	"github.com/aistudio/backend/internal/database/models"
	"github.com/aistudio/backend/internal/workflow"
	"gorm.io/gorm"
)

// WorkflowService handles workflow business logic.
type WorkflowService struct {
	db     *gorm.DB
	engine *workflow.Engine
}

// NewWorkflowService creates a new WorkflowService.
func NewWorkflowService(db *gorm.DB, engine *workflow.Engine) *WorkflowService {
	return &WorkflowService{db: db, engine: engine}
}

// List returns all workflows, optionally filtered by project ID.
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

// Get returns a single workflow by ID.
func (s *WorkflowService) Get(id string) (*models.Workflow, error) {
	var wf models.Workflow
	if err := s.db.First(&wf, id).Error; err != nil {
		return nil, fmt.Errorf("workflow not found: %s", id)
	}
	return &wf, nil
}

// Create creates a new workflow.
func (s *WorkflowService) Create(projectID uint, name, definition string) (*models.Workflow, error) {
	wf := models.Workflow{
		ProjectID:  projectID,
		Name:       name,
		Definition: definition,
		Status:     "draft",
	}
	if err := s.db.Create(&wf).Error; err != nil {
		return nil, fmt.Errorf("create workflow: %w", err)
	}
	return &wf, nil
}

// Update updates an existing workflow.
func (s *WorkflowService) Update(id string, updates map[string]interface{}) (*models.Workflow, error) {
	var wf models.Workflow
	if err := s.db.First(&wf, id).Error; err != nil {
		return nil, fmt.Errorf("workflow not found: %s", id)
	}

	allowed := map[string]bool{"name": true, "definition": true, "status": true}
	filtered := make(map[string]interface{})
	for k, v := range updates {
		if allowed[k] {
			filtered[k] = v
		}
	}

	if len(filtered) > 0 {
		if err := s.db.Model(&wf).Updates(filtered).Error; err != nil {
			return nil, fmt.Errorf("update workflow: %w", err)
		}
		s.db.First(&wf, wf.ID)
	}
	return &wf, nil
}

// Delete removes a workflow by ID.
func (s *WorkflowService) Delete(id string) error {
	var wf models.Workflow
	if err := s.db.First(&wf, id).Error; err != nil {
		return fmt.Errorf("workflow not found: %s", id)
	}
	if err := s.db.Delete(&wf).Error; err != nil {
		return fmt.Errorf("delete workflow: %w", err)
	}
	return nil
}

// Run executes a workflow by its database ID.
// Parses the stored Definition JSON and runs it through the engine.
func (s *WorkflowService) Run(ctx context.Context, id string) (*workflow.ExecutionResult, error) {
	wf, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	if wf.Definition == "" {
		return nil, fmt.Errorf("workflow %s has no definition", id)
	}

	result, err := s.engine.Run(ctx, []byte(wf.Definition))
	if err != nil {
		return nil, fmt.Errorf("run workflow: %w", err)
	}

	// Update status
	s.db.Model(wf).Update("status", "completed")
	return result, nil
}

// ListNodeTypes returns all registered workflow node types.
func (s *WorkflowService) ListNodeTypes() []workflow.NodeDefinition {
	return s.engine.Registry().List()
}