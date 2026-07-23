// Package ansys generates complete ANSYS simulation projects from workflows.
//
// The generated project follows ANSYS Workbench project structure conventions
// and includes everything needed to run independently of AIStudio:
//
//	project-name/
//	├── .aistudio/
//	│   └── workflow.json          # Original workflow copy
//	├── workbench.wbpj             # ANSYS Workbench project file
//	├── mechanical/                # Mechanical simulation files
//	│   ├── model.dat              # Mechanical APDL input
//	│   └── solver.out             # Solver output
//	├── fluent/                    # Fluent CFD files
//	│   ├── case.cas               # Fluent case file
//	│   └── journal.jou            # Fluent journal
//	├── scripts/                   # Generated APDL scripts
//	│   ├── mesh.apdl
//	│   ├── solve.apdl
//	│   └── post.apdl
//	├── results/                   # Output directory
//	├── workflow.json
//	├── README.md
//	└── .gitignore
//
// Design Principles:
//   - NEVER executes ANSYS
//   - NEVER modifies workflow.json
//   - Output must be usable WITHOUT AIStudio
package ansys

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

// Generator generates ANSYS projects from workflows.
type Generator struct {
	common.BaseGenerator
}

// NewGenerator creates a new ANSYS Generator.
func NewGenerator() *Generator {
	return &Generator{
		BaseGenerator: common.BaseGenerator{
			TargetID:      common.Target("ansys"),
			GeneratorName: "ANSYS Project Generator",
			GeneratorDesc: "Generates standard ANSYS Workbench projects for FEA, CFD, and multiphysics simulations",
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
	Nodes        []ansysNodeData
	Executions   []executionData
}

type ansysNodeData struct {
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
	DependsOn  []string
}

// Generate creates an ANSYS Workbench project from the workflow.
func (g *Generator) Generate(ctx context.Context, wf *common.Workflow, opts *common.CompileOptions) (*common.GenerateResult, error) {
	if err := g.Validate(wf); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	projectName := sanitizeName(wf.Name)
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = path.Join("generated", projectName, "ansys")
	}

	// Create project structure
	dirs := []string{
		filepath.Join(outputDir, ".aistudio"),
		filepath.Join(outputDir, "mechanical"),
		filepath.Join(outputDir, "fluent"),
		filepath.Join(outputDir, "scripts"),
		filepath.Join(outputDir, "results"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", d, err)
		}
	}

	// Build template data
	td := templateData{
		ProjectName:  projectName,
		WorkflowName: wf.Name,
		WorkflowID:   wf.ID,
		Target:       "ansys",
		Version:      "2023 R2",
		Description:  wf.Description,
		Author:       "AIStudio",
	}

	// Map nodes
	nodes := wf.Nodes
	if wf.SortedNodes != nil && len(wf.SortedNodes) > 0 {
		nodes = wf.SortedNodes
	}

	for _, node := range nodes {
		nd := ansysNodeData{
			ID:          node.ID,
			Name:        sanitizeName(node.Name),
			Type:        node.Type,
			Description: node.Description,
			ScriptName:  nodeTypeToScript(node.Type),
			Config:      node.Config,
		}
		td.Nodes = append(td.Nodes, nd)

		exec := executionData{
			NodeID:     node.ID,
			NodeName:   nd.Name,
			NodeType:   node.Type,
			ScriptName: nd.ScriptName,
			DependsOn:  getNodeDependencies(node.ID, wf.Edges),
		}
		td.Executions = append(td.Executions, exec)
	}

	files := []common.GeneratedFile{}

	// Render templates
	templatesToRender := []struct {
		tmplName string
		outPath  string
	}{
		{"workbench.wbpj.tmpl", filepath.Join(outputDir, "workbench.wbpj")},
		{"journal.wbjn.tmpl", filepath.Join(outputDir, "journal.wbjn")},
		{"mechanical/model.dat.tmpl", filepath.Join(outputDir, "mechanical", "model.dat")},
		{"scripts/mesh.apdl.tmpl", filepath.Join(outputDir, "scripts", "mesh.apdl")},
		{"scripts/solve.apdl.tmpl", filepath.Join(outputDir, "scripts", "solve.apdl")},
		{"scripts/post.apdl.tmpl", filepath.Join(outputDir, "scripts", "post.apdl")},
		{"README.md.tmpl", filepath.Join(outputDir, "README.md")},
		{".gitignore.tmpl", filepath.Join(outputDir, ".gitignore")},
	}

	for _, t := range templatesToRender {
		content, err := g.renderTemplate(t.tmplName, td)
		if err != nil {
			return nil, fmt.Errorf("render %s: %w", t.tmplName, err)
		}
		files = append(files, common.GeneratedFile{
			Path:    t.outPath,
			Content: []byte(content),
			Mode:    0644,
		})
	}

	// Write workflow.json (original copy)
	wfJSON, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal workflow: %w", err)
	}
	files = append(files, common.GeneratedFile{
		Path:    filepath.Join(outputDir, ".aistudio", "workflow.json"),
		Content: wfJSON,
		Mode:    0644,
	})

	entryPoints := []string{
		filepath.Join(outputDir, "workbench.wbpj"),
		filepath.Join(outputDir, "journal.wbjn"),
	}

	return &common.GenerateResult{
		Target:      common.Target("ansys"),
		ProjectRoot: outputDir,
		EntryPoints: entryPoints,
		Files:       files,
	}, nil
}

// Validate checks if the workflow can be generated as an ANSYS project.
func (g *Generator) Validate(wf *common.Workflow) error {
	if wf == nil {
		return fmt.Errorf("workflow is nil")
	}
	if len(wf.Nodes) == 0 {
		return fmt.Errorf("workflow has no nodes")
	}
	// ANSYS needs at least one supported simulation node
	hasSimulation := false
	for _, node := range wf.Nodes {
		if isANSYSNode(node.Type) {
			hasSimulation = true
			break
		}
	}
	if !hasSimulation {
		return fmt.Errorf("no ANSYS simulation nodes found in workflow")
	}
	return nil
}

// RuntimeRequirement returns runtime dependencies for ANSYS projects.
func (g *Generator) RuntimeRequirement(wf *common.Workflow) (*common.RuntimeRequirement, error) {
	return &common.RuntimeRequirement{
		Runtime:    "ansys",
		MinVersion: "2023 R1",
		Packages:   []string{},
		Tools: []string{
			"ANSYS Workbench",
			"ANSYS Mechanical APDL",
		},
	}, nil
}

// EstimateResources estimates disk space for the generated project.
func (g *Generator) EstimateResources(wf *common.Workflow) (*common.ResourceEstimate, error) {
	fileCount := 8 + len(wf.Nodes)*2
	estimatedSize := int64(fileCount * 2048)

	hasFluent := false
	for _, node := range wf.Nodes {
		if strings.Contains(node.Type, "fluent") {
			hasFluent = true
			break
		}
	}
	if hasFluent {
		estimatedSize += 100 * 1024 * 1024 // Fluent typically generates larger files
	}

	return &common.ResourceEstimate{
		EstimatedFiles: fileCount,
		EstimatedSize:  estimatedSize,
	}, nil
}

// CompileTimeValidate performs compile-time validation.
func (g *Generator) CompileTimeValidate(ctx context.Context) error {
	return nil
}

// Plan creates a compilation plan for ANSYS generation.
func (g *Generator) Plan(ctx context.Context, wf *common.Workflow, opts *common.CompileOptions) (*common.CompilePlan, error) {
	est, err := g.EstimateResources(wf)
	if err != nil {
		return nil, err
	}
	req, err := g.RuntimeRequirement(wf)
	if err != nil {
		return nil, err
	}

	return &common.CompilePlan{
		GeneratorID:   string(g.ID()),
		GeneratorName: g.Name(),
		ProjectName:   sanitizeName(wf.Name),
		OutputDir:     path.Join("generated", sanitizeName(wf.Name), "ansys"),
		ResourceEstimate: common.ResourceEstimate{
			EstimatedFiles: est.EstimatedFiles,
			EstimatedSize:  est.EstimatedSize,
		},
		RuntimeRequirement: *req,
		Validated:          true,
		Warnings:           []string{},
	}, nil
}

func (g *Generator) renderTemplate(name string, data interface{}) (string, error) {
	tmplContent, err := templateFS.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("read embedded template %s: %w", name, err)
	}

	tmpl, err := template.New(name).Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", name, err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %s: %w", name, err)
	}

	return buf.String(), nil
}

// nodeTypeToScript maps workflow node types to ANSYS APDL script names.
func nodeTypeToScript(nodeType string) string {
	scripts := map[string]string{
		"ansys.mechanical":  "mechanical_analysis.apdl",
		"ansys.fluent":      "fluent_simulation.jou",
		"ansys.apdl":        "custom_apdl.apdl",
		"ansys.journal":     "custom_journal.wbjn",
		"ansys.material":    "material_def.apdl",
		"ansys.mesh":        "mesh_generation.apdl",
		"ansys.solver":      "solver_config.apdl",
		"simulation.mechanical": "mechanical_analysis.apdl",
		"simulation.fluent":     "fluent_simulation.jou",
		"simulation.apdl":       "custom_apdl.apdl",
		"simulation.journal":    "custom_journal.wbjn",
	}
	if script, ok := scripts[nodeType]; ok {
		return script
	}
	return fmt.Sprintf("%s.apdl", strings.ReplaceAll(nodeType, ".", "_"))
}

// isANSYSNode checks if a node type is ANSYS-related.
func isANSYSNode(nodeType string) bool {
	ansysTypes := []string{
		"ansys.", "simulation.mechanical", "simulation.fluent",
		"simulation.apdl", "simulation.journal",
	}
	for _, prefix := range ansysTypes {
		if strings.HasPrefix(nodeType, prefix) {
			return true
		}
	}
	return false
}

// getNodeDependencies returns node IDs that the given node depends on.
func getNodeDependencies(nodeID string, edges []common.Edge) []string {
	var deps []string
	for _, edge := range edges {
		if edge.Target == nodeID {
			deps = append(deps, edge.Source)
		}
	}
	return deps
}

func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "ansys_project"
	}
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ToLower(name)
	return name
}
