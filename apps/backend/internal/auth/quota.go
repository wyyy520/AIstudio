package auth

import (
	"fmt"
	"time"

	"github.com/aistudio/backend/internal/database/models"
	"gorm.io/gorm"
)

type QuotaManager struct {
	db *gorm.DB
}

func NewQuotaManager(db *gorm.DB) *QuotaManager {
	return &QuotaManager{db: db}
}

type QuotaResource string

const (
	QuotaConcurrentTasks  QuotaResource = "concurrent_tasks"
	QuotaDailyTasks       QuotaResource = "daily_tasks"
	QuotaCPUUsage         QuotaResource = "cpu_usage"
	QuotaGPUUsage         QuotaResource = "gpu_usage"
	QuotaDiskSpace        QuotaResource = "disk_space"
	QuotaAPIRequests      QuotaResource = "api_requests"
)

var defaultQuotas = map[QuotaResource]int64{
	QuotaConcurrentTasks: 3,
	QuotaDailyTasks:      20,
	QuotaCPUUsage:        80,
	QuotaGPUUsage:        1,
	QuotaDiskSpace:       10 * 1024 * 1024 * 1024,
	QuotaAPIRequests:     1000,
}

func (m *QuotaManager) InitDefaults(userID uint) error {
	tx := m.db.Begin()

	for resource, limit := range defaultQuotas {
		now := time.Now()
		periodEnd := now.Add(24 * time.Hour).Truncate(24 * time.Hour)

		quota := &models.Quota{
			UserID:       userID,
			ResourceType: string(resource),
			Limit:        limit,
			Used:         0,
			PeriodStart:  now,
			PeriodEnd:    periodEnd,
		}
		if err := tx.Create(quota).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("init quota %s: %w", resource, err)
		}
	}

	return tx.Commit().Error
}

func (m *QuotaManager) GetUserQuotas(userID uint) ([]models.Quota, error) {
	var quotas []models.Quota
	if err := m.db.Where("user_id = ?", userID).Find(&quotas).Error; err != nil {
		return nil, fmt.Errorf("get user quotas: %w", err)
	}
	return quotas, nil
}

func (m *QuotaManager) CheckQuota(userID uint, resource QuotaResource) error {
	quota, err := m.getOrCreate(userID, resource)
	if err != nil {
		return err
	}
	if quota.Limit >= 0 && quota.Used >= quota.Limit {
		return ErrQuotaExceeded
	}
	return nil
}

func (m *QuotaManager) IncrementUsage(userID uint, resource QuotaResource, delta int64) error {
	quota, err := m.getOrCreate(userID, resource)
	if err != nil {
		return err
	}

	if quota.Limit >= 0 && quota.Used+delta > quota.Limit {
		return ErrQuotaExceeded
	}

	return m.db.Model(&models.Quota{}).Where("id = ?", quota.ID).
		Update("used", gorm.Expr("used + ?", delta)).Error
}

func (m *QuotaManager) DecrementUsage(userID uint, resource QuotaResource, delta int64) error {
	return m.db.Model(&models.Quota{}).
		Where("user_id = ? AND resource_type = ?", userID, string(resource)).
		Update("used", gorm.Expr("GREATEST(used - ?, 0)", delta)).Error
}

func (m *QuotaManager) ResetIfExpired(userID uint) error {
	now := time.Now()
	var quotas []models.Quota
	if err := m.db.Where("user_id = ? AND period_end < ?", userID, now).
		Find(&quotas).Error; err != nil {
		return err
	}

	for _, q := range quotas {
		nextEnd := q.PeriodEnd.Add(24 * time.Hour)
		m.db.Model(&q).Updates(map[string]interface{}{
			"used":         0,
			"period_start": q.PeriodEnd,
			"period_end":   nextEnd,
		})
	}
	return nil
}

func (m *QuotaManager) UpdateLimit(userID uint, resource QuotaResource, limit int64) error {
	result := m.db.Model(&models.Quota{}).
		Where("user_id = ? AND resource_type = ?", userID, string(resource)).
		Update("limit", limit)
	if result.Error != nil {
		return fmt.Errorf("update quota limit: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		now := time.Now()
		quota := &models.Quota{
			UserID:       userID,
			ResourceType: string(resource),
			Limit:        limit,
			Used:         0,
			PeriodStart:  now,
			PeriodEnd:    now.Add(24 * time.Hour).Truncate(24 * time.Hour),
		}
		return m.db.Create(quota).Error
	}
	return nil
}

func (m *QuotaManager) getOrCreate(userID uint, resource QuotaResource) (*models.Quota, error) {
	var quota models.Quota
	err := m.db.Where("user_id = ? AND resource_type = ?", userID, string(resource)).
		First(&quota).Error
	if err == nil {
		if quota.PeriodEnd.Before(time.Now()) {
			nextEnd := quota.PeriodEnd.Add(24 * time.Hour)
			m.db.Model(&quota).Updates(map[string]interface{}{
				"used":         0,
				"period_start": quota.PeriodEnd,
				"period_end":   nextEnd,
			})
			quota.Used = 0
			quota.PeriodStart = quota.PeriodEnd
			quota.PeriodEnd = nextEnd
		}
		return &quota, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("get quota: %w", err)
	}

	limit, exists := defaultQuotas[resource]
	if !exists {
		limit = -1
	}

	now := time.Now()
	quota = models.Quota{
		UserID:       userID,
		ResourceType: string(resource),
		Limit:        limit,
		Used:         0,
		PeriodStart:  now,
		PeriodEnd:    now.Add(24 * time.Hour).Truncate(24 * time.Hour),
	}
	if err := m.db.Create(&quota).Error; err != nil {
		return nil, fmt.Errorf("create quota: %w", err)
	}
	return &quota, nil
}
