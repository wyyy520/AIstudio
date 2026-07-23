package generators

import (
	"context"

	"github.com/aistudio/backend/internal/compiler"
	"github.com/aistudio/backend/internal/workflow"
	pkgcommon "github.com/aistudio/packages/generators/common"
	pkgpython "github.com/aistudio/packages/generators/python"
)

// ensure PythonAdapter implements compiler.Generator
var _ compiler.Generator = (*PythonAdapter)(nil)

// PythonAdapter wraps the packages Python generator to implement the
// internal compiler.Generator interface.
type PythonAdapter struct {
	inner *pkgpython.Generator
}

// NewPythonAdapter creates a new adapter for the packages Python generator.
func NewPythonAdapter() *PythonAdapter {
	return &PythonAdapter{
		inner: pkgpython.NewGenerator(),
	}
}

func (a *PythonAdapter) ID() workflow.Target {
	return workflow.TargetPython
}

func (a *PythonAdapter) Name() string {
	return a.inner.Name()
}

func (a *PythonAdapter) Description() string {
	return a.inner.Description()
}

func (a *PythonAdapter) Version() string {
	return a.inner.Version()
}

func (a *PythonAdapter) Validate(wf *workflow.Workflow) error {
	pwf := toPkgWorkflow(wf)
	return a.inner.Validate(pwf)
}

func (a *PythonAdapter) RuntimeRequirement(wf *workflow.Workflow) (*compiler.RuntimeRequirement, error) {
	pwf := toPkgWorkflow(wf)
	prr, err := a.inner.RuntimeRequirement(pwf)
	if err != nil {
		return nil, err
	}
	return &compiler.RuntimeRequirement{
		Name:        prr.Name,
		Version:     prr.Version,
		Python:      prr.Python,
		Packages:    prr.Packages,
		Commands:    prr.Commands,
		GPU:         prr.GPU,
		MinMemoryMB: prr.MinMemoryMB,
		MinDiskMB:   prr.MinDiskMB,
	}, nil
}

func (a *PythonAdapter) EstimateResources(wf *workflow.Workflow) (*compiler.ResourceEstimate, error) {
	pwf := toPkgWorkflow(wf)
	pre, err := a.inner.EstimateResources(pwf)
	if err != nil {
		return nil, err
	}
	return &compiler.ResourceEstimate{
		EstimatedFiles:  pre.EstimatedFiles,
		EstimatedSizeKB: pre.EstimatedSizeKB,
		RequiresGPU:     pre.RequiresGPU,
		MinMemoryMB:     pre.MinMemoryMB,
		MinDiskMB:       pre.MinDiskMB,
	}, nil
}

func (a *PythonAdapter) CompileTimeValidate(ctx context.Context) error {
	return a.inner.CompileTimeValidate(ctx)
}

func (a *PythonAdapter) Generate(ctx context.Context, plan *compiler.ExecutionPlan, opts compiler.CompileOptions) (*compiler.GenerateResult, error) {
	pwf := toPkgWorkflow(plan.Workflow)
	popts := toPkgOptions(opts)
	pres, err := a.inner.Generate(ctx, pwf, popts)
	if err != nil {
		return nil, err
	}

	files := make([]compiler.GeneratedFile, len(pres.Files))
	for i, f := range pres.Files {
		files[i] = compiler.GeneratedFile{
			Path:    f.Path,
			Content: f.Content,
			Mode:    f.Mode,
		}
	}

	return &compiler.GenerateResult{
		Target:      workflow.TargetPython,
		ProjectRoot: pres.ProjectRoot,
		EntryPoints: pres.EntryPoints,
		Files:       files,
		ProjectName: pres.ProjectName,
	}, nil
}

// ============================================================================
// Type Conversion Helpers
// ============================================================================

func toPkgWorkflow(wf *workflow.Workflow) *pkgcommon.Workflow {
	if wf == nil {
		return nil
	}
	nodes := make([]pkgcommon.Node, len(wf.Nodes))
	for i, n := range wf.Nodes {
		nodes[i] = pkgcommon.Node{
			ID:          n.ID,
			Type:        string(n.Type),
			Name:        n.Name,
			Description: n.Description,
			Config:      n.Config,
		}
	}

	edges := make([]pkgcommon.Edge, len(wf.Edges))
	for i, e := range wf.Edges {
		edges[i] = pkgcommon.Edge{
			ID: e.ID,
			Source: pkgcommon.EdgeEndpoint{
				NodeID: e.Source.NodeID,
				PortID: e.Source.PortID,
			},
			Target: pkgcommon.EdgeEndpoint{
				NodeID: e.Target.NodeID,
				PortID: e.Target.PortID,
			},
		}
	}

	// Convert ports for nodes
	for i := range nodes {
		ni := &wf.Nodes[i]
		nodes[i].Inputs = make([]pkgcommon.Port, len(ni.Inputs))
		for j, p := range ni.Inputs {
			nodes[i].Inputs[j] = pkgcommon.Port{
				ID:          p.ID,
				Name:        p.Name,
				Type:        string(p.Type),
				Description: p.Description,
				Required:    p.Required,
			}
		}
		nodes[i].Outputs = make([]pkgcommon.Port, len(ni.Outputs))
		for j, p := range ni.Outputs {
			nodes[i].Outputs[j] = pkgcommon.Port{
				ID:          p.ID,
				Name:        p.Name,
				Type:        string(p.Type),
				Description: p.Description,
				Required:    p.Required,
			}
		}
	}

	return &pkgcommon.Workflow{
		SchemaVersion: wf.SchemaVersion,
		ID:            wf.ID,
		Name:          wf.Name,
		Description:   wf.Description,
		Version:       wf.Version,
		Author:        wf.Author,
		Tags:          wf.Tags,
		Metadata:      wf.Metadata,
		Variables:     wf.Variables,
		Target:        pkgcommon.Target(string(wf.Target)),
		Nodes:         nodes,
		Edges:         edges,
	}
}

func toPkgOptions(opts compiler.CompileOptions) pkgcommon.CompileOptions {
	return pkgcommon.CompileOptions{
		OutputDir:   opts.OutputDir,
		Target:      pkgcommon.Target(string(opts.Target)),
		Variables:   opts.Variables,
		Force:       opts.Force,
		DryRun:      opts.DryRun,
		ProjectName: opts.ProjectName,
	}
}


