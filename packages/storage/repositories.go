package storage

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetByID(id uint) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByEmail(email string) (*User, error)
	List() ([]User, error)
	Create(user *User) error
	Update(id uint, updates map[string]any) error
	Delete(id uint) error
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) GetByID(id uint) (*User, error) {
	var user User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetByUsername(username string) (*User, error) {
	var user User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetByEmail(email string) (*User, error) {
	var user User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) List() ([]User, error) {
	var users []User
	if err := r.db.Order("updated_at DESC").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (r *userRepo) Create(user *User) error {
	return r.db.Create(user).Error
}

func (r *userRepo) Update(id uint, updates map[string]any) error {
	return r.db.Model(&User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *userRepo) Delete(id uint) error {
	return r.db.Delete(&User{}, id).Error
}

type SessionRepository interface {
	Create(session *Session) error
	GetByToken(token string) (*Session, error)
	GetByUserID(userID uint) ([]Session, error)
	UpdateLastAccess(sessionID uint) error
	DeleteByToken(token string) error
	DeleteByUserID(userID uint) error
	CleanupExpired() error
}

type sessionRepo struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepo{db: db}
}

func (r *sessionRepo) Create(session *Session) error {
	return r.db.Create(session).Error
}

func (r *sessionRepo) GetByToken(token string) (*Session, error) {
	var session Session
	if err := r.db.Where("token = ?", token).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepo) GetByUserID(userID uint) ([]Session, error) {
	var sessions []Session
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *sessionRepo) UpdateLastAccess(sessionID uint) error {
	return r.db.Model(&Session{}).Where("id = ?", sessionID).
		Update("last_access_at", time.Now()).Error
}

func (r *sessionRepo) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&Session{}).Error
}

func (r *sessionRepo) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&Session{}).Error
}

func (r *sessionRepo) CleanupExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&Session{}).Error
}

type APIKeyRepository interface {
	Create(key *APIKey) error
	GetByID(id uint) (*APIKey, error)
	GetByUser(userID uint) ([]APIKey, error)
	GetByUserAndProvider(userID uint, provider string) (*APIKey, error)
	UpdateLastUsed(id uint) error
	UpdateStatus(id uint, userID uint, isActive bool) error
	Delete(id uint, userID uint) error
}

type apiKeyRepo struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) APIKeyRepository {
	return &apiKeyRepo{db: db}
}

func (r *apiKeyRepo) Create(key *APIKey) error {
	return r.db.Create(key).Error
}

func (r *apiKeyRepo) GetByID(id uint) (*APIKey, error) {
	var key APIKey
	if err := r.db.First(&key, id).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *apiKeyRepo) GetByUser(userID uint) ([]APIKey, error) {
	var keys []APIKey
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *apiKeyRepo) GetByUserAndProvider(userID uint, provider string) (*APIKey, error) {
	var key APIKey
	if err := r.db.Where("user_id = ? AND provider = ? AND is_active = ?",
		userID, provider, true).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *apiKeyRepo) UpdateLastUsed(id uint) error {
	return r.db.Model(&APIKey{}).Where("id = ?", id).
		Update("last_used_at", time.Now()).Error
}

func (r *apiKeyRepo) UpdateStatus(id uint, userID uint, isActive bool) error {
	return r.db.Model(&APIKey{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_active", isActive).Error
}

func (r *apiKeyRepo) Delete(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&APIKey{}).Error
}

type PermissionRepository interface {
	Create(perm *Permission) error
	GetByUser(userID uint) ([]Permission, error)
	Delete(id uint) error
}

type permissionRepo struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepo{db: db}
}

func (r *permissionRepo) Create(perm *Permission) error {
	return r.db.Create(perm).Error
}

func (r *permissionRepo) GetByUser(userID uint) ([]Permission, error) {
	var perms []Permission
	if err := r.db.Where("user_id = ?", userID).Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *permissionRepo) Delete(id uint) error {
	return r.db.Delete(&Permission{}, id).Error
}

type QuotaRepository interface {
	GetByUserAndResource(userID uint, resource string) (*Quota, error)
	GetByUser(userID uint) ([]Quota, error)
	Create(quota *Quota) error
	Update(quota *Quota) error
	IncrementUsed(id uint, delta int64) error
	UpdateLimit(userID uint, resource string, total int64) error
}

type quotaRepo struct {
	db *gorm.DB
}

func NewQuotaRepository(db *gorm.DB) QuotaRepository {
	return &quotaRepo{db: db}
}

func (r *quotaRepo) GetByUserAndResource(userID uint, resource string) (*Quota, error) {
	var quota Quota
	if err := r.db.Where("user_id = ? AND resource = ?", userID, resource).First(&quota).Error; err != nil {
		return nil, err
	}
	return &quota, nil
}

func (r *quotaRepo) GetByUser(userID uint) ([]Quota, error) {
	var quotas []Quota
	if err := r.db.Where("user_id = ?", userID).Find(&quotas).Error; err != nil {
		return nil, err
	}
	return quotas, nil
}

func (r *quotaRepo) Create(quota *Quota) error {
	return r.db.Create(quota).Error
}

func (r *quotaRepo) Update(quota *Quota) error {
	return r.db.Save(quota).Error
}

func (r *quotaRepo) IncrementUsed(id uint, delta int64) error {
	return r.db.Model(&Quota{}).Where("id = ?", id).
		Update("used", gorm.Expr("used + ?", delta)).Error
}

func (r *quotaRepo) UpdateLimit(userID uint, resource string, total int64) error {
	return r.db.Model(&Quota{}).
		Where("user_id = ? AND resource = ?", userID, resource).
		Update("total", total).Error
}

type ProjectRepository interface {
	Create(project *Project) error
	GetByID(id uint) (*Project, error)
	GetByUser(userID uint) ([]Project, error)
	Update(id uint, updates map[string]any) error
	Delete(id uint) error
}

type projectRepo struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepo{db: db}
}

func (r *projectRepo) Create(project *Project) error {
	return r.db.Create(project).Error
}

func (r *projectRepo) GetByID(id uint) (*Project, error) {
	var project Project
	if err := r.db.First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepo) GetByUser(userID uint) ([]Project, error) {
	var projects []Project
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepo) Update(id uint, updates map[string]any) error {
	return r.db.Model(&Project{}).Where("id = ?", id).Updates(updates).Error
}

func (r *projectRepo) Delete(id uint) error {
	return r.db.Delete(&Project{}, id).Error
}

type PluginRepository interface {
	Create(plugin *Plugin) error
	GetByID(id uint) (*Plugin, error)
	List() ([]Plugin, error)
	Update(id uint, updates map[string]any) error
	Delete(id uint) error
}

type pluginRepo struct {
	db *gorm.DB
}

func NewPluginRepository(db *gorm.DB) PluginRepository {
	return &pluginRepo{db: db}
}

func (r *pluginRepo) Create(plugin *Plugin) error {
	return r.db.Create(plugin).Error
}

func (r *pluginRepo) GetByID(id uint) (*Plugin, error) {
	var plugin Plugin
	if err := r.db.First(&plugin, id).Error; err != nil {
		return nil, err
	}
	return &plugin, nil
}

func (r *pluginRepo) List() ([]Plugin, error) {
	var plugins []Plugin
	if err := r.db.Order("created_at DESC").Find(&plugins).Error; err != nil {
		return nil, err
	}
	return plugins, nil
}

func (r *pluginRepo) Update(id uint, updates map[string]any) error {
	return r.db.Model(&Plugin{}).Where("id = ?", id).Updates(updates).Error
}

func (r *pluginRepo) Delete(id uint) error {
	return r.db.Delete(&Plugin{}, id).Error
}

type TaskRepository interface {
	Create(task *Task) error
	GetByID(id uint) (*Task, error)
	GetByUser(userID uint) ([]Task, error)
	Update(id uint, updates map[string]any) error
	Delete(id uint) error
}

type taskRepo struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepo{db: db}
}

func (r *taskRepo) Create(task *Task) error {
	return r.db.Create(task).Error
}

func (r *taskRepo) GetByID(id uint) (*Task, error) {
	var task Task
	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepo) GetByUser(userID uint) ([]Task, error) {
	var tasks []Task
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepo) Update(id uint, updates map[string]any) error {
	return r.db.Model(&Task{}).Where("id = ?", id).Updates(updates).Error
}

func (r *taskRepo) Delete(id uint) error {
	return r.db.Delete(&Task{}, id).Error
}

type TaskLogRepository interface {
	Create(log *TaskLog) error
	GetByTaskID(taskID string) ([]TaskLog, error)
	DeleteByTaskID(taskID string) error
}

type taskLogRepo struct {
	db *gorm.DB
}

func NewTaskLogRepository(db *gorm.DB) TaskLogRepository {
	return &taskLogRepo{db: db}
}

func (r *taskLogRepo) Create(log *TaskLog) error {
	return r.db.Create(log).Error
}

func (r *taskLogRepo) GetByTaskID(taskID string) ([]TaskLog, error) {
	var logs []TaskLog
	if err := r.db.Where("task_id = ?", taskID).Order("created_at ASC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *taskLogRepo) DeleteByTaskID(taskID string) error {
	return r.db.Where("task_id = ?", taskID).Delete(&TaskLog{}).Error
}

type Repositories struct {
	Users     UserRepository
	Sessions  SessionRepository
	APIKeys   APIKeyRepository
	Perms     PermissionRepository
	Quotas    QuotaRepository
	Projects  ProjectRepository
	Plugins   PluginRepository
	Tasks     TaskRepository
	TaskLogs  TaskLogRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Users:    NewUserRepository(db),
		Sessions: NewSessionRepository(db),
		APIKeys:  NewAPIKeyRepository(db),
		Perms:    NewPermissionRepository(db),
		Quotas:   NewQuotaRepository(db),
		Projects: NewProjectRepository(db),
		Plugins:  NewPluginRepository(db),
		Tasks:    NewTaskRepository(db),
		TaskLogs: NewTaskLogRepository(db),
	}
}
