package auth

import (
	"fmt"
	"time"

	"github.com/aistudio/backend/internal/database/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserManager struct {
	db *gorm.DB
}

func NewUserManager(db *gorm.DB) *UserManager {
	return &UserManager{db: db}
}

func (m *UserManager) Create(username, email, password string, opts ...UserOption) (*models.User, error) {
	var existing models.User
	if err := m.db.Where("username = ? OR email = ?", username, email).First(&existing).Error; err == nil {
		return nil, ErrDuplicateUser
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         "user",
		Status:       "active",
	}
	for _, opt := range opts {
		opt(user)
	}

	if err := m.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

type UserOption func(*models.User)

func WithNickname(nickname string) UserOption {
	return func(u *models.User) { u.Nickname = nickname }
}

func WithRole(role string) UserOption {
	return func(u *models.User) { u.Role = role }
}

func WithAvatar(avatar string) UserOption {
	return func(u *models.User) { u.Avatar = avatar }
}

func (m *UserManager) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := m.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &user, nil
}

func (m *UserManager) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := m.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return &user, nil
}

func (m *UserManager) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := m.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

func (m *UserManager) List() ([]models.User, error) {
	var users []models.User
	if err := m.db.Order("updated_at DESC").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (m *UserManager) Update(id uint, updates map[string]interface{}) (*models.User, error) {
	allowed := map[string]bool{
		"nickname": true,
		"email":    true,
		"avatar":   true,
		"role":     true,
		"status":   true,
	}
	filtered := make(map[string]interface{})
	for k, v := range updates {
		if allowed[k] {
			filtered[k] = v
		}
	}

	if len(filtered) == 0 {
		return m.GetByID(id)
	}

	if err := m.db.Model(&models.User{}).Where("id = ?", id).Updates(filtered).Error; err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return m.GetByID(id)
}

func (m *UserManager) Delete(id uint) error {
	result := m.db.Delete(&models.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (m *UserManager) ValidatePassword(user *models.User, password string) error {
	if user.Status != "active" {
		return ErrUserDisabled
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidCredential
	}
	return nil
}

func (m *UserManager) ChangePassword(id uint, oldPassword, newPassword string) error {
	user, err := m.GetByID(id)
	if err != nil {
		return err
	}
	if err := m.ValidatePassword(user, oldPassword); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := m.db.Model(&models.User{}).Where("id = ?", id).
		Update("password_hash", string(hash)).Error; err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

func (m *UserManager) UpdateLastLogin(id uint) {
	now := time.Now()
	m.db.Model(&models.User{}).Where("id = ?", id).
		Update("last_login_at", &now)
}
