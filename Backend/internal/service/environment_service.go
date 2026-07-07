package service

import "github.com/aistudio/backend/internal/environment"

type EnvironmentService struct {
	manager *environment.Manager
}

func NewEnvironmentService(mgr *environment.Manager) *EnvironmentService {
	return &EnvironmentService{manager: mgr}
}

func (s *EnvironmentService) GetStatus() environment.EnvironmentStatus {
	return s.manager.GetStatus()
}

func (s *EnvironmentService) InstallDependency(name string) error {
	return s.manager.InstallDependency(name)
}
