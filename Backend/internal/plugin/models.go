package plugin

import "time"

// Status defines plugin lifecycle status.
type Status string

const (
	StatusInstalled Status = "installed"
	StatusEnabled   Status = "enabled"
	StatusDisabled  Status = "disabled"
	StatusError     Status = "error"
)

// Plugin represents a registered plugin.
type Plugin struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Entry       string    `json:"entry"`       // entry point file (e.g., "main.py")
	Config      string    `json:"config,omitempty"` // raw JSON config
	Status      Status    `json:"status"`
	Path        string    `json:"path"`        // absolute path on disk
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// PluginManifest is the structure of plugin.json.
type PluginManifest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Entry       string `json:"entry"`
}