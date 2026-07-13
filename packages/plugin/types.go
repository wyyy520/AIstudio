package plugin

import "time"

// PluginType categorizes plugins by capability domain.
type PluginType string

const (
	PluginTypeVision     PluginType = "vision"
	PluginTypeNLP        PluginType = "nlp"
	PluginTypeTimeseries PluginType = "timeseries"
	PluginTypeSimulation PluginType = "simulation"
	PluginTypeMCP        PluginType = "mcp"
	PluginTypeSystem     PluginType = "system"
)

// PluginStatus defines plugin lifecycle status.
type PluginStatus string

const (
	StatusNotInstalled PluginStatus = "not_installed"
	StatusInstalled    PluginStatus = "installed"
	StatusEnabled      PluginStatus = "enabled"
	StatusDisabled     PluginStatus = "disabled"
	StatusError        PluginStatus = "error"
)

// Plugin represents a registered plugin with full metadata.
type Plugin struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	Author      string       `json:"author"`
	Type        PluginType   `json:"category"`
	Description string       `json:"description"`
	Status      PluginStatus `json:"status"`
	Enabled     bool         `json:"enabled"`
	Nodes       []PluginNode `json:"nodes"`
	Manifest    *ManifestV2  `json:"-"`
	SourceDir   string       `json:"-"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

// PluginSummary is a lightweight representation for list views.
type PluginSummary struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	Author      string       `json:"author"`
	Type        PluginType   `json:"category"`
	Description string       `json:"description"`
	Status      PluginStatus `json:"status"`
	Enabled     bool         `json:"enabled"`
	NodeCount   int          `json:"nodeCount"`
	Kind        string       `json:"kind"`
}

// ToSummary converts a Plugin to a PluginSummary.
func (p *Plugin) ToSummary() PluginSummary {
	kind := ""
	if p.Manifest != nil {
		kind = p.Manifest.Kind
	}
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
		Kind:        kind,
	}
}
