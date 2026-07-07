package service

import (
	"fmt"

	"github.com/aistudio/backend/internal/database/models"
	"gorm.io/gorm"
)

// UserService handles user business logic.
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new UserService.
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// List returns all users.
func (s *UserService) List() ([]models.User, error) {
	var users []models.User
	if err := s.db.Order("updated_at DESC").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

// Get returns a single user by ID.
func (s *UserService) Get(id string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	return &user, nil
}

// Create creates a new user.
func (s *UserService) Create(username, email, password string) (*models.User, error) {
	user := models.User{
		Username: username,
		Email:    email,
		Password: password, // TODO: hash password in production
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &user, nil
}

// Update updates an existing user.
func (s *UserService) Update(id string, updates map[string]interface{}) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("user not found: %s", id)
	}

	allowed := map[string]bool{"username": true, "email": true, "password": true}
	filtered := make(map[string]interface{})
	for k, v := range updates {
		if allowed[k] {
			filtered[k] = v
		}
	}

	if len(filtered) > 0 {
		if err := s.db.Model(&user).Updates(filtered).Error; err != nil {
			return nil, fmt.Errorf("update user: %w", err)
		}
		s.db.First(&user, user.ID)
	}
	return &user, nil
}

// Delete removes a user by ID.
func (s *UserService) Delete(id string) error {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return fmt.Errorf("user not found: %s", id)
	}
	if err := s.db.Delete(&user).Error; err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}