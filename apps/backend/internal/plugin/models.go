package plugin

import (
	"time"
)

// ============================================================================
// V2 Data Structures — Pure Declaration
// ============================================================================

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

// Status defines plugin lifecycle status.
type Status string

const (
	StatusNotInstalled Status = "not_installed"
	StatusInstalled    Status = "installed"
	StatusEnabled      Status = "enabled"
	StatusDisabled     Status = "disabled"
	StatusError        Status = "error"
)

// ManifestV2 represents a parsed plugin manifest v2 from plugin.json.
type ManifestV2 struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	Version          string       `json:"version"`
	MinSchemaVersion string       `json:"min_schema_version"`
	Kind             string       `json:"kind"`
	Description      string       `json:"description,omitempty"`
	Author           string       `json:"author,omitempty"`
	Nodes            []PluginNode `json:"nodes"`
	RuntimeBundle    string       `json:"runtime_bundle,omitempty"`
	SupportedTargets []string     `json:"supported_targets,omitempty"`
	Language         string       `json:"language,omitempty"`
	EntryPoint       string       `json:"entry_point,omitempty"`
}

// ManifestV2Kind enumerates valid manifest kinds.
const (
	ManifestKindAlgorithm = "algorithm"
	ManifestKindRuntime   = "runtime"
	ManifestKindSystem    = "system"
	ManifestKindAdapter   = "adapter"
	ManifestKindTool      = "tool"
)

// ValidManifestKinds returns all valid manifest kinds.
func ValidManifestKinds() []string {
	return []string{
		ManifestKindAlgorithm,
		ManifestKindRuntime,
		ManifestKindSystem,
		ManifestKindAdapter,
		ManifestKindTool,
	}
}

// PluginNode defines a node type provided by a plugin manifest.
type PluginNode struct {
	Type         string            `json:"type"`
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	Inputs       []PortInfo        `json:"inputs,omitempty"`
	Outputs      []PortInfo        `json:"outputs,omitempty"`
	ConfigSchema *ConfigSchema     `json:"config_schema,omitempty"`
}

// ConfigSchema holds a JSON Schema for validating node configuration.
type ConfigSchema struct {
	Type       string                   `json:"type"`
	Properties map[string]SchemaProperty `json:"properties,omitempty"`
	Required   []string                 `json:"required,omitempty"`
}

// SchemaProperty describes a single property in a JSON Schema.
type SchemaProperty struct {
	Type        string   `json:"type"`
	Default     any      `json:"default,omitempty"`
	Description string   `json:"description,omitempty"`
	Minimum     *float64 `json:"minimum,omitempty"`
	Maximum     *float64 `json:"maximum,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Items       *struct {
		Type string `json:"type"`
	} `json:"items,omitempty"`
}

// PortInfo describes a typed port on a node.
type PortInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// Plugin represents a registered plugin with full metadata.
type Plugin struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Author      string     `json:"author"`
	Type        PluginType `json:"category"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	Enabled     bool       `json:"enabled"`
	Nodes       []PluginNode `json:"nodes"`
	Manifest    *ManifestV2  `json:"-"`
	SourceDir   string     `json:"-"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// PluginSummary is a lightweight representation for list views.
type PluginSummary struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Author      string     `json:"author"`
	Type        PluginType `json:"category"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	Enabled     bool       `json:"enabled"`
	NodeCount   int        `json:"nodeCount"`
	Kind        string     `json:"kind"`
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
