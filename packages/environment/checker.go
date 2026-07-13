package environment

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func CheckRequirements(ctx context.Context, req *Requirement) *EnvironmentReport {
	report := DetectAll(ctx)

	if req == nil {
		return report
	}

	if req.Python != "" && report.PythonVersion != "" {
		if !VersionSatisfies(report.PythonVersion, req.Python) {
			report.Issues = append(report.Issues, &Issue{
				Severity:    SeverityError,
				Component:   "python",
				Code:        "PYTHON_VERSION_MISMATCH",
				Title:       "Python version mismatch",
				Description: fmt.Sprintf("Python %s required, found %s", req.Python, report.PythonVersion),
				Suggestion:  "Install Python " + req.Python,
			})
		}
	}

	if req.GPU && !report.GPUAvailable {
		report.Issues = append(report.Issues, &Issue{
			Severity:    SeverityError,
			Component:   "cuda",
			Code:        "GPU_REQUIRED",
			Title:       "GPU required but not available",
			Description: "The runtime requires a GPU but none was detected",
			Suggestion:  "Install NVIDIA driver and CUDA toolkit",
		})
	}

	for _, cmd := range req.Commands {
		if _, found := report.Commands[cmd]; !found {
			available := checkCommandAvailable(cmd)
			report.Commands[cmd] = available
			if !available {
				report.Issues = append(report.Issues, &Issue{
					Severity:    SeverityError,
					Component:   cmd,
					Code:        "COMMAND_NOT_FOUND_" + strings.ToUpper(cmd),
					Title:       "Required command not found: " + cmd,
					Description: fmt.Sprintf("The command %q is required but not available on this system", cmd),
					Suggestion:  "Install " + cmd + " using your package manager",
				})
			}
		}
	}

	for _, pkg := range req.Packages {
		pkgName := ExtractPackageName(pkg)
		if _, found := report.Packages[pkgName]; !found {
			installed, version := checkPipPackage(pkgName)
			if installed {
				report.Packages[pkgName] = version
			} else {
				report.Issues = append(report.Issues, &Issue{
					Severity:    SeverityError,
					Component:   pkgName,
					Code:        "PACKAGE_NOT_FOUND_" + strings.ToUpper(pkgName),
					Title:       "Required package not found: " + pkgName,
					Description: fmt.Sprintf("The Python package %q is required but not installed", pkgName),
					Suggestion:  "pip install " + pkg,
					AutoFixable: true,
				})
			}
		}
	}

	report.Ready = len(report.Issues) == 0
	return report
}

func ParseVersionConstraint(constraint string) (op string, version string) {
	constraint = strings.TrimSpace(constraint)
	for _, possibleOp := range []string{">=", "<=", "==", "!=", ">", "<", "~="} {
		if strings.HasPrefix(constraint, possibleOp) {
			return possibleOp, strings.TrimSpace(strings.TrimPrefix(constraint, possibleOp))
		}
	}
	return "==", constraint
}

func VersionSatisfies(version, constraint string) bool {
	version = strings.TrimSpace(version)
	constraint = strings.TrimSpace(constraint)

	if constraint == "" {
		return true
	}

	op, required := ParseVersionConstraint(constraint)

	if op == "" || op == "==" {
		return compareVersions(version, required) == 0
	}

	cmp := compareVersions(version, required)
	switch op {
	case ">=":
		return cmp >= 0
	case "<=":
		return cmp <= 0
	case "!=":
		return cmp != 0
	case ">":
		return cmp > 0
	case "<":
		return cmp < 0
	case "~=":
		vParts := strings.Split(version, ".")
		rParts := strings.Split(required, ".")
		if len(vParts) >= 2 && len(rParts) >= 2 {
			return vParts[0] == rParts[0] && compareVersions(vParts[1], rParts[1]) >= 0
		}
		return cmp >= 0
	}
	return true
}

func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var aInt, bInt int
		if i < len(aParts) {
			fmt.Sscanf(aParts[i], "%d", &aInt)
		}
		if i < len(bParts) {
			fmt.Sscanf(bParts[i], "%d", &bInt)
		}
		if aInt < bInt {
			return -1
		}
		if aInt > bInt {
			return 1
		}
	}
	return 0
}

func ExtractPackageName(pkgSpec string) string {
	for _, sep := range []string{">=", "<=", "==", "!=", ">", "<", "~="} {
		if idx := strings.Index(pkgSpec, sep); idx >= 0 {
			return pkgSpec[:idx]
		}
	}
	return pkgSpec
}

func checkCommandAvailable(name string) bool {
	checkCmd := "which"
	if runtime.GOOS == "windows" {
		checkCmd = "where"
	}
	return exec.Command(checkCmd, name).Run() == nil
}

func checkPipPackage(name string) (bool, string) {
	pipCmd := "pip3"
	if runtime.GOOS == "windows" {
		pipCmd = "pip"
	}

	cmd := exec.Command(pipCmd, "show", name)
	output, err := cmd.Output()
	if err != nil {
		return false, ""
	}

	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "Version:") {
			return true, strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
		}
	}
	return true, ""
}
