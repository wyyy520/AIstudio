// Package matlab generates complete, runnable MATLAB projects from workflows.
//
// The generated project follows MATLAB project structure conventions
// and includes everything needed to run independently of AIStudio:
//
//	project-name/
//	├── .aistudio/
//	│   └── workflow.json       # Original workflow copy
//	├── scripts/
//	│   ├── load_dataset.m      # Per-node scripts
//	│   ├── train_model.m
//	│   └── ...
//	├── results/                # Output directory
//	├── startup.m               # Setup script
//	├── run.m                   # Main execution script
//	├── workflow.json           # Workflow definition
//	├── README.md
//	└── .gitignore
//
// Design Principles:
//   - NEVER executes code
//   - NEVER modifies workflow.json
//   - Output must be usable WITHOUT AIStudio
package matlab

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

// Generator generates MATLAB projects from workflows.
type Generator struct {
	common.BaseGenerator
}

// NewGenerator creates a new MATLAB Generator.
func NewGenerator() *Generator {
	return &Generator{
		BaseGenerator: common.BaseGenerator{
			TargetID:      common.Target("matlab"),
			GeneratorName: "MATLAB Project Generator",
			GeneratorDesc: "Generates standard MATLAB projects with proper structure, entry points, and run scripts",
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
	Nodes        []nodeData
	Executions   []executionData
}

type nodeData struct {
	ID          string
	Name        string
	Type        string
	Description string
	ScriptName  string
	Config      map[string]any
}

type executionData struct {
	NodeID     string
	NodeName   string
	NodeType   string
	ScriptName string
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
		Name:        "matlab",
		Version:     "R2020b+",
		Commands:    []string{"matlab"},
		MinMemoryMB: 4096,
		MinDiskMB:   1024,
	}, nil
}

func (g *Generator) Plan(ctx context.Context, wf *common.Workflow, opts common.CompileOptions) (*common.CompilePlan, error) {
	return g.DefaultPlan(ctx, wf, opts)
}

func nodeTypeToScriptName(nodeType string) string {
	switch nodeType {
	case "data_loader":
		return "load_dataset"
	case "data_preprocessor":
		return "preprocess_data"
	case "data_augmentation":
		return "augment_data"
	case "data_split":
		return "split_data"
	case "model_trainer":
		return "train_model"
	case "model_evaluator":
		return "evaluate_model"
	case "model_exporter":
		return "export_model"
	case "model_inference":
		return "run_inference"
	case "feature_extractor":
		return "extract_features"
	case "hyperparameter_tuning":
		return "tune_hyperparameters"
	case "visualization":
		return "visualize_results"
	case "metric_computation":
		return "compute_metrics"
	default:
		return sanitizeName(nodeType)
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
	files = append(files, g.renderFile("startup_m.tmpl", td, "startup.m", 0644)...)
	files = append(files, g.renderFile("run_m.tmpl", td, "run.m", 0644)...)

	for _, nd := range td.Nodes {
		nodeTD := struct {
			ScriptName      string
			NodeID          string
			NodeName        string
			NodeType        string
			ClassName       string
			NodeDescription string
		}{
			ScriptName:      nd.ScriptName,
			NodeID:          nd.ID,
			NodeName:        nd.Name,
			NodeType:        nd.Type,
			ClassName:       toClassName(nd.Name),
			NodeDescription: nd.Description,
		}
		files = append(files, g.renderFile("script_m.tmpl", nodeTD, fmt.Sprintf("scripts/%s.m", nd.ScriptName), 0644)...)
	}

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

	entryPoints := []string{filepath.Join(outputDir, "startup.m"), filepath.Join(outputDir, "run.m")}

	return &common.GenerateResult{
		Target:      common.Target("matlab"),
		ProjectRoot: outputDir,
		EntryPoints: entryPoints,
		Files:       files,
		ProjectName: projectName,
	}, nil
}

func (g *Generator) buildTemplateData(wf *common.Workflow, projectName string) templateData {
	td := templateData{
		ProjectName:  projectName,
		WorkflowName: wf.Name,
		WorkflowID:   wf.ID,
		Target:       string(wf.Target),
		Version:      "1.0.0",
		Description:  wf.Description,
		Author:       wf.Author,
	}

	for _, n := range wf.Nodes {
		scriptName := nodeTypeToScriptName(n.Type)
		nd := nodeData{
			ID:          n.ID,
			Name:        n.Name,
			Type:        n.Type,
			Description: n.Description,
			ScriptName:  scriptName,
			Config:      n.Config,
		}
		td.Nodes = append(td.Nodes, nd)
		td.Executions = append(td.Executions, executionData{
			NodeID:     n.ID,
			NodeName:   n.Name,
			NodeType:   n.Type,
			ScriptName: scriptName,
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
