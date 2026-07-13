package generators

import (
	"context"

	"github.com/aistudio/backend/internal/compiler"
	"github.com/aistudio/backend/internal/workflow"
	pkgcpp "github.com/aistudio/packages/generators/cpp"
	pkgdocker "github.com/aistudio/packages/generators/docker"
	pkgjava "github.com/aistudio/packages/generators/java"
	pkgmatlab "github.com/aistudio/packages/generators/matlab"
	pkgros2 "github.com/aistudio/packages/generators/ros2"
	pkgstm32 "github.com/aistudio/packages/generators/stm32"
	pkgunity "github.com/aistudio/packages/generators/unity"
)

// ============================================================================
// MATLAB Adapter
// ============================================================================

var _ compiler.Generator = (*MATLABAdapter)(nil)

type MATLABAdapter struct {
	inner *pkgmatlab.Generator
}

func NewMATLABAdapter() *MATLABAdapter {
	return &MATLABAdapter{inner: pkgmatlab.NewGenerator()}
}

func (a *MATLABAdapter) ID() workflow.Target                     { return workflow.TargetMATLAB }
func (a *MATLABAdapter) Name() string                             { return a.inner.Name() }
func (a *MATLABAdapter) Description() string                      { return a.inner.Description() }
func (a *MATLABAdapter) Version() string                          { return a.inner.Version() }
func (a *MATLABAdapter) Validate(wf *workflow.Workflow) error     { return a.inner.Validate(toPkgWorkflow(wf)) }
func (a *MATLABAdapter) CompileTimeValidate(ctx context.Context) error { return a.inner.CompileTimeValidate(ctx) }

func (a *MATLABAdapter) RuntimeRequirement(wf *workflow.Workflow) (*compiler.RuntimeRequirement, error) {
	prr, err := a.inner.RuntimeRequirement(toPkgWorkflow(wf))
	if err != nil {
		return nil, err
	}
	return &compiler.RuntimeRequirement{
		Name: prr.Name, Version: prr.Version, Python: prr.Python,
		Packages: prr.Packages, Commands: prr.Commands, GPU: prr.GPU,
		MinMemoryMB: prr.MinMemoryMB, MinDiskMB: prr.MinDiskMB,
	}, nil
}

func (a *MATLABAdapter) EstimateResources(wf *workflow.Workflow) (*compiler.ResourceEstimate, error) {
	pre, err := a.inner.EstimateResources(toPkgWorkflow(wf))
	if err != nil {
		return nil, err
	}
	return &compiler.ResourceEstimate{
		EstimatedFiles: pre.EstimatedFiles, EstimatedSizeKB: pre.EstimatedSizeKB,
		RequiresGPU: pre.RequiresGPU, MinMemoryMB: pre.MinMemoryMB, MinDiskMB: pre.MinDiskMB,
	}, nil
}

func (a *MATLABAdapter) Generate(ctx context.Context, wf *workflow.Workflow, opts compiler.CompileOptions) (*compiler.GenerateResult, error) {
	pres, err := a.inner.Generate(ctx, toPkgWorkflow(wf), toPkgOptions(opts))
	if err != nil {
		return nil, err
	}
	files := make([]compiler.GeneratedFile, len(pres.Files))
	for i, f := range pres.Files {
		files[i] = compiler.GeneratedFile{Path: f.Path, Content: f.Content, Mode: f.Mode}
	}
	return &compiler.GenerateResult{
		Target: workflow.TargetMATLAB, ProjectRoot: pres.ProjectRoot,
		EntryPoints: pres.EntryPoints, Files: files, ProjectName: pres.ProjectName,
	}, nil
}

// ============================================================================
// ROS2 Adapter
// ============================================================================

var _ compiler.Generator = (*ROS2Adapter)(nil)

type ROS2Adapter struct {
	inner *pkgros2.Generator
}

func NewROS2Adapter() *ROS2Adapter {
	return &ROS2Adapter{inner: pkgros2.NewGenerator()}
}

func (a *ROS2Adapter) ID() workflow.Target                     { return workflow.TargetROS2 }
func (a *ROS2Adapter) Name() string                             { return a.inner.Name() }
func (a *ROS2Adapter) Description() string                      { return a.inner.Description() }
func (a *ROS2Adapter) Version() string                          { return a.inner.Version() }
func (a *ROS2Adapter) Validate(wf *workflow.Workflow) error     { return a.inner.Validate(toPkgWorkflow(wf)) }
func (a *ROS2Adapter) CompileTimeValidate(ctx context.Context) error { return a.inner.CompileTimeValidate(ctx) }

func (a *ROS2Adapter) RuntimeRequirement(wf *workflow.Workflow) (*compiler.RuntimeRequirement, error) {
	prr, err := a.inner.RuntimeRequirement(toPkgWorkflow(wf))
	if err != nil {
		return nil, err
	}
	return &compiler.RuntimeRequirement{
		Name: prr.Name, Version: prr.Version, Python: prr.Python,
		Packages: prr.Packages, Commands: prr.Commands, GPU: prr.GPU,
		MinMemoryMB: prr.MinMemoryMB, MinDiskMB: prr.MinDiskMB,
	}, nil
}

func (a *ROS2Adapter) EstimateResources(wf *workflow.Workflow) (*compiler.ResourceEstimate, error) {
	pre, err := a.inner.EstimateResources(toPkgWorkflow(wf))
	if err != nil {
		return nil, err
	}
	return &compiler.ResourceEstimate{
		EstimatedFiles: pre.EstimatedFiles, EstimatedSizeKB: pre.EstimatedSizeKB,
		RequiresGPU: pre.RequiresGPU, MinMemoryMB: pre.MinMemoryMB, MinDiskMB: pre.MinDiskMB,
	}, nil
}

func (a *ROS2Adapter) Generate(ctx context.Context, wf *workflow.Workflow, opts compiler.CompileOptions) (*compiler.GenerateResult, error) {
	pres, err := a.inner.Generate(ctx, toPkgWorkflow(wf), toPkgOptions(opts))
	if err != nil {
		return nil, err
	}
	files := make([]compiler.GeneratedFile, len(pres.Files))
	for i, f := range pres.Files {
		files[i] = compiler.GeneratedFile{Path: f.Path, Content: f.Content, Mode: f.Mode}
	}
	return &compiler.GenerateResult{
		Target: workflow.TargetROS2, ProjectRoot: pres.ProjectRoot,
		EntryPoints: pres.EntryPoints, Files: files, ProjectName: pres.ProjectName,
	}, nil
}

// ============================================================================
// Docker Adapter
// ============================================================================

var _ compiler.Generator = (*DockerAdapter)(nil)

type DockerAdapter struct {
	inner *pkgdocker.Generator
}

func NewDockerAdapter() *DockerAdapter {
	return &DockerAdapter{inner: pkgdocker.NewGenerator()}
}

func (a *DockerAdapter) ID() workflow.Target                     { return workflow.TargetDocker }
func (a *DockerAdapter) Name() string                             { return a.inner.Name() }
func (a *DockerAdapter) Description() string                      { return a.inner.Description() }
func (a *DockerAdapter) Version() string                          { return a.inner.Version() }
func (a *DockerAdapter) Validate(wf *workflow.Workflow) error     { return a.inner.Validate(toPkgWorkflow(wf)) }
func (a *DockerAdapter) CompileTimeValidate(ctx context.Context) error { return a.inner.CompileTimeValidate(ctx) }

func (a *DockerAdapter) RuntimeRequirement(wf *workflow.Workflow) (*compiler.RuntimeRequirement, error) {
	prr, err := a.inner.RuntimeRequirement(toPkgWorkflow(wf))
	if err != nil {
		return nil, err
	}
	return &compiler.RuntimeRequirement{
		Name: prr.Name, Version: prr.Version, Python: prr.Python,
		Packages: prr.Packages, Commands: prr.Commands, GPU: prr.GPU,
		MinMemoryMB: prr.MinMemoryMB, MinDiskMB: prr.MinDiskMB,
	}, nil
}

func (a *DockerAdapter) EstimateResources(wf *workflow.Workflow) (*compiler.ResourceEstimate, error) {
	pre, err := a.inner.EstimateResources(toPkgWorkflow(wf))
	if err != nil {
		return nil, err
	}
	return &compiler.ResourceEstimate{
		EstimatedFiles: pre.EstimatedFiles, EstimatedSizeKB: pre.EstimatedSizeKB,
		RequiresGPU: pre.RequiresGPU, MinMemoryMB: pre.MinMemoryMB, MinDiskMB: pre.MinDiskMB,
	}, nil
}

func (a *DockerAdapter) Generate(ctx context.Context, wf *workflow.Workflow, opts compiler.CompileOptions) (*compiler.GenerateResult, error) {
	pres, err := a.inner.Generate(ctx, toPkgWorkflow(wf), toPkgOptions(opts))
	if err != nil {
		return nil, err
	}
	files := make([]compiler.GeneratedFile, len(pres.Files))
	for i, f := range pres.Files {
		files[i] = compiler.GeneratedFile{Path: f.Path, Content: f.Content, Mode: f.Mode}
	}
	return &compiler.GenerateResult{
		Target: workflow.TargetDocker, ProjectRoot: pres.ProjectRoot,
		EntryPoints: pres.EntryPoints, Files: files, ProjectName: pres.ProjectName,
	}, nil
}

// ============================================================================
// STM32 Adapter
// ============================================================================

var _ compiler.Generator = (*STM32Adapter)(nil)

type STM32Adapter struct {
	inner *pkgstm32.Generator
}

func NewSTM32Adapter() *STM32Adapter {
	return &STM32Adapter{inner: pkgstm32.NewGenerator()}
}

func (a *STM32Adapter) ID() workflow.Target                     { return workflow.TargetSTM32 }
func (a *STM32Adapter) Name() string                             { return a.inner.Name() }
func (a *STM32Adapter) Description() string                      { return a.inner.Description() }
func (a *STM32Adapter) Version() string                          { return a.inner.Version() }
func (a *STM32Adapter) Validate(wf *workflow.Workflow) error     { return a.inner.Validate(toPkgWorkflow(wf)) }
func (a *STM32Adapter) CompileTimeValidate(ctx context.Context) error { return a.inner.CompileTimeValidate(ctx) }

func (a *STM32Adapter) RuntimeRequirement(wf *workflow.Workflow) (*compiler.RuntimeRequirement, error) {
	prr, err := a.inner.RuntimeRequirement(toPkgWorkflow(wf))
	if err != nil {
		return nil, err
	}
	return &compiler.RuntimeRequirement{
		Name: prr.Name, Version: prr.Version, Python: prr.Python,
		Packages: prr.Packages, Commands: prr.Commands, GPU: prr.GPU,
		MinMemoryMB: prr.MinMemoryMB, MinDiskMB: prr.MinDiskMB,
	}, nil
}

func (a *STM32Adapter) EstimateResources(wf *workflow.Workflow) (*compiler.ResourceEstimate, error) {
	pre, err := a.inner.EstimateResources(toPkgWorkflow(wf))
	if err != nil {
		return nil, err
	}
	return &compiler.ResourceEstimate{
		EstimatedFiles: pre.EstimatedFiles, EstimatedSizeKB: pre.EstimatedSizeKB,
		RequiresGPU: pre.RequiresGPU, MinMemoryMB: pre.MinMemoryMB, MinDiskMB: pre.MinDiskMB,
	}, nil
}

func (a *STM32Adapter) Generate(ctx context.Context, wf *workflow.Workflow, opts compiler.CompileOptions) (*compiler.GenerateResult, error) {
	pres, err := a.inner.Generate(ctx, toPkgWorkflow(wf), toPkgOptions(opts))
	if err != nil {
		return nil, err
	}
	files := make([]compiler.GeneratedFile, len(pres.Files))
	for i, f := range pres.Files {
		files[i] = compiler.GeneratedFile{Path: f.Path, Content: f.Content, Mode: f.Mode}
	}
	return &compiler.GenerateResult{
		Target: workflow.TargetSTM32, ProjectRoot: pres.ProjectRoot,
		EntryPoints: pres.EntryPoints, Files: files, ProjectName: pres.ProjectName,
	}, nil
}

// ============================================================================
// C++ Adapter
// ============================================================================

var _ compiler.Generator = (*CPPAdapter)(nil)

type CPPAdapter struct {
	inner *pkgcpp.Generator
}

func NewCPPAdapter() *CPPAdapter {
	return &CPPAdapter{inner: pkgcpp.NewGenerator()}
}

func (a *CPPAdapter) ID() workflow.Target                     { return workflow.TargetCPP }
func (a *CPPAdapter) Name() string                             { return a.inner.Name() }
func (a *CPPAdapter) Description() string                      { return a.inner.Description() }
func (a *CPPAdapter) Version() string                          { return a.inner.Version() }
func (a *CPPAdapter) Validate(wf *workflow.Workflow) error     { return a.inner.Validate(toPkgWorkflow(wf)) }
func (a *CPPAdapter) CompileTimeValidate(ctx context.Context) error { return a.inner.CompileTimeValidate(ctx) }

func (a *CPPAdapter) RuntimeRequirement(wf *workflow.Workflow) (*compiler.RuntimeRequirement, error) {
	prr, err := a.inner.RuntimeRequirement(toPkgWorkflow(wf))
	if err != nil {
		return nil, err
	}
	return &compiler.RuntimeRequirement{
		Name: prr.Name, Version: prr.Version, Python: prr.Python,
		Packages: prr.Packages, Commands: prr.Commands, GPU: prr.GPU,
		MinMemoryMB: prr.MinMemoryMB, MinDiskMB: prr.MinDiskMB,
	}, nil
}

func (a *CPPAdapter) EstimateResources(wf *workflow.Workflow) (*compiler.ResourceEstimate, error) {
	pre, err := a.inner.EstimateResources(toPkgWorkflow(wf))
	if err != nil {
		return nil, err
	}
	return &compiler.ResourceEstimate{
		EstimatedFiles: pre.EstimatedFiles, EstimatedSizeKB: pre.EstimatedSizeKB,
		RequiresGPU: pre.RequiresGPU, MinMemoryMB: pre.MinMemoryMB, MinDiskMB: pre.MinDiskMB,
	}, nil
}

func (a *CPPAdapter) Generate(ctx context.Context, wf *workflow.Workflow, opts compiler.CompileOptions) (*compiler.GenerateResult, error) {
	pres, err := a.inner.Generate(ctx, toPkgWorkflow(wf), toPkgOptions(opts))
	if err != nil {
		return nil, err
	}
	files := make([]compiler.GeneratedFile, len(pres.Files))
	for i, f := range pres.Files {
		files[i] = compiler.GeneratedFile{Path: f.Path, Content: f.Content, Mode: f.Mode}
	}
	return &compiler.GenerateResult{
		Target: workflow.TargetCPP, ProjectRoot: pres.ProjectRoot,
		EntryPoints: pres.EntryPoints, Files: files, ProjectName: pres.ProjectName,
	}, nil
}

// ============================================================================
// Unity Adapter
// ============================================================================

var _ compiler.Generator = (*UnityAdapter)(nil)

type UnityAdapter struct {
	inner *pkgunity.Generator
}

func NewUnityAdapter() *UnityAdapter {
	return &UnityAdapter{inner: pkgunity.NewGenerator()}
}

func (a *UnityAdapter) ID() workflow.Target                     { return workflow.TargetUnity }
func (a *UnityAdapter) Name() string                             { return a.inner.Name() }
func (a *UnityAdapter) Description() string                      { return a.inner.Description() }
func (a *UnityAdapter) Version() string                          { return a.inner.Version() }
func (a *UnityAdapter) Validate(wf *workflow.Workflow) error     { return a.inner.Validate(toPkgWorkflow(wf)) }
func (a *UnityAdapter) CompileTimeValidate(ctx context.Context) error { return a.inner.CompileTimeValidate(ctx) }

func (a *UnityAdapter) RuntimeRequirement(wf *workflow.Workflow) (*compiler.RuntimeRequirement, error) {
	prr, err := a.inner.RuntimeRequirement(toPkgWorkflow(wf))
	if err != nil {
		return nil, err
	}
	return &compiler.RuntimeRequirement{
		Name: prr.Name, Version: prr.Version, Python: prr.Python,
		Packages: prr.Packages, Commands: prr.Commands, GPU: prr.GPU,
		MinMemoryMB: prr.MinMemoryMB, MinDiskMB: prr.MinDiskMB,
	}, nil
}

func (a *UnityAdapter) EstimateResources(wf *workflow.Workflow) (*compiler.ResourceEstimate, error) {
	pre, err := a.inner.EstimateResources(toPkgWorkflow(wf))
	if err != nil {
		return nil, err
	}
	return &compiler.ResourceEstimate{
		EstimatedFiles: pre.EstimatedFiles, EstimatedSizeKB: pre.EstimatedSizeKB,
		RequiresGPU: pre.RequiresGPU, MinMemoryMB: pre.MinMemoryMB, MinDiskMB: pre.MinDiskMB,
	}, nil
}

func (a *UnityAdapter) Generate(ctx context.Context, wf *workflow.Workflow, opts compiler.CompileOptions) (*compiler.GenerateResult, error) {
	pres, err := a.inner.Generate(ctx, toPkgWorkflow(wf), toPkgOptions(opts))
	if err != nil {
		return nil, err
	}
	files := make([]compiler.GeneratedFile, len(pres.Files))
	for i, f := range pres.Files {
		files[i] = compiler.GeneratedFile{Path: f.Path, Content: f.Content, Mode: f.Mode}
	}
	return &compiler.GenerateResult{
		Target: workflow.TargetUnity, ProjectRoot: pres.ProjectRoot,
		EntryPoints: pres.EntryPoints, Files: files, ProjectName: pres.ProjectName,
	}, nil
}

// ============================================================================
// Java Adapter
// ============================================================================

var _ compiler.Generator = (*JavaAdapter)(nil)

type JavaAdapter struct {
	inner *pkgjava.Generator
}

func NewJavaAdapter() *JavaAdapter {
	return &JavaAdapter{inner: pkgjava.NewGenerator()}
}

func (a *JavaAdapter) ID() workflow.Target                     { return workflow.TargetJava }
func (a *JavaAdapter) Name() string                             { return a.inner.Name() }
func (a *JavaAdapter) Description() string                      { return a.inner.Description() }
func (a *JavaAdapter) Version() string                          { return a.inner.Version() }
func (a *JavaAdapter) Validate(wf *workflow.Workflow) error     { return a.inner.Validate(toPkgWorkflow(wf)) }
func (a *JavaAdapter) CompileTimeValidate(ctx context.Context) error { return a.inner.CompileTimeValidate(ctx) }

func (a *JavaAdapter) RuntimeRequirement(wf *workflow.Workflow) (*compiler.RuntimeRequirement, error) {
	prr, err := a.inner.RuntimeRequirement(toPkgWorkflow(wf))
	if err != nil {
		return nil, err
	}
	return &compiler.RuntimeRequirement{
		Name: prr.Name, Version: prr.Version, Python: prr.Python,
		Packages: prr.Packages, Commands: prr.Commands, GPU: prr.GPU,
		MinMemoryMB: prr.MinMemoryMB, MinDiskMB: prr.MinDiskMB,
	}, nil
}

func (a *JavaAdapter) EstimateResources(wf *workflow.Workflow) (*compiler.ResourceEstimate, error) {
	pre, err := a.inner.EstimateResources(toPkgWorkflow(wf))
	if err != nil {
		return nil, err
	}
	return &compiler.ResourceEstimate{
		EstimatedFiles: pre.EstimatedFiles, EstimatedSizeKB: pre.EstimatedSizeKB,
		RequiresGPU: pre.RequiresGPU, MinMemoryMB: pre.MinMemoryMB, MinDiskMB: pre.MinDiskMB,
	}, nil
}

func (a *JavaAdapter) Generate(ctx context.Context, wf *workflow.Workflow, opts compiler.CompileOptions) (*compiler.GenerateResult, error) {
	pres, err := a.inner.Generate(ctx, toPkgWorkflow(wf), toPkgOptions(opts))
	if err != nil {
		return nil, err
	}
	files := make([]compiler.GeneratedFile, len(pres.Files))
	for i, f := range pres.Files {
		files[i] = compiler.GeneratedFile{Path: f.Path, Content: f.Content, Mode: f.Mode}
	}
	return &compiler.GenerateResult{
		Target: workflow.TargetJava, ProjectRoot: pres.ProjectRoot,
		EntryPoints: pres.EntryPoints, Files: files, ProjectName: pres.ProjectName,
	}, nil
}