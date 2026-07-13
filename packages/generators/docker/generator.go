// Package docker generates complete Docker-based projects from workflows.
//
// The generated project includes Dockerfiles, docker-compose configuration,
// and everything needed to build and run independently of AIStudio:
//
//	project-name/
//	├── .aistudio/
//	│   └── workflow.json
//	├── config/
//	│   ├── docker-compose.override.yml
//	│   └── service_configs/
//	├── Dockerfile
//	├── docker-compose.yml
//	├── entrypoint.sh
//	├── .dockerignore
//	├── workflow.json
//	├── README.md
//	└── .gitignore
//
// Design Principles:
//   - NEVER executes docker build
//   - NEVER runs containers
//   - NEVER modifies workflow.json
//   - Output must be usable WITHOUT AIStudio
package docker

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

// Generator generates Docker projects from workflows.
type Generator struct {
	common.BaseGenerator
}

// NewGenerator creates a new Docker Generator.
func NewGenerator() *Generator {
	return &Generator{
		BaseGenerator: common.BaseGenerator{
			TargetID:      common.Target("docker"),
			GeneratorName: "Docker Project Generator",
			GeneratorDesc: "Generates Docker-based projects with multi-service orchestration",
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
	Services     []serviceData
}

type serviceData struct {
	Name        string
	NodeID      string
	NodeType    string
	Description string
	Image       string
	Ports       []string
	Volumes     []string
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
	est.MinDiskMB = 2048
	return est, nil
}

func (g *Generator) RuntimeRequirement(wf *common.Workflow) (*common.RuntimeRequirement, error) {
	return &common.RuntimeRequirement{
		Name:        "docker",
		Version:     "20.10+",
		Commands:    []string{"docker", "docker-compose"},
		MinMemoryMB: 2048,
		MinDiskMB:   10240,
	}, nil
}

func (g *Generator) Plan(ctx context.Context, wf *common.Workflow, opts common.CompileOptions) (*common.CompilePlan, error) {
	return g.DefaultPlan(ctx, wf, opts)
}

func nodeTypeToImage(nodeType string) string {
	switch nodeType {
	case "data_loader":
		return "python:3.11-slim"
	case "data_preprocessor":
		return "python:3.11-slim"
	case "data_augmentation":
		return "python:3.11-slim"
	case "data_split":
		return "python:3.11-slim"
	case "model_trainer":
		return "python:3.11-slim"
	case "model_evaluator":
		return "python:3.11-slim"
	case "model_exporter":
		return "python:3.11-slim"
	case "model_inference":
		return "python:3.11-slim"
	case "feature_extractor":
		return "python:3.11-slim"
	case "hyperparameter_tuning":
		return "python:3.11-slim"
	case "visualization":
		return "python:3.11-slim"
	case "metric_computation":
		return "python:3.11-slim"
	default:
		return "python:3.11-slim"
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
	files = append(files, g.renderFile("Dockerfile.tmpl", td, "Dockerfile", 0644)...)
	files = append(files, g.renderFile("docker_compose.tmpl", td, "docker-compose.yml", 0644)...)
	files = append(files, g.renderFile("dockerignore.tmpl", td, ".dockerignore", 0644)...)
	files = append(files, g.renderFile("entrypoint_sh.tmpl", td, "entrypoint.sh", 0755)...)

	// Generate per-service config files if needed
	hasServices := len(td.Services) > 0
	if hasServices {
		for _, svc := range td.Services {
			svcData := struct {
				ServiceName string
				NodeType    string
			}{
				ServiceName: svc.Name,
				NodeType:    svc.NodeType,
			}
			files = append(files, g.renderFile("service_script.tmpl", svcData, fmt.Sprintf("config/service_%s.py", svc.Name), 0644)...)
		}
	}

	// docker-compose override template
	files = append(files, g.renderFile("compose_override.tmpl", td, "config/docker-compose.override.yml", 0644)...)

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
		filepath.Join(outputDir, "docker-compose.yml"),
		filepath.Join(outputDir, "Dockerfile"),
	}

	return &common.GenerateResult{
		Target:      common.Target("docker"),
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
		svc := serviceData{
			Name:        sanitizeName(n.Name),
			NodeID:      n.ID,
			NodeType:    n.Type,
			Description: n.Description,
			Image:       nodeTypeToImage(n.Type),
			Ports:       []string{},
			Volumes:     []string{"./data:/data", "./models:/models"},
		}
		td.Services = append(td.Services, svc)
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
