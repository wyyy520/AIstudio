package plugin

import (
	"time"
)

// Status defines plugin lifecycle status.
type Status string

const (
	StatusNotInstalled Status = "not_installed"
	StatusInstalling   Status = "installing"
	StatusInstalled    Status = "installed"
	StatusUpdating     Status = "updating"
	StatusError        Status = "error"
	StatusEnabled      Status = "enabled"
	StatusDisabled     Status = "disabled"
)

// PluginType categorizes plugins by capability domain.
type PluginType string

const (
	PluginTypeVision      PluginType = "vision"
	PluginTypeNLP         PluginType = "nlp"
	PluginTypeTimeseries  PluginType = "timeseries"
	PluginTypeSimulation  PluginType = "simulation"
	PluginTypeMCP         PluginType = "mcp"
	PluginTypeSystem      PluginType = "system"
)

// PluginSource indicates where the plugin was obtained from.
type PluginSource string

const (
	PluginSourceMarket   PluginSource = "market"
	PluginSourceLocal    PluginSource = "local"
	PluginSourceGit      PluginSource = "git"
)

// Plugin represents a registered plugin with full metadata.
type Plugin struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Author       string                 `json:"author"`
	Type         PluginType             `json:"type"`
	Description  string                 `json:"description"`
	Entry        string                 `json:"entry"`
	Source       PluginSource           `json:"source"`
	Path         string                 `json:"path"`
	Config       string                 `json:"config,omitempty"`
	Dependencies []Dependency           `json:"dependencies"`
	Status       Status                 `json:"status"`
	Enabled      bool                   `json:"enabled"`
	Nodes        []NodeRegistration     `json:"nodes"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// NodeRegistration describes a workflow node provided by a plugin.
type NodeRegistration struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Inputs      []PortInfo `json:"inputs"`
	Outputs     []PortInfo `json:"outputs"`
}

// PortInfo describes a port on a node registration.
type PortInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// PluginManifest is the structure of plugin.json on disk.
type PluginManifest struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Version      string             `json:"version"`
	Author       string             `json:"author"`
	Type         string             `json:"type"`
	Description  string             `json:"description"`
	Entry        string             `json:"entry"`
	Source       string             `json:"source"`
	Dependencies []Dependency       `json:"dependencies"`
	Nodes        []NodeRegistration `json:"nodes"`
}

// PluginSummary is a lightweight representation for list views.
type PluginSummary struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Author      string     `json:"author"`
	Type        PluginType `json:"type"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	Enabled     bool       `json:"enabled"`
	NodeCount   int        `json:"node_count"`
}

// ToSummary converts a Plugin to a PluginSummary.
func (p *Plugin) ToSummary() PluginSummary {
	return PluginSummary{
		ID:          p.ID,
		Name:        p.Name,
		Version:     p.Version,
		Author:      p.Author,
		Type:        p.Type,
		Description: p.Description,
		Status:      p.Status,
		Enabled:     p.Enabled,
		NodeCount:   len(p.Nodes),
	}
}