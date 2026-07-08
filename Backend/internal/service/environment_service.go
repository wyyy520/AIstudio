package service

import (
	"log"

	"github.com/aistudio/backend/internal/environment"
)

type EnvironmentService struct {
	manager *environment.Manager
}

func NewEnvironmentService(mgr *environment.Manager) *EnvironmentService {
	return &EnvironmentService{manager: mgr}
}

func (s *EnvironmentService) GetStatus() environment.EnvironmentStatus {
	status := s.manager.GetStatus()
	if envStatus, ok := status.(environment.EnvironmentStatus); ok {
		return envStatus
	}
	// Fallback: return empty status
	log.Println("[env-service] GetStatus returned unexpected type")
	return environment.EnvironmentStatus{}
}

func (s *EnvironmentService) Check() interface{} {
	return s.manager.Check()
}

func (s *EnvironmentService) GetRepairPlan() interface{} {
	return nil
}

func (s *EnvironmentService) Repair() interface{} {
	return nil
}

func (s *EnvironmentService) InstallDependency(name string) error {
	return s.manager.InstallDependency(name)
}

func (s *EnvironmentService) GetLogs() []environment.LogEntry {
	return s.manager.GetLogs()
}

func (s *EnvironmentService) ClearLogs() {
	s.manager.ClearLogs()
}