package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Memory provides persistent storage for agent preferences, workflows, and history.
// Uses SQLite via GORM for persistence.
type Memory struct {
	db *gorm.DB
}

// NewMemory creates a new Memory with the given database connection.
// Auto-migrates the required tables.
func NewMemory(db *gorm.DB) (*Memory, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required for agent memory")
	}

	if err := db.AutoMigrate(&Preference{}, &WorkflowRecord{}, &ConversationEntry{}); err != nil {
		return nil, fmt.Errorf("failed to migrate agent memory tables: %w", err)
	}

	log.Println("[agent-memory] tables migrated successfully")
	return &Memory{db: db}, nil
}

// ---- Preference ----

// Preference stores user preferences as key-value pairs.
type Preference struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index;not null" json:"user_id"`
	Key       string    `gorm:"not null" json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SavePreference stores or updates a user preference.
func (m *Memory) SavePreference(userID, key, value string) error {
	var pref Preference
	result := m.db.Where("user_id = ? AND key = ?", userID, key).First(&pref)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			pref = Preference{
				UserID: userID,
				Key:    key,
				Value:  value,
			}
			return m.db.Create(&pref).Error
		}
		return result.Error
	}

	pref.Value = value
	return m.db.Save(&pref).Error
}

// GetPreference retrieves a user preference by key.
func (m *Memory) GetPreference(userID, key string) (string, error) {
	var pref Preference
	result := m.db.Where("user_id = ? AND key = ?", userID, key).First(&pref)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", result.Error
	}
	return pref.Value, nil
}

// GetAllPreferences returns all preferences for a user.
func (m *Memory) GetAllPreferences(userID string) (map[string]string, error) {
	var prefs []Preference
	result := m.db.Where("user_id = ?", userID).Find(&prefs)
	if result.Error != nil {
		return nil, result.Error
	}

	out := make(map[string]string)
	for _, p := range prefs {
		out[p.Key] = p.Value
	}
	return out, nil
}

// DeletePreference removes a user preference.
func (m *Memory) DeletePreference(userID, key string) error {
	return m.db.Where("user_id = ? AND key = ?", userID, key).Delete(&Preference{}).Error
}

// ---- Workflow Record ----

// WorkflowRecord stores a workflow that was created via the agent.
type WorkflowRecord struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	ProjectID   string          `gorm:"index;not null" json:"project_id"`
	Name        string          `json:"name"`
	Goal        string          `json:"goal"`
	Plugin      string          `json:"plugin"`
	WorkflowID  string          `json:"workflow_id"`
	Definition  json.RawMessage `gorm:"type:json" json:"definition"`
	Parameters  json.RawMessage `gorm:"type:json" json:"parameters"`
	CreatedAt   time.Time       `json:"created_at"`
}

// SaveWorkflow stores a workflow record.
func (m *Memory) SaveWorkflow(record WorkflowRecord) error {
	return m.db.Create(&record).Error
}

// GetRecentWorkflows returns the most recent workflows.
func (m *Memory) GetRecentWorkflows(limit int) ([]WorkflowRecord, error) {
	var records []WorkflowRecord
	result := m.db.Order("created_at DESC").Limit(limit).Find(&records)
	return records, result.Error
}

// GetWorkflowsByProject returns workflows for a specific project.
func (m *Memory) GetWorkflowsByProject(projectID string) ([]WorkflowRecord, error) {
	var records []WorkflowRecord
	result := m.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&records)
	return records, result.Error
}

// GetWorkflowsByGoal searches workflows by goal text (simple LIKE match).
func (m *Memory) GetWorkflowsByGoal(goal string, limit int) ([]WorkflowRecord, error) {
	var records []WorkflowRecord
	result := m.db.Where("goal LIKE ?", "%"+goal+"%").Order("created_at DESC").Limit(limit).Find(&records)
	return records, result.Error
}

// ---- Conversation Entry ----

// ConversationEntry stores a single agent conversation turn.
type ConversationEntry struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID string    `gorm:"index" json:"project_id"`
	UserID    string    `gorm:"index" json:"user_id"`
	Role      string    `json:"role"` // "user" or "agent"
	Content   string    `json:"content"`
	Goal      string    `json:"goal"`
	CreatedAt time.Time `json:"created_at"`
}

// SaveConversation stores a conversation entry.
func (m *Memory) SaveConversation(entry ConversationEntry) error {
	return m.db.Create(&entry).Error
}

// GetConversationHistory returns recent conversation entries.
func (m *Memory) GetConversationHistory(projectID string, limit int) ([]ConversationEntry, error) {
	var entries []ConversationEntry
	query := m.db.Order("created_at DESC").Limit(limit)
	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	result := query.Find(&entries)
	return entries, result.Error
}

// GetConversationByGoal returns conversations related to a specific goal.
func (m *Memory) GetConversationByGoal(goal string, limit int) ([]ConversationEntry, error) {
	var entries []ConversationEntry
	result := m.db.Where("goal LIKE ?", "%"+goal+"%").Order("created_at DESC").Limit(limit).Find(&entries)
	return entries, result.Error
}

// SearchConversations searches conversations by keyword (simple LIKE-based search)
func (m *Memory) SearchConversations(query string, limit int) ([]ConversationEntry, error) {
	var entries []ConversationEntry
	result := m.db.Where("content LIKE ?", "%"+query+"%").
		Order("created_at DESC").
		Limit(limit).
		Find(&entries)
	return entries, result.Error
}

// NewMemoryInMemory creates an in-memory Memory for testing.
func NewMemoryInMemory() (*Memory, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open in-memory database: %w", err)
	}

	if err := db.AutoMigrate(&Preference{}, &WorkflowRecord{}, &ConversationEntry{}); err != nil {
		return nil, fmt.Errorf("failed to migrate in-memory tables: %w", err)
	}

	return &Memory{db: db}, nil
}