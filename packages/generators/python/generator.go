// Package python generates complete, runnable Python projects from workflows.
//
// The generated project follows modern Python packaging standards (PEP 621)
// and includes everything needed to run independently of AIStudio:
//
//	project-name/
//	├── .aistudio/
//	│   └── workflow.json       # Original workflow copy
//	├── src/
//	│   ├── __init__.py
//	│   ├── data_loader.py      # Per-node modules
//	│   ├── model.py
//	│   ├── train.py
//	│   └── utils.py
//	├── tests/
//	│   ├── __init__.py
//	│   └── test_main.py
//	├── config/
//	│   └── config.yaml
//	├── data/                   # (gitignored)
//	├── models/                 # (gitignored)
//	├── outputs/                # (gitignored)
//	├── main.py                 # Entry point
//	├── workflow.json           # Workflow definition
//	├── pyproject.toml          # PEP 621 packaging
//	├── requirements.txt        # Dependencies
//	├── README.md
//	└── .gitignore
//
// Design Principles:
//   - NEVER executes code
//   - NEVER installs dependencies
//   - NEVER modifies workflow.json
//   - Output must pass python -m py_compile
//   - Output must be usable WITHOUT AIStudio
package python

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/aistudio/packages/generators/common"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// Generator generates Python projects from workflows.
type Generator struct {
	common.BaseGenerator
}

// NewGenerator creates a new Python Generator.
func NewGenerator() *Generator {
	return &Generator{
		BaseGenerator: common.BaseGenerator{
			TargetID:      common.Target("python"),
			GeneratorName: "Python Project Generator",
			GeneratorDesc: "Generates standard Python projects with proper structure, dependencies, and entry points",
			GeneratorVer:  "2.0.0",
		},
	}
}

// templateData holds all data for template rendering.
type templateData struct {
	ProjectName  string
	WorkflowName string
	WorkflowID   string
	Target       string
	Version      string
	Description  string
	Author       string
	Dependencies []string
	Packages     []string
	Nodes        []nodeData
	Imports      []importData
	Executions   []executionData
	RequiresGPU  bool
	IsYOLO       bool
}

type nodeData struct {
	ID          string
	Name        string
	Type        string
	Description string
	Config      map[string]any
}

type importData struct {
	ModuleName string
	ClassName  string
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
	est, err := g.DefaultEstimateResources(wf)
	if err != nil {
		return nil, err
	}
	est.RequiresGPU = g.requiresGPU(wf)
	return est, nil
}

func (g *Generator) RuntimeRequirement(wf *common.Workflow) (*common.RuntimeRequirement, error) {
	pkgs := collectPackages(wf)
	return &common.RuntimeRequirement{
		Name:     "python",
		Version:  "3.9+",
		Python:   ">=3.9",
		Packages: pkgs,
		Commands: []string{"python", "pip"},
		GPU:      g.requiresGPU(wf),
	}, nil
}

func (g *Generator) Plan(ctx context.Context, wf *common.Workflow, opts common.CompileOptions) (*common.CompilePlan, error) {
	return g.DefaultPlan(ctx, wf, opts)
}

func (g *Generator) Generate(ctx context.Context, wf *common.Workflow, opts common.CompileOptions) (*common.GenerateResult, error) {
	projectName := opts.ProjectName
	if projectName == "" {
		projectName = common.SanitizeName(wf.Name)
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

	isYOLO := g.isYOLOWorkflow(wf)

	td := g.buildTemplateData(wf, projectName, isYOLO)

	var files []common.GeneratedFile

	// Generate all project files using shared template renderer
	files = append(files, common.RenderToFiles(templateFS, "templates/main.py.tmpl", td, "main.py", 0755)...)
	files = append(files, common.RenderToFiles(templateFS, "templates/pyproject.toml.tmpl", td, "pyproject.toml", 0644)...)
	files = append(files, common.RenderToFiles(templateFS, "templates/requirements.txt.tmpl", td, "requirements.txt", 0644)...)
	files = append(files, common.RenderToFiles(templateFS, "templates/README.md.tmpl", td, "README.md", 0644)...)
	files = append(files, common.RenderToFiles(templateFS, "templates/gitignore.tmpl", td, ".gitignore", 0644)...)
	files = append(files, common.RenderToFiles(templateFS, "templates/config_yaml.tmpl", td, "config/config.yaml", 0644)...)
	files = append(files, common.RenderToFiles(templateFS, "templates/src_init.py.tmpl", td, "src/__init__.py", 0644)...)
	files = append(files, common.RenderToFiles(templateFS, "templates/src_utils.py.tmpl", td, "src/utils.py", 0644)...)
	files = append(files, common.RenderToFiles(templateFS, "templates/tests_init.py.tmpl", td, "tests/__init__.py", 0644)...)
	files = append(files, common.RenderToFiles(templateFS, "templates/tests_main.py.tmpl", td, "tests/test_main.py", 0644)...)

	// Generate per-node source files
	for _, nd := range td.Nodes {
		nodeTD := struct {
			ModuleName string
			NodeID     string
			NodeName   string
			NodeType   string
			ClassName  string
		}{
			ModuleName: common.SanitizeName(nd.Name),
			NodeID:     nd.ID,
			NodeName:   nd.Name,
			NodeType:   nd.Type,
			ClassName:  common.ToClassName(nd.Name),
		}

		tmplName := fmt.Sprintf("templates/%s", g.nodeTypeToTemplate(nd.Type, isYOLO))
		files = append(files, common.RenderToFiles(templateFS, tmplName, nodeTD, fmt.Sprintf("src/%s.py", common.SanitizeName(nd.Name)), 0644)...)
	}

	// workflow.json copy
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

	entryPoints := []string{filepath.Join(outputDir, "main.py")}

	return &common.GenerateResult{
		Target:      common.Target("python"),
		ProjectRoot: outputDir,
		EntryPoints: entryPoints,
		Files:       files,
		ProjectName: projectName,
	}, nil
}

func (g *Generator) buildTemplateData(wf *common.Workflow, projectName string, isYOLO bool) templateData {
	td := templateData{
		ProjectName:  projectName,
		WorkflowName: wf.Name,
		WorkflowID:   wf.ID,
		Target:       string(wf.Target),
		Version:      fmt.Sprintf("1.0.0"),
		Description:  wf.Description,
		Author:       wf.Author,
		Dependencies: collectPackages(wf),
		Packages:     collectPackages(wf),
		RequiresGPU:  g.requiresGPU(wf),
		IsYOLO:       isYOLO,
	}

	// Sort nodes in DAG order for proper execution sequence
	sortedNodes := g.sortNodesByDAG(wf)

	for _, n := range sortedNodes {
		nd := nodeData{
			ID:          n.ID,
			Name:        n.Name,
			Type:        n.Type,
			Description: n.Description,
			Config:      n.Config,
		}
		td.Nodes = append(td.Nodes, nd)

		moduleName := common.SanitizeName(n.Name)
		className := common.ToClassName(n.Name)

		td.Imports = append(td.Imports, importData{
			ModuleName: moduleName,
			ClassName:  className,
		})
		td.Executions = append(td.Executions, executionData{
			NodeID:    n.ID,
			NodeName:  n.Name,
			NodeType:  string(n.Type),
			ClassName: className,
		})
	}

	return td
}

func (g *Generator) isYOLOWorkflow(wf *common.Workflow) bool {
	for _, n := range wf.Nodes {
		if n.Type == "model_trainer" {
			if cfg, ok := n.Config["framework"]; ok {
				if fw, ok := cfg.(string); ok && strings.EqualFold(fw, "yolo") {
					return true
				}
			}
			if cfg, ok := n.Config["model"]; ok {
				if model, ok := cfg.(string); ok {
					if strings.HasPrefix(strings.ToLower(model), "yolo") {
						return true
					}
				}
			}
		}
	}
	return false
}

func (g *Generator) requiresGPU(wf *common.Workflow) bool {
	for _, n := range wf.Nodes {
		if n.Type == "model_trainer" {
			if device, ok := n.Config["device"]; ok {
				if deviceStr, ok := device.(string); ok && deviceStr == "cuda" {
					return true
				}
			}
		}
	}
	return false
}

// nodeTypeToTemplate returns the appropriate template filename for a node type.
func (g *Generator) nodeTypeToTemplate(nodeType string, isYOLO bool) string {
	if isYOLO {
		switch nodeType {
		case "model_trainer":
			return "yolo_train.py.tmpl"
		case "model_inference":
			return "yolo_predict.py.tmpl"
		}
	}

	switch nodeType {
	case "data_loader":
		return "data_loader.py.tmpl"
	case "data_preprocessor":
		return "data_preprocessor.py.tmpl"
	case "data_split":
		return "data_split.py.tmpl"
	case "model_trainer":
		return "model_trainer.py.tmpl"
	case "model_inference":
		return "model_inference.py.tmpl"
	default:
		return "src_node.py.tmpl"
	}
}

// sortNodesByDAG returns workflow nodes in topological order based on edges.
func (g *Generator) sortNodesByDAG(wf *common.Workflow) []common.Node {
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)
	for _, n := range wf.Nodes {
		inDegree[n.ID] = 0
	}
	for _, e := range wf.Edges {
		adjList[e.Source.NodeID] = append(adjList[e.Source.NodeID], e.Target.NodeID)
		inDegree[e.Target.NodeID]++
	}

	var queue []string
	for _, n := range wf.Nodes {
		if inDegree[n.ID] == 0 {
			queue = append(queue, n.ID)
		}
	}
	sort.Strings(queue)

	nodeMap := make(map[string]common.Node)
	for _, n := range wf.Nodes {
		nodeMap[n.ID] = n
	}

	var sorted []common.Node
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if n, ok := nodeMap[id]; ok {
			sorted = append(sorted, n)
		}
		for _, neighbor := range adjList[id] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(sorted) == len(wf.Nodes) {
		return sorted
	}
	return wf.Nodes
}

// ============================================================================
// Utility functions
// ============================================================================

func collectPackages(wf *common.Workflow) []string {
	pkgs := []string{
		"pyyaml>=6.0",
	}

	hasYOLO := false
	for _, n := range wf.Nodes {
		switch n.Type {
		case "model_trainer":
			if cfg, ok := n.Config["framework"]; ok {
				if fw, ok := cfg.(string); ok && strings.EqualFold(fw, "yolo") {
					hasYOLO = true
				}
			}
			if !hasYOLO {
				pkgs = append(pkgs, "torch>=2.0.0")
				pkgs = append(pkgs, "torchvision>=0.15.0")
			}
		case "model_inference":
			needsTorch := true
			if cfg, ok := n.Config["framework"]; ok {
				if fw, ok := cfg.(string); ok && strings.EqualFold(fw, "yolo") {
					needsTorch = false
					hasYOLO = true
				}
			}
			if needsTorch {
				pkgs = append(pkgs, "torch>=2.0.0")
			}
		case "data_loader", "data_preprocessor":
			pkgs = append(pkgs, "numpy>=1.24.0")
			if n.Type == "data_preprocessor" {
				pkgs = append(pkgs, "scikit-learn>=1.3.0")
			}
		case "visualization":
			pkgs = append(pkgs, "matplotlib>=3.7.0")
		case "hyperparameter_tuning":
			pkgs = append(pkgs, "optuna>=3.0.0")
		}
	}

	if hasYOLO {
		pkgs = append(pkgs, "ultralytics>=8.0.0")
	}

	return uniqueStrings(pkgs)
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

func uniqueStrings(s []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}