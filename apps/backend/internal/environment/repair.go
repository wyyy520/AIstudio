package environment

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// RepairEnvironment analyzes issues and executes a repair plan.
func (m *Manager) RepairEnvironment() *RepairResult {
	m.log.Info("repair", "Starting environment repair...")

	result := &RepairResult{
		Plan: RepairPlan{
			Steps: []RepairStep{},
		},
		CompletedAt: time.Now(),
	}

	// Step 1: Check environment to find issues
	check := m.CheckEnvironment()
	result.Plan.IssuesFound = len(check.Issues)

	if check.Passed && len(check.Issues) == 0 {
		m.log.Info("repair", "Environment is healthy, no repair needed")
		result.Success = true
		result.Plan.Analysis = "Environment is healthy. No issues found."
		return result
	}

	// Step 2: Analyze issues and generate repair plan
	plan := m.AnalyzeIssues(check.Issues)
	result.Plan = *plan

	m.log.Info("repair", fmt.Sprintf("Analysis: %d issues, %d auto-fixable, %d need manual",
		plan.IssuesFound, plan.AutoFixable, plan.NeedsManual))

	// Step 3: Execute auto-fixable steps
	executedCount := 0
	failedCount := 0
	var logs []string

	for i := range result.Plan.Steps {
		step := &result.Plan.Steps[i]
		step.Status = "running"
		m.log.Info("repair", fmt.Sprintf("Step %d: %s", step.Step, step.Description))

		err := m.executeRepairStep(step)
		if err != nil {
			step.Status = "failed"
			step.Error = err.Error()
			failedCount++
			logs = append(logs, fmt.Sprintf("[FAIL] Step %d: %s - %s",
				step.Step, step.Description, err.Error()))
			m.log.Error("repair",
				fmt.Sprintf("Step %d failed: %s", step.Step, err.Error()), "")
		} else {
			step.Status = "completed"
			executedCount++
			logs = append(logs, fmt.Sprintf("[OK] Step %d: %s",
				step.Step, step.Description))
			m.log.Info("repair", fmt.Sprintf("Step %d completed", step.Step))
		}
	}

	result.StepsExecuted = executedCount
	result.StepsFailed = failedCount
	result.Logs = logs
	result.Success = failedCount == 0
	result.Plan.IssuesFixed = executedCount
	result.CompletedAt = time.Now()

	if result.Success {
		m.log.Info("repair", "Repair completed successfully")
	} else {
		m.log.Warn("repair", fmt.Sprintf("Repair completed with %d failed steps", failedCount))
	}

	return result
}

// AnalyzeIssues takes a list of issues and generates a repair plan.
func (m *Manager) AnalyzeIssues(issues []Issue) *RepairPlan {
	plan := &RepairPlan{
		IssuesFound: len(issues),
		Steps:       []RepairStep{},
	}

	autoFixable := 0
	needsManual := 0
	stepNum := 0

	for _, issue := range issues {
		if issue.AutoFixable {
			autoFixable++
			steps := m.generateFixSteps(issue, &stepNum)
			plan.Steps = append(plan.Steps, steps...)
		} else {
			needsManual++
			stepNum++
			plan.Steps = append(plan.Steps, RepairStep{
				Step:        stepNum,
				Action:      "manual_fix",
				Description: issue.Title + ": " + issue.Suggestion,
				Status:      "pending",
			})
		}
	}

	plan.AutoFixable = autoFixable
	plan.NeedsManual = needsManual

	// Generate analysis summary
	if autoFixable > 0 && needsManual == 0 {
		plan.Analysis = fmt.Sprintf(
			"Found %d issue(s), all can be fixed automatically. Click 'Repair' to fix.",
			len(issues))
	} else if autoFixable > 0 && needsManual > 0 {
		plan.Analysis = fmt.Sprintf(
			"Found %d issue(s): %d can be fixed automatically, %d require manual action.",
			len(issues), autoFixable, needsManual)
	} else {
		plan.Analysis = fmt.Sprintf(
			"Found %d issue(s) that require manual intervention. Please follow the steps below.",
			len(issues))
	}

	return plan
}

// generateFixSteps generates repair steps for a single issue.
func (m *Manager) generateFixSteps(issue Issue, stepNum *int) []RepairStep {
	var steps []RepairStep

	switch {
	case issue.Code == "PIP_NOT_FOUND":
		*stepNum++
		steps = append(steps, RepairStep{
			Step:        *stepNum,
			Action:      "install_pip",
			Description: "Install pip",
			Command:     "python -m ensurepip --upgrade",
			Status:      "pending",
		})

	case strings.HasPrefix(issue.Code, "DEP_MISSING_"):
		pkgName := issue.Component
		*stepNum++
		steps = append(steps, RepairStep{
			Step:        *stepNum,
			Action:      "install_package",
			Description: "Install " + pkgName,
			Command:     "pip install " + pkgName,
			Package:     pkgName,
			Status:      "pending",
		})

	case issue.Code == "CUDA_VERSION_MISMATCH":
		*stepNum++
		cudaVer := strings.ReplaceAll(issue.Description, ".", "")
		steps = append(steps, RepairStep{
			Step:        *stepNum,
			Action:      "reinstall_pytorch",
			Description: "Reinstall PyTorch with CUDA " + issue.Description,
			Command:     "pip install torch --index-url https://download.pytorch.org/whl/cu" + cudaVer,
			Package:     "torch",
			Status:      "pending",
		})

	default:
		*stepNum++
		steps = append(steps, RepairStep{
			Step:        *stepNum,
			Action:      "manual_fix",
			Description: issue.Title + ": " + issue.Suggestion,
			Status:      "pending",
		})
	}

	return steps
}

// executeRepairStep runs a single repair step.
func (m *Manager) executeRepairStep(step *RepairStep) error {
	switch step.Action {
	case "install_package":
		if step.Package == "" {
			return fmt.Errorf("no package specified for install")
		}
		return InstallDependency(step.Package)

	case "install_pip":
		cmd := exec.Command(getPythonCommand(), "-m", "ensurepip", "--upgrade")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("ensurepip failed: %w\n%s", err, string(output))
		}
		return nil

	case "reinstall_pytorch":
		// Uninstall then reinstall
		uninstallCmd := exec.Command(getPipCommand(), "uninstall", "-y", "torch")
		_ = uninstallCmd.Run()
		return InstallDependency("torch")

	default:
		return fmt.Errorf("cannot auto-execute step: %s", step.Action)
	}
}

// GetRepairPlan returns a repair plan without executing it.
// This is useful for previewing what will be fixed before committing.
func (m *Manager) GetRepairPlan() *RepairPlan {
	check := m.CheckEnvironment()
	return m.AnalyzeIssues(check.Issues)
}