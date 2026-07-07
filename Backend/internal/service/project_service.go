package service

import (
	"fmt"

	"github.com/aistudio/backend/internal/database/models"
	"gorm.io/gorm"
)

// ProjectService handles project business logic.
type ProjectService struct {
	db *gorm.DB
}

// NewProjectService creates a new ProjectService.
func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
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

// Create creates a new project.
func (s *ProjectService) Create(name, description string, ownerID uint) (*models.Project, error) {
	project := models.Project{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		Status:      "active",
	}
	if err := s.db.Create(&project).Error; err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return &project, nil
}

// Update updates an existing project.
func (s *ProjectService) Update(id string, updates map[string]interface{}) (*models.Project, error) {
	var project models.Project
	if err := s.db.First(&project, id).Error; err != nil {
		return nil, fmt.Errorf("project not found: %s", id)
	}

	// Only allow specific fields
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

	// Reload to get updated values
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