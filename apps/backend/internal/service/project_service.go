package service

import (
	"fmt"
	"time"

	"github.com/aistudio/backend/internal/database/models"
	"github.com/aistudio/backend/internal/project"
	"github.com/aistudio/backend/internal/workflow"
	"gorm.io/gorm"
)

// ProjectService handles project business logic.
type ProjectService struct {
	db *gorm.DB
	pm *project.Manager
}

// NewProjectService creates a new ProjectService.
func NewProjectService(db *gorm.DB, pm *project.Manager) *ProjectService {
	return &ProjectService{db: db, pm: pm}
}

// List returns all projects.
func (s *ProjectService) List() ([]models.Project, error) {
	var projects []models.Project
	if err := s.db.Order("updated_at DESC").Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	return projects, nil
}

// Get returns a single project by ID.
func (s *ProjectService) Get(id string) (*models.Project, error) {
	var project models.Project
	if err := s.db.First(&project, id).Error; err != nil {
		return nil, fmt.Errorf("project not found: %s", id)
	}
	return &project, nil
}

// Create creates a new project on the filesystem and in the database.
func (s *ProjectService) Create(name, description string, ownerID uint) (*models.Project, error) {
	// Create on filesystem first (includes workflow.json)
	fsProject, err := s.pm.Create(project.CreateOptions{
		Name:        name,
		Description: description,
		Target:      "python",
	})
	if err != nil {
		return nil, fmt.Errorf("create project on filesystem: %w", err)
	}

	// Create in database
	now := time.Now()
	project := models.Project{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.db.Create(&project).Error; err != nil {
		// Rollback filesystem creation
		s.pm.DeletePermanently(fsProject.ID)
		return nil, fmt.Errorf("create project in database: %w", err)
	}

	// Create workflow metadata in database
	wfPath := s.pm.GetWorkflowPath(fsProject.ID)
	wfMeta := models.Workflow{
		ProjectID:     project.ID,
		Name:          name,
		SchemaVersion: workflow.CurrentSchemaVersion,
		Version:       1,
		Path:          wfPath,
		Status:        "active",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.db.Create(&wfMeta).Error; err != nil {
		return nil, fmt.Errorf("create workflow metadata: %w", err)
	}

	// Reload to get populated ID
	s.db.First(&project, project.ID)
	return &project, nil
}

// Update updates an existing project.
func (s *ProjectService) Update(id string, updates map[string]interface{}) (*models.Project, error) {
	var project models.Project
	if err := s.db.First(&project, id).Error; err != nil {
		return nil, fmt.Errorf("project not found: %s", id)
	}

	allowed := map[string]bool{"name": true, "description": true, "status": true}
	filtered := make(map[string]interface{})
	for k, v := range updates {
		if allowed[k] {
			filtered[k] = v
		}
	}

	if len(filtered) == 0 {
		return &project, nil
	}

	if err := s.db.Model(&project).Updates(filtered).Error; err != nil {
		return nil, fmt.Errorf("update project: %w", err)
	}

	s.db.First(&project, project.ID)
	return &project, nil
}

// Delete removes a project by ID.
func (s *ProjectService) Delete(id string) error {
	var project models.Project
	if err := s.db.First(&project, id).Error; err != nil {
		return fmt.Errorf("project not found: %s", id)
	}
	if err := s.db.Delete(&project).Error; err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
}
