package auth

import (
	"strings"
	"sync"

	"github.com/aistudio/backend/internal/database/models"
	"gorm.io/gorm"
)

type ResourceAction string

const (
	ActionCreate ResourceAction = "create"
	ActionRead   ResourceAction = "read"
	ActionUpdate ResourceAction = "update"
	ActionDelete ResourceAction = "delete"
	ActionExecute ResourceAction = "execute"
	ActionManage  ResourceAction = "manage"
	ActionUpload  ResourceAction = "upload"
)

type Resource string

const (
	ResourceUser     Resource = "user"
	ResourceProject   Resource = "project"
	ResourceWorkflow  Resource = "workflow"
	ResourceTask      Resource = "task"
	ResourcePlugin    Resource = "plugin"
	ResourceAPIKey    Resource = "api_key"
	ResourceAgent     Resource = "agent"
	ResourceQuota     Resource = "quota"
	ResourceSystem    Resource = "system"
	ResourceLog       Resource = "log"
)

type PermissionManager struct {
	db     *gorm.DB
	mu     sync.RWMutex
	cached map[string]map[string]bool
}

func NewPermissionManager(db *gorm.DB) *PermissionManager {
	m := &PermissionManager{
		db:     db,
		cached: make(map[string]map[string]bool),
	}
	m.seedDefaults()
	return m
}

func (m *PermissionManager) seedDefaults() {
	defaultPermissions := []struct {
		role     string
		resource string
		actions  []string
	}{
		{"admin", "user", []string{"create", "read", "update", "delete", "manage"}},
		{"admin", "project", []string{"create", "read", "update", "delete", "manage"}},
		{"admin", "workflow", []string{"create", "read", "update", "delete", "execute", "manage"}},
		{"admin", "task", []string{"create", "read", "update", "delete", "manage"}},
		{"admin", "plugin", []string{"create", "read", "update", "delete", "execute", "upload", "manage"}},
		{"admin", "api_key", []string{"create", "read", "update", "delete", "manage"}},
		{"admin", "agent", []string{"create", "read", "execute", "manage"}},
		{"admin", "quota", []string{"read", "update", "manage"}},
		{"admin", "system", []string{"read", "manage"}},
		{"admin", "log", []string{"read", "delete"}},

		{"developer", "project", []string{"create", "read", "update", "delete"}},
		{"developer", "workflow", []string{"create", "read", "update", "delete", "execute"}},
		{"developer", "task", []string{"create", "read", "update", "delete"}},
		{"developer", "plugin", []string{"read", "execute", "upload"}},
		{"developer", "api_key", []string{"create", "read", "update", "delete"}},
		{"developer", "agent", []string{"create", "read", "execute"}},

		{"user", "project", []string{"create", "read", "update", "delete"}},
		{"user", "workflow", []string{"create", "read", "update", "delete", "execute"}},
		{"user", "task", []string{"create", "read"}},
		{"user", "plugin", []string{"read", "execute"}},
		{"user", "api_key", []string{"create", "read", "update", "delete"}},
		{"user", "agent", []string{"read", "execute"}},
	}

	for _, p := range defaultPermissions {
		for _, action := range p.actions {
			key := action + ":" + p.resource
			if m.cached[p.role] == nil {
				m.cached[p.role] = make(map[string]bool)
			}
			m.cached[p.role][key] = true
		}
	}
}

func (m *PermissionManager) HasPermission(role string, resource Resource, action ResourceAction) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rolePerms, ok := m.cached[role]
	if !ok {
		return false
	}
	if rolePerms["manage:"+string(resource)] {
		return true
	}
	return rolePerms[string(action)+":"+string(resource)]
}

func (m *PermissionManager) RequirePermission(role string, resource Resource, action ResourceAction) error {
	if !m.HasPermission(role, resource, action) {
		return ErrPermissionDenied
	}
	return nil
}

func (m *PermissionManager) GetRole(user *models.User) string {
	return user.Role
}

func (m *PermissionManager) GetAllPermissions() map[string][]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]string)
	for role, perms := range m.cached {
		for key := range perms {
			result[role] = append(result[role], key)
		}
	}
	return result
}

func (m *PermissionManager) Roles() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	roles := make([]string, 0, len(m.cached))
	for role := range m.cached {
		roles = append(roles, role)
	}
	return roles
}

func (m *PermissionManager) SyncToDB() error {
	for role, perms := range m.cached {
		for key := range perms {
			action, resource := parsePermKey(key)
			if action == "" || resource == "" {
				continue
			}
			var count int64
			m.db.Model(&models.Permission{}).
				Where("role = ? AND resource = ? AND action = ?", role, resource, action).
				Count(&count)
			if count == 0 {
				m.db.Create(&models.Permission{
					Role:     role,
					Resource: resource,
					Action:   action,
				})
			}
		}
	}
	return nil
}

func parsePermKey(key string) (action, resource string) {
	idx := strings.IndexByte(key, ':')
	if idx < 0 {
		return "", ""
	}
	return key[:idx], key[idx+1:]
}
