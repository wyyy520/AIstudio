package cloud

import "time"

type SyncService interface {
	StartSync(projectID string) (*SyncStatus, error)
	GetSyncStatus(projectID string) (*SyncStatus, error)
	StopSync(projectID string) error
}

type syncService struct {
	statuses map[string]*SyncStatus
}

func NewSyncService() SyncService {
	return &syncService{
		statuses: make(map[string]*SyncStatus),
	}
}

func (s *syncService) StartSync(projectID string) (*SyncStatus, error) {
	status := &SyncStatus{
		Syncing:  true,
		LastSync: time.Now(),
	}
	s.statuses[projectID] = status
	return status, nil
}

func (s *syncService) GetSyncStatus(projectID string) (*SyncStatus, error) {
	status, ok := s.statuses[projectID]
	if !ok {
		return &SyncStatus{Idle: true}, nil
	}
	return status, nil
}

func (s *syncService) StopSync(projectID string) error {
	if status, ok := s.statuses[projectID]; ok {
		status.Syncing = false
		status.Idle = true
	}
	return nil
}
