// Package ros2 generates complete, runnable ROS2 packages from workflows.
//
// The generated project follows ROS2 Humble/Jazzy package conventions
// and includes everything needed to build and run independently of AIStudio:
//
//	project-name/
//	├── .aistudio/
//	│   └── workflow.json
//	├── aistudio_ws/
//	│   ├── __init__.py
//	│   ├── data_loader_node.py
//	│   ├── inference_node.py
//	│   └── ...
//	├── launch/
//	│   └── workflow_launch.py
//	├── resource/
//	│   └── project_name
//	├── test/
//	│   ├── __init__.py
//	│   └── test_nodes.py
//	├── package.xml
//	├── setup.py
//	├── setup.cfg
//	├── workflow.json
//	├── README.md
//	└── .gitignore
//
// Design Principles:
//   - NEVER executes code
//   - NEVER installs dependencies
//   - NEVER modifies workflow.json
//   - Output must be usable WITHOUT AIStudio (just colcon build)
package ros2

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/aistudio/packages/generators/common"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// Generator generates ROS2 packages from workflows.
type Generator struct {
	common.BaseGenerator
}

// NewGenerator creates a new ROS2 Generator.
func NewGenerator() *Generator {
	return &Generator{
		BaseGenerator: common.BaseGenerator{
			TargetID:      common.Target("ros2"),
			GeneratorName: "ROS2 Package Generator",
			GeneratorDesc: "Generates standard ROS2 packages with nodes, launch files, and tests",
			GeneratorVer:  "1.0.0",
		},
	}
}

type templateData struct {
	ProjectName  string
	WorkflowName string
	WorkflowID   string
	Target       string
	Version      string
	Description  string
	Author       string
	PackageName  string
	Nodes        []nodeData
	Executions   []executionData
}

type nodeData struct {
	ID          string
	Name        string
	Type        string
	Description string
	NodeName    string
	Config      map[string]any
}

type executionData struct {
	NodeID        string
	NodeName      string
	NodeType      string
	NodeNameClean string
}

func (g *Generator) Validate(wf *common.Workflow) error {
	return g.DefaultValidate(wf)
}

func (g *Generator) CompileTimeValidate(ctx context.Context) error {
	return g.DefaultCompileTimeValidate(ctx)
}

func (g *Generator) EstimateResources(wf *common.Workflow) (*common.ResourceEstimate, error) {
	return g.DefaultEstimateResources(wf)
}

func (g *Generator) RuntimeRequirement(wf *common.Workflow) (*common.RuntimeRequirement, error) {
	return &common.RuntimeRequirement{
		Name:        "ros",
		Version:     "Humble/Jazzy",
		Commands:    []string{"colcon", "ros2"},
		MinMemoryMB: 2048,
		MinDiskMB:   4096,
	}, nil
}

func (g *Generator) Plan(ctx context.Context, wf *common.Workflow, opts common.CompileOptions) (*common.CompilePlan, error) {
	return g.DefaultPlan(ctx, wf, opts)
}

func nodeTypeToROS2Name(nodeType string) string {
	switch nodeType {
	case "data_loader":
		return "data_loader_node"
	case "data_preprocessor":
		return "data_preprocessor_node"
	case "data_augmentation":
		return "data_augmentation_node"
	case "data_split":
		return "data_split_node"
	case "model_trainer":
		return "model_trainer_node"
	case "model_evaluator":
		return "model_evaluator_node"
	case "model_exporter":
		return "model_exporter_node"
	case "model_inference":
		return "inference_node"
	case "feature_extractor":
		return "feature_extractor_node"
	case "hyperparameter_tuning":
		return "hyperparameter_tuner_node"
	case "visualization":
		return "visualization_node"
	case "metric_computation":
		return "metric_computation_node"
	default:
		return sanitizeName(nodeType) + "_node"
	}
}

func (g *Generator) Generate(ctx context.Context, wf *common.Workflow, opts common.CompileOptions) (*common.GenerateResult, error) {
	projectName := opts.ProjectName
	if projectName == "" {
		projectName = sanitizeName(wf.Name)
	}

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = filepath.Join(".", projectName)
	}

	if !opts.DryRun {
		if !opts.Force {
			if _, err := os.Stat(outputDir); err == nil {
				return nil, fmt.Errorf("output directory already exists: %s (use Force=true to overwrite)", outputDir)
			}
		}
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	td := g.buildTemplateData(wf, projectName)

	var files []common.GeneratedFile

	files = append(files, g.renderFile("README.md.tmpl", td, "README.md", 0644)...)
	files = append(files, g.renderFile("gitignore.tmpl", td, ".gitignore", 0644)...)
	files = append(files, g.renderFile("package_xml.tmpl", td, "package.xml", 0644)...)
	files = append(files, g.renderFile("setup_py.tmpl", td, "setup.py", 0755)...)
	files = append(files, g.renderFile("setup_cfg.tmpl", td, "setup.cfg", 0644)...)
	files = append(files, g.renderFile("launch_file.tmpl", td, fmt.Sprintf("launch/%s_launch.py", td.PackageName), 0644)...)
	files = append(files, g.renderFile("resource_marker.tmpl", td, fmt.Sprintf("resource/%s", td.PackageName), 0644)...)

	// Aistudio WS init
	wsInitData := struct {
		PackageName string
	}{PackageName: td.PackageName}
	files = append(files, g.renderFile("ws_init.tmpl", wsInitData, "aistudio_ws/__init__.py", 0644)...)

	for _, nd := range td.Nodes {
		nodeTD := struct {
			PackageName string
			NodeName    string
			NodeID      string
			NodeType    string
			ClassName   string
		}{
			PackageName: td.PackageName,
			NodeName:    nd.NodeName,
			NodeID:      nd.ID,
			NodeType:    nd.Type,
			ClassName:   toClassName(nd.NodeName),
		}
		files = append(files, g.renderFile("node_py.tmpl", nodeTD, fmt.Sprintf("aistudio_ws/%s.py", nd.NodeName), 0644)...)
	}

	// Test file
	files = append(files, g.renderFile("test_py.tmpl", td, "test/test_nodes.py", 0644)...)
	files = append(files, g.renderFile("test_init.tmpl", td, "test/__init__.py", 0644)...)

	wfJSON, err := json.MarshalIndent(wf, "", "  ")
	if err == nil {
		files = append(files, common.GeneratedFile{
			Path:    ".aistudio/workflow.json",
			Content: string(wfJSON),
			Mode:    0644,
		})
	}

	if !opts.DryRun {
		for _, f := range files {
			fullPath := filepath.Join(outputDir, f.Path)
			dir := filepath.Dir(fullPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
			if err := os.WriteFile(fullPath, []byte(f.Content), os.FileMode(f.Mode)); err != nil {
				return nil, fmt.Errorf("failed to write file %s: %w", f.Path, err)
			}
		}
	}

	entryPoints := []string{
		filepath.Join(outputDir, "setup.py"),
		filepath.Join(outputDir, fmt.Sprintf("launch/%s_launch.py", td.PackageName)),
	}

	return &common.GenerateResult{
		Target:      common.Target("ros2"),
		ProjectRoot: outputDir,
		EntryPoints: entryPoints,
		Files:       files,
		ProjectName: projectName,
	}, nil
}

func (g *Generator) buildTemplateData(wf *common.Workflow, projectName string) templateData {
	pkgName := sanitizeName(projectName)
	td := templateData{
		ProjectName:  projectName,
		WorkflowName: wf.Name,
		WorkflowID:   wf.ID,
		Target:       string(wf.Target),
		Version:      "1.0.0",
		Description:  wf.Description,
		Author:       wf.Author,
		PackageName:  pkgName,
	}

	for _, n := range wf.Nodes {
		nodeName := nodeTypeToROS2Name(n.Type)
		nd := nodeData{
			ID:          n.ID,
			Name:        n.Name,
			Type:        n.Type,
			Description: n.Description,
			NodeName:    nodeName,
			Config:      n.Config,
		}
		td.Nodes = append(td.Nodes, nd)
		td.Executions = append(td.Executions, executionData{
			NodeID:        n.ID,
			NodeName:      n.Name,
			NodeType:      n.Type,
			NodeNameClean: nodeName,
		})
	}

	return td
}

func (g *Generator) renderFile(tmplName string, data any, outputPath string, mode uint32) []common.GeneratedFile {
	tmplContent, err := templateFS.ReadFile(path.Join("templates", tmplName))
	if err != nil {
		return nil
	}

	tmpl := template.New(tmplName).Funcs(template.FuncMap{
		"title": strings.Title,
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
	})

	tmpl, err = tmpl.Parse(string(tmplContent))
	if err != nil {
		return nil
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil
	}

	return []common.GeneratedFile{
		{Path: outputPath, Content: buf.String(), Mode: mode},
	}
}

func sanitizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	var result []rune
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result = append(result, r)
		}
	}
	s := string(result)
	if s == "" {
		return "unnamed"
	}
	return s
}

func toClassName(name string) string {
	parts := strings.Split(sanitizeName(name), "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}
