package service

import (
	"strconv"

	"github.com/aistudio/backend/internal/auth"
	"github.com/aistudio/backend/internal/database/models"
)

type UserService struct {
	users *auth.UserManager
}

func NewUserService(users *auth.UserManager) *UserService {
	return &UserService{users: users}
}

func (s *UserService) List() ([]models.User, error) {
	return s.users.List()
}

func (s *UserService) Get(id string) (*models.User, error) {
	return s.getUser(id)
}

func (s *UserService) Create(username, email, password string) (*models.User, error) {
	return s.users.Create(username, email, password)
}

func (s *UserService) Update(id string, updates map[string]interface{}) (*models.User, error) {
	uid, err := parseID(id)
	if err != nil {
		return nil, err
	}
	return s.users.Update(uid, updates)
}

func (s *UserService) Delete(id string) error {
	uid, err := parseID(id)
	if err != nil {
		return err
	}
	return s.users.Delete(uid)
}

func (s *UserService) getUser(id string) (*models.User, error) {
	uid, err := parseID(id)
	if err != nil {
		return nil, err
	}
	return s.users.GetByID(uid)
}

func parseID(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}