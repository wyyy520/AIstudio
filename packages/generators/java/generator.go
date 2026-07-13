// Package java generates complete Java Maven projects from workflows.
//
// The generated project follows standard Maven project conventions
// and includes everything needed to build independently of AIStudio:
//
//	project-name/
//	├── .aistudio/
//	│   └── workflow.json
//	├── src/
//	│   ├── main/
//	│   │   └── java/
//	│   │       └── com/
//	│   │           └── aistudio/
//	│   │               ├── App.java
//	│   │               ├── DataLoaderNode.java
//	│   │               ├── TrainModelNode.java
//	│   │               └── ...
//	│   └── test/
//	│       └── java/
//	│           └── com/
//	│               └── aistudio/
//	│                   ├── AppTest.java
//	│                   └── NodeTest.java
//	├── pom.xml
//	├── workflow.json
//	├── README.md
//	└── .gitignore
//
// Design Principles:
//   - NEVER executes mvn
//   - NEVER compiles code
//   - NEVER modifies workflow.json
//   - Output must be buildable WITHOUT AIStudio (mvn package)
package java_

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

// Generator generates Java Maven projects from workflows.
type Generator struct {
	common.BaseGenerator
}

// NewGenerator creates a new Java Generator.
func NewGenerator() *Generator {
	return &Generator{
		BaseGenerator: common.BaseGenerator{
			TargetID:      common.Target("java"),
			GeneratorName: "Java Maven Project Generator",
			GeneratorDesc: "Generates standard Java Maven projects with proper package structure and testing",
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
	PackagePath  string
	MainClass    string
	Nodes        []nodeData
	Executions   []executionData
}

type nodeData struct {
	ID          string
	Name        string
	Type        string
	Description string
	ClassName   string
	Config      map[string]any
}

type executionData struct {
	NodeID    string
	NodeName  string
	NodeType  string
	ClassName string
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
		Name:        "java",
		Version:     "17+",
		Commands:    []string{"java", "javac", "mvn"},
		MinMemoryMB: 1024,
		MinDiskMB:   1024,
	}, nil
}

func (g *Generator) Plan(ctx context.Context, wf *common.Workflow, opts common.CompileOptions) (*common.CompilePlan, error) {
	return g.DefaultPlan(ctx, wf, opts)
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
	files = append(files, g.renderFile("pom_xml.tmpl", td, "pom.xml", 0644)...)
	files = append(files, g.renderFile("main_java.tmpl", td, fmt.Sprintf("src/main/java/%s/App.java", td.PackagePath), 0644)...)

	for _, nd := range td.Nodes {
		nodeTD := struct {
			PackageName string
			ClassName   string
			NodeID      string
			NodeName    string
			NodeType    string
		}{
			PackageName: td.PackageName,
			ClassName:   nd.ClassName,
			NodeID:      nd.ID,
			NodeName:    nd.Name,
			NodeType:    nd.Type,
		}
		files = append(files, g.renderFile("node_java.tmpl", nodeTD, fmt.Sprintf("src/main/java/%s/%s.java", td.PackagePath, nd.ClassName), 0644)...)
	}

	files = append(files, g.renderFile("test_java.tmpl", td, fmt.Sprintf("src/test/java/%s/AppTest.java", td.PackagePath), 0644)...)

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
		filepath.Join(outputDir, "pom.xml"),
		filepath.Join(outputDir, fmt.Sprintf("src/main/java/%s/App.java", td.PackagePath)),
	}

	return &common.GenerateResult{
		Target:      common.Target("java"),
		ProjectRoot: outputDir,
		EntryPoints: entryPoints,
		Files:       files,
		ProjectName: projectName,
	}, nil
}

func (g *Generator) buildTemplateData(wf *common.Workflow, projectName string) templateData {
	pkgName := "com.aistudio." + sanitizeName(projectName)
	pkgPath := strings.ReplaceAll(pkgName, ".", "/")
	mainClass := toClassName(projectName) + "App"

	td := templateData{
		ProjectName:  projectName,
		WorkflowName: wf.Name,
		WorkflowID:   wf.ID,
		Target:       string(wf.Target),
		Version:      "1.0.0",
		Description:  wf.Description,
		Author:       wf.Author,
		PackageName:  pkgName,
		PackagePath:  pkgPath,
		MainClass:    mainClass,
	}

	for _, n := range wf.Nodes {
		className := toClassName(n.Name) + "Node"
		nd := nodeData{
			ID:          n.ID,
			Name:        n.Name,
			Type:        n.Type,
			Description: n.Description,
			ClassName:   className,
			Config:      n.Config,
		}
		td.Nodes = append(td.Nodes, nd)
		td.Executions = append(td.Executions, executionData{
			NodeID:    n.ID,
			NodeName:  n.Name,
			NodeType:  n.Type,
			ClassName: className,
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
