package security

import (
	"fmt"
	"strings"
	"sync"
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
	ResourceProject  Resource = "project"
	ResourceWorkflow Resource = "workflow"
	ResourceTask     Resource = "task"
	ResourcePlugin   Resource = "plugin"
	ResourceAPIKey   Resource = "api_key"
	ResourceAgent    Resource = "agent"
	ResourceQuota    Resource = "quota"
	ResourceSystem   Resource = "system"
	ResourceLog      Resource = "log"
)

type PermissionChecker struct {
	mu     sync.RWMutex
	cached map[string]map[string]bool
}

func NewPermissionChecker() *PermissionChecker {
	c := &PermissionChecker{
		cached: make(map[string]map[string]bool),
	}
	c.seedDefaults()
	return c
}

func (c *PermissionChecker) seedDefaults() {
	type permDef struct {
		role     string
		resource string
		actions  []string
	}
	defaults := []permDef{
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
	for _, p := range defaults {
		for _, action := range p.actions {
			key := action + ":" + p.resource
			if c.cached[p.role] == nil {
				c.cached[p.role] = make(map[string]bool)
			}
			c.cached[p.role][key] = true
		}
	}
}

func (c *PermissionChecker) HasPermission(role string, resource Resource, action ResourceAction) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	rolePerms, ok := c.cached[role]
	if !ok {
		return false
	}
	if rolePerms["manage:"+string(resource)] {
		return true
	}
	return rolePerms[string(action)+":"+string(resource)]
}

func (c *PermissionChecker) RequirePermission(role string, resource Resource, action ResourceAction) error {
	if !c.HasPermission(role, resource, action) {
		return fmt.Errorf("permission denied: %s cannot %s %s", role, action, resource)
	}
	return nil
}

func (c *PermissionChecker) AllPermissions() map[string][]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[string][]string)
	for role, perms := range c.cached {
		for key := range perms {
			result[role] = append(result[role], key)
		}
	}
	return result
}

func (c *PermissionChecker) Roles() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	roles := make([]string, 0, len(c.cached))
	for role := range c.cached {
		roles = append(roles, role)
	}
	return roles
}

func (c *PermissionChecker) AddPermission(role string, resource Resource, actions ...ResourceAction) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cached[role] == nil {
		c.cached[role] = make(map[string]bool)
	}
	for _, action := range actions {
		key := strings.ToLower(string(action)) + ":" + strings.ToLower(string(resource))
		c.cached[role][key] = true
	}
}
