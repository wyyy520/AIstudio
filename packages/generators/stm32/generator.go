// Package stm32 generates STM32 CubeMX-compatible project structures from workflows.
//
// The generated project follows STM32CubeMX project conventions
// and includes everything needed to build independently of AIStudio:
//
//	project-name/
//	├── .aistudio/
//	│   └── workflow.json
//	├── Core/
//	│   ├── Inc/
//	│   │   ├── main.h
//	│   │   ├── stm32_hal.h
//	│   │   └── aistudio_nodes.h
//	│   └── Src/
//	│       ├── main.c
//	│       ├── stm32_it.c
//	│       ├── aistudio_nodes.c
//	│       └── ...
//	├── {{.ProjectName}}.ioc
//	├── Makefile
//	├── workflow.json
//	├── README.md
//	└── .gitignore
//
// Design Principles:
//   - NEVER executes make
//   - NEVER modifies workflow.json
//   - Output is a PROJECT STRUCTURE generator - creates the framework
//   - Output must be usable WITHOUT AIStudio
package stm32

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

// Generator generates STM32 projects from workflows.
type Generator struct {
	common.BaseGenerator
}

// NewGenerator creates a new STM32 Generator.
func NewGenerator() *Generator {
	return &Generator{
		BaseGenerator: common.BaseGenerator{
			TargetID:      common.Target("stm32"),
			GeneratorName: "STM32 Project Generator",
			GeneratorDesc: "Generates STM32 CubeMX-compatible embedded project structures",
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
	MCUSeries    string
	Nodes        []nodeData
	Executions   []executionData
}

type nodeData struct {
	ID          string
	Name        string
	Type        string
	Description string
	Config      map[string]any
}

type executionData struct {
	NodeID   string
	NodeName string
	NodeType string
	FuncName string
}

func (g *Generator) Validate(wf *common.Workflow) error {
	return g.DefaultValidate(wf)
}

func (g *Generator) CompileTimeValidate(ctx context.Context) error {
	return g.DefaultCompileTimeValidate(ctx)
}

func (g *Generator) EstimateResources(wf *common.Workflow) (*common.ResourceEstimate, error) {
	est, err := g.DefaultEstimateResources(wf)
	if err != nil {
		return nil, err
	}
	est.MinMemoryMB = 512
	est.MinDiskMB = 512
	return est, nil
}

func (g *Generator) RuntimeRequirement(wf *common.Workflow) (*common.RuntimeRequirement, error) {
	return &common.RuntimeRequirement{
		Name:        "stm32",
		Version:     "CubeMX 6.0+",
		Commands:    []string{"make", "arm-none-eabi-gcc"},
		MinMemoryMB: 1024,
		MinDiskMB:   2048,
	}, nil
}

func (g *Generator) Plan(ctx context.Context, wf *common.Workflow, opts common.CompileOptions) (*common.CompilePlan, error) {
	return g.DefaultPlan(ctx, wf, opts)
}

func nodeTypeToFuncName(nodeType string) string {
	switch nodeType {
	case "data_loader":
		return "AISTUDIO_DataLoad"
	case "data_preprocessor":
		return "AISTUDIO_Preprocess"
	case "data_augmentation":
		return "AISTUDIO_Augment"
	case "data_split":
		return "AISTUDIO_DataSplit"
	case "model_trainer":
		return "AISTUDIO_TrainModel"
	case "model_evaluator":
		return "AISTUDIO_EvaluateModel"
	case "model_exporter":
		return "AISTUDIO_ExportModel"
	case "model_inference":
		return "AISTUDIO_RunInference"
	case "feature_extractor":
		return "AISTUDIO_ExtractFeatures"
	case "hyperparameter_tuning":
		return "AISTUDIO_TuneHyperparams"
	case "visualization":
		return "AISTUDIO_Visualize"
	case "metric_computation":
		return "AISTUDIO_ComputeMetrics"
	default:
		return "AISTUDIO_" + toClassName(nodeType)
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
	files = append(files, g.renderFile("ioc_xml.tmpl", td, fmt.Sprintf("%s.ioc", projectName), 0644)...)
	files = append(files, g.renderFile("Makefile.tmpl", td, "Makefile", 0644)...)
	files = append(files, g.renderFile("main_c.tmpl", td, "Core/Src/main.c", 0644)...)
	files = append(files, g.renderFile("main_h.tmpl", td, "Core/Inc/main.h", 0644)...)
	files = append(files, g.renderFile("stm32_it_c.tmpl", td, "Core/Src/stm32_it.c", 0644)...)
	files = append(files, g.renderFile("nodes_c.tmpl", td, "Core/Src/aistudio_nodes.c", 0644)...)
	files = append(files, g.renderFile("nodes_h.tmpl", td, "Core/Inc/aistudio_nodes.h", 0644)...)

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
		filepath.Join(outputDir, "Core/Src/main.c"),
		filepath.Join(outputDir, "Makefile"),
	}

	return &common.GenerateResult{
		Target:      common.Target("stm32"),
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
		MCUSeries:    "STM32F4",
	}

	for _, n := range wf.Nodes {
		nd := nodeData{
			ID:          n.ID,
			Name:        n.Name,
			Type:        n.Type,
			Description: n.Description,
			Config:      n.Config,
		}
		td.Nodes = append(td.Nodes, nd)
		td.Executions = append(td.Executions, executionData{
			NodeID:   n.ID,
			NodeName: n.Name,
			NodeType: n.Type,
			FuncName: nodeTypeToFuncName(n.Type),
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
