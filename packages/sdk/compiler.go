package sdk

import (
	"context"

	"github.com/aistudio/packages/compiler"
)

type CompileResult = compiler.CompileResult
type TargetInfo = compiler.TargetInfo
type CompilePlan = compiler.CompilePlan

func Compile(wf *Workflow, target Target, outputDir string) (*CompileResult, error) {
	comp := compiler.NewCompiler(nil)
	opts := compiler.CompileOptions{
		OutputDir: outputDir,
		Target:    target,
	}
	return comp.Compile(context.Background(), wf, opts)
}

func ListTargets() []TargetInfo {
	comp := compiler.NewCompiler(nil)
	return comp.ListTargets()
}

func DryRun(wf *Workflow, target Target) (*CompilePlan, error) {
	comp := compiler.NewCompiler(nil)
	opts := compiler.CompileOptions{
		Target: target,
		DryRun: true,
	}
	return comp.Plan(context.Background(), wf, opts)
}