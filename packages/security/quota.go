package security

import (
	"fmt"
)

type QuotaStore interface {
	GetUsage(userID string, resource string) (used int64, total int64, err error)
	IncrementUsage(userID string, resource string, delta int64) error
	DecrementUsage(userID string, resource string, delta int64) error
	SetLimit(userID string, resource string, total int64) error
}

type QuotaManager struct {
	store QuotaStore
}

func NewQuotaManager(store QuotaStore) *QuotaManager {
	return &QuotaManager{store: store}
}

func (m *QuotaManager) CheckQuota(userID string, resource string) error {
	used, total, err := m.store.GetUsage(userID, resource)
	if err != nil {
		return fmt.Errorf("get quota: %w", err)
	}
	if total >= 0 && used >= total {
		return fmt.Errorf("quota exceeded for %s", resource)
	}
	return nil
}

func (m *QuotaManager) IncrementUsage(userID string, resource string, delta int64) error {
	used, total, err := m.store.GetUsage(userID, resource)
	if err != nil {
		return fmt.Errorf("get quota: %w", err)
	}
	if total >= 0 && used+delta > total {
		return fmt.Errorf("quota exceeded for %s", resource)
	}
	return m.store.IncrementUsage(userID, resource, delta)
}

func (m *QuotaManager) DecrementUsage(userID string, resource string, delta int64) error {
	return m.store.DecrementUsage(userID, resource, delta)
}

func (m *QuotaManager) GetUsage(userID string, resource string) (used, total int64, err error) {
	return m.store.GetUsage(userID, resource)
}

func (m *QuotaManager) SetLimit(userID string, resource string, total int64) error {
	return m.store.SetLimit(userID, resource, total)
}

type QuotaResource string

const (
	QuotaConcurrentTasks QuotaResource = "concurrent_tasks"
	QuotaDailyTasks      QuotaResource = "daily_tasks"
	QuotaCPUUsage        QuotaResource = "cpu_usage"
	QuotaGPUUsage        QuotaResource = "gpu_usage"
	QuotaDiskSpace       QuotaResource = "disk_space"
	QuotaAPIRequests     QuotaResource = "api_requests"
)

var DefaultQuotaLimits = map[QuotaResource]int64{
	QuotaConcurrentTasks: 3,
	QuotaDailyTasks:      20,
	QuotaCPUUsage:        80,
	QuotaGPUUsage:        1,
	QuotaDiskSpace:       10 * 1024 * 1024 * 1024,
	QuotaAPIRequests:     1000,
}
