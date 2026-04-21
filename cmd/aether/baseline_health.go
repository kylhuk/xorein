package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type checkStatus string

const (
	checkPassed  checkStatus = "passed"
	checkFailed  checkStatus = "failed"
	checkSkipped checkStatus = "skipped"
)

type baselineCheck struct {
	Name    string
	Status  checkStatus
	Command string
	Reason  string
	Output  string
}

type baselineHealthReport struct {
	Root   string
	Checks []baselineCheck
}

var (
	healthCommandOutput   = commandOutput
	defaultHealthLookPath = exec.LookPath
	healthLookPath        = defaultHealthLookPath
)

func runBaselineHealth(w io.Writer, root string) error {
	report, err := collectBaselineHealth(context.Background(), root)
	if err != nil {
		return err
	}
	printBaselineHealth(w, report)
	return nil
}

func collectBaselineHealth(ctx context.Context, root string) (baselineHealthReport, error) {
	if strings.TrimSpace(root) == "" {
		root = "."
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return baselineHealthReport{}, err
	}
	if _, err := os.Stat(absRoot); err != nil {
		return baselineHealthReport{}, err
	}
	report := baselineHealthReport{Root: absRoot}
	if _, err := os.Stat(filepath.Join(absRoot, "go.mod")); err != nil {
		report.Checks = []baselineCheck{
			{Name: "build", Status: checkSkipped, Reason: "no go.mod found"},
			{Name: "test", Status: checkSkipped, Reason: "no go.mod found"},
			{Name: "lint", Status: checkSkipped, Reason: "no lint configuration found"},
			{Name: "typecheck", Status: checkSkipped, Reason: "no go.mod found"},
		}
		return report, nil
	}

	report.Checks = append(report.Checks,
		runGoCheck(ctx, absRoot, "build", "go build ./...", "go", "build", "./..."),
		runGoCheck(ctx, absRoot, "test", "go test ./...", "go", "test", "./..."),
		runLintCheck(ctx, absRoot),
		runGoCheck(ctx, absRoot, "typecheck", "go test -run ^$ ./...", "go", "test", "-run", "^$", "./..."),
	)
	return report, nil
}

func runGoCheck(ctx context.Context, root, name, command, executable string, args ...string) baselineCheck {
	output, err := healthCommandOutput(ctx, root, executable, args...)
	check := baselineCheck{
		Name:    name,
		Command: command,
		Output:  strings.TrimSpace(output),
	}
	if err != nil {
		check.Status = checkFailed
		if check.Output == "" {
			check.Reason = err.Error()
		}
		return check
	}
	check.Status = checkPassed
	return check
}

func runLintCheck(ctx context.Context, root string) baselineCheck {
	if findLintConfig(root) == "" {
		return baselineCheck{
			Name:   "lint",
			Status: checkSkipped,
			Reason: "no lint configuration found",
		}
	}
	if _, err := healthLookPath("golangci-lint"); err != nil {
		return baselineCheck{
			Name:    "lint",
			Status:  checkSkipped,
			Command: "golangci-lint run ./...",
			Reason:  "golangci-lint not installed",
		}
	}
	output, err := healthCommandOutput(ctx, root, "golangci-lint", "run", "./...")
	check := baselineCheck{
		Name:    "lint",
		Command: "golangci-lint run ./...",
		Output:  strings.TrimSpace(output),
	}
	if err != nil {
		check.Status = checkFailed
		if check.Output == "" {
			check.Reason = err.Error()
		}
		return check
	}
	check.Status = checkPassed
	return check
}

func findLintConfig(root string) string {
	for _, name := range []string{".golangci.yml", ".golangci.yaml", ".golangci.toml", ".golangci.json"} {
		path := filepath.Join(root, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func printBaselineHealth(w io.Writer, report baselineHealthReport) {
	_, _ = fmt.Fprintln(w, "Baseline health check")
	_, _ = fmt.Fprintf(w, "root: %s\n", report.Root)
	for _, check := range report.Checks {
		line := fmt.Sprintf("%s: %s", check.Name, check.Status)
		if strings.TrimSpace(check.Command) != "" {
			line += " (" + check.Command + ")"
		}
		if strings.TrimSpace(check.Reason) != "" {
			line += " - " + check.Reason
		}
		_, _ = fmt.Fprintln(w, line)
		if strings.TrimSpace(check.Output) != "" {
			for _, outputLine := range strings.Split(check.Output, "\n") {
				if strings.TrimSpace(outputLine) == "" {
					continue
				}
				_, _ = fmt.Fprintf(w, "  %s\n", outputLine)
			}
		}
	}
}

func commandOutput(ctx context.Context, dir, executable string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("%w: %s %s", err, executable, strings.Join(args, " "))
	}
	return string(output), nil
}
