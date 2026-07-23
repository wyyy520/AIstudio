// Package compiler provides the Plugin Manifest Generator.
//
// Plugin Manifest Generator analyzes the workflow and generates plugin_manifest.json
// which lists all plugins required by the workflow.
package compiler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aistudio/packages/workflow"
)

// ============================================================================
// Plugin Manifest Types
// ============================================================================

// PluginManifest lists all plugins required by a workflow.
type PluginManifest struct {
	Version      string          `json:"version"`
	GeneratedAt  string          `json:"generated_at"`
	WorkflowID   string          `json:"workflow_id"`
	WorkflowName string          `json:"workflow_name"`
	Plugins      []PluginEntry   `json:"plugins"`
	Nodes        []NodePluginMap `json:"nodes"`
}

// PluginEntry describes a required plugin.
type PluginEntry struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Domain      string   `json:"domain"`
	Nodes       []string `json:"nodes"`
	Description string   `json:"description,omitempty"`
}

// NodePluginMap maps a node to its providing plugin.
type NodePluginMap struct {
	NodeID     string `json:"node_id"`
	NodeType   string `json:"node_type"`
	PluginName string `json:"plugin_name"`
}

// ============================================================================
// Plugin Manifest Generator
// ============================================================================

// PluginManifestGenerator generates plugin manifests from workflows.
type PluginManifestGenerator struct {
	pluginRegistry map[string]*PluginEntry
}

// NewPluginManifestGenerator creates a new PluginManifestGenerator.
func NewPluginManifestGenerator() *PluginManifestGenerator {
	return &PluginManifestGenerator{
		pluginRegistry: make(map[string]*PluginEntry),
	}
}

// RegisterPlugin registers a plugin for manifest generation.
func (g *PluginManifestGenerator) RegisterPlugin(entry *PluginEntry) {
	g.pluginRegistry[entry.Name] = entry
}

// Generate analyzes a workflow and generates a plugin manifest.
func (g *PluginManifestGenerator) Generate(wf *workflow.Workflow) *PluginManifest {
	manifest := &PluginManifest{
		Version:      "1.0.0",
		GeneratedAt:  time.Now().Format(time.RFC3339),
		WorkflowID:   wf.ID,
		WorkflowName: wf.Name,
		Plugins:      []PluginEntry{},
		Nodes:        []NodePluginMap{},
	}

	// Group nodes by plugin/domain
	nodeToPlugin := make(map[string]string)
	pluginNodes := make(map[string][]string)

	for _, node := range wf.Nodes {
		// Determine plugin from node domain or type
		pluginName := g.resolvePluginForNode(node)
		if pluginName == "" {
			pluginName = "builtin" // Built-in nodes don't require plugins
		}

		nodeToPlugin[node.ID] = pluginName
		pluginNodes[pluginName] = append(pluginNodes[pluginName], node.ID)

		manifest.Nodes = append(manifest.Nodes, NodePluginMap{
			NodeID:     node.ID,
			NodeType:   string(node.Type),
			PluginName: pluginName,
		})
	}

	// Build plugin entries
	for pluginName, nodeIDs := range pluginNodes {
		if pluginName == "builtin" {
			continue // Skip builtin nodes
		}

		entry := &PluginEntry{
			Name:   pluginName,
			Nodes:  nodeIDs,
			Domain: pluginName,
		}

		// Get version from registry if available
		if regEntry, ok := g.pluginRegistry[pluginName]; ok {
			entry.Version = regEntry.Version
			entry.Description = regEntry.Description
		} else {
			entry.Version = "latest"
		}

		manifest.Plugins = append(manifest.Plugins, *entry)
	}

	return manifest
}

// resolvePluginForNode determines which plugin provides a node.
func (g *PluginManifestGenerator) resolvePluginForNode(node workflow.Node) string {
	// Check if node has explicit plugin info
	if node.Plugin != "" {
		return node.Plugin
	}

	// Infer plugin from domain
	if node.Domain != "" {
		switch node.Domain {
		case "python":
			return "python"
		case "matlab":
			return "matlab"
		case "stm32":
			return "stm32"
		case "ansys":
			return "ansys"
		case "ros2":
			return "ros2"
		default:
			return node.Domain
		}
	}

	// Infer from node type
	nodeType := string(node.Type)
	if g.isBuiltinNode(nodeType) {
		return "builtin"
	}

	return ""
}

// isBuiltinNode checks if a node type is built-in (no plugin required).
func (g *PluginManifestGenerator) isBuiltinNode(nodeType string) bool {
	builtinNodes := map[string]bool{
		"dataset":      true,
		"datasource":   true,
		"csv":          true,
		"json":         true,
		"excel":        true,
		"image":        true,
		"video":        true,
		"ifelse":       true,
		"switch":       true,
		"loop":         true,
		"while":        true,
		"foreach":      true,
		"merge":        true,
		"delay":        true,
		"parallel":     true,
		"timer":        true,
		"log":          true,
		"print":        true,
		"debug":        true,
		"export":       true,
		"save":         true,
		"notification": true,
	}
	return builtinNodes[nodeType]
}

// WriteManifest writes the plugin manifest to a file.
func (g *PluginManifestGenerator) WriteManifest(manifest *PluginManifest, outputDir string) (string, error) {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal plugin_manifest.json: %w", err)
	}

	path := filepath.Join(outputDir, "plugin_manifest.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write plugin_manifest.json: %w", err)
	}

	return path, nil
}

// GenerateAndWrite generates a plugin manifest and writes it to disk.
func (g *PluginManifestGenerator) GenerateAndWrite(wf *workflow.Workflow, outputDir string) (string, error) {
	manifest := g.Generate(wf)
	return g.WriteManifest(manifest, outputDir)
}
