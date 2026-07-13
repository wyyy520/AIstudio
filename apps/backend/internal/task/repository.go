package task

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// TaskRecord is the GORM model for persisting tasks to the database.
type TaskRecord struct {
	ID          uint       `gorm:"primaryKey" json:"-"`
	TaskID      string     `gorm:"uniqueIndex;size:128;not null" json:"id"`
	ProjectID   string     `gorm:"index;size:128" json:"projectId"`
	WorkflowID  string     `gorm:"index;size:128" json:"workflowId"`
	Type        string     `gorm:"size:32" json:"type"`
	Name        string     `gorm:"size:256" json:"name"`
	Status      string     `gorm:"size:32;not null;default:waiting" json:"status"`
	Progress    float64    `gorm:"default:0" json:"progress"`
	Priority    int        `gorm:"default:1" json:"priority"`
	Handler     string     `gorm:"size:128" json:"handler"`
	Result      string     `gorm:"type:text" json:"result,omitempty"`
	Error       string     `gorm:"type:text" json:"error,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	StartTime   *time.Time `json:"startedAt,omitempty"`
	EndTime     *time.Time `json:"completedAt,omitempty"`
}

func (TaskRecord) TableName() string {
	return "tasks"
}

// TaskRepository provides database persistence for tasks.
type TaskRepository struct {
	db *gorm.DB
}

// NewTaskRepository creates a new task repository.
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// AutoMigrate ensures the tasks table schema is up to date.
func (r *TaskRepository) AutoMigrate() error {
	if err := r.db.AutoMigrate(&TaskRecord{}); err != nil {
		return err
	}
	log.Println("[task-repo] tasks table migration completed")
	return nil
}

// Save persists a task to the database.
func (r *TaskRepository) Save(task *Task) error {
	record := r.taskToRecord(task)
	result := r.db.Create(record)
	if result.Error != nil {
		return result.Error
	}
	log.Printf("[task-repo] saved task: %s", task.ID)
	return nil
}

// Update persists the current task state to the database.
func (r *TaskRepository) Update(task *Task) error {
	record := r.taskToRecord(task)
	result := r.db.Model(&TaskRecord{}).Where("task_id = ?", task.ID).Updates(map[string]interface{}{
		"status":     record.Status,
		"progress":   record.Progress,
		"result":     record.Result,
		"error":      record.Error,
		"start_time": record.StartTime,
		"end_time":   record.EndTime,
		"updated_at": time.Now(),
	})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByID retrieves a task record by task_id.
func (r *TaskRepository) FindByID(taskID string) (*TaskRecord, error) {
	var record TaskRecord
	result := r.db.Where("task_id = ?", taskID).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

// FindAll retrieves all task records.
func (r *TaskRepository) FindAll() ([]TaskRecord, error) {
	var records []TaskRecord
	result := r.db.Order("created_at DESC").Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	return records, nil
}

// FindByStatus retrieves task records filtered by status.
func (r *TaskRepository) FindByStatus(status string) ([]TaskRecord, error) {
	var records []TaskRecord
	result := r.db.Where("status = ?", status).Order("created_at DESC").Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	return records, nil
}

// Delete removes a task record by task_id.
func (r *TaskRepository) Delete(taskID string) error {
	result := r.db.Where("task_id = ?", taskID).Delete(&TaskRecord{})
	if result.Error != nil {
		return result.Error
	}
	log.Printf("[task-repo] deleted task: %s", taskID)
	return nil
}

// taskToRecord converts a Task domain object to a TaskRecord for persistence.
func (r *TaskRepository) taskToRecord(task *Task) *TaskRecord {
	record := &TaskRecord{
		TaskID:     task.ID,
		ProjectID:  task.ProjectID,
		WorkflowID: task.WorkflowID,
		Type:       string(task.Type),
		Name:       task.Name,
		Status:     string(task.Status),
		Progress:   task.Progress,
		Priority:   int(task.Priority),
		Handler:    task.Handler,
		Error:      task.Error,
		CreatedAt:  task.CreatedAt,
		UpdatedAt:  task.UpdatedAt,
		StartTime:  task.StartTime,
		EndTime:    task.EndTime,
	}

	if task.Result != nil {
		if resultStr, ok := task.Result.(string); ok {
			record.Result = resultStr
		}
	}

	return record
}