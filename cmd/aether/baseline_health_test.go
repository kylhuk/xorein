package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectBaselineHealthForGoRepoWithoutLintConfig(t *testing.T) {
	root := t.TempDir()
	writeBaselineHealthFile(t, filepath.Join(root, "go.mod"), "module example.com/test\n\ngo 1.22\n")

	commands := make([]string, 0, 3)
	healthCommandOutput = func(ctx context.Context, dir, executable string, args ...string) (string, error) {
		commands = append(commands, executable+" "+strings.Join(args, " "))
		return "ok", nil
	}
	healthLookPath = func(file string) (string, error) {
		if file == "golangci-lint" {
			return "", errors.New("not found")
		}
		return "/usr/bin/" + file, nil
	}
	t.Cleanup(resetHealthHooks)

	report, err := collectBaselineHealth(context.Background(), root)
	if err != nil {
		t.Fatalf("collectBaselineHealth() error = %v", err)
	}
	if len(report.Checks) != 4 {
		t.Fatalf("checks = %#v", report.Checks)
	}
	if report.Checks[0].Name != "build" || report.Checks[0].Status != checkPassed {
		t.Fatalf("build check = %#v", report.Checks[0])
	}
	if report.Checks[1].Name != "test" || report.Checks[1].Status != checkPassed {
		t.Fatalf("test check = %#v", report.Checks[1])
	}
	if report.Checks[2].Name != "lint" || report.Checks[2].Status != checkSkipped || !strings.Contains(report.Checks[2].Reason, "no lint configuration") {
		t.Fatalf("lint check = %#v", report.Checks[2])
	}
	if report.Checks[3].Name != "typecheck" || report.Checks[3].Status != checkPassed {
		t.Fatalf("typecheck check = %#v", report.Checks[3])
	}
	if len(commands) != 3 {
		t.Fatalf("commands = %#v", commands)
	}
	if commands[0] != "go build ./..." {
		t.Fatalf("build command = %q", commands[0])
	}
	if commands[1] != "go test ./..." {
		t.Fatalf("test command = %q", commands[1])
	}
	if commands[2] != "go test -run ^$ ./..." {
		t.Fatalf("typecheck command = %q", commands[2])
	}
}

func TestCollectBaselineHealthWithLintConfigSkipsWhenToolMissing(t *testing.T) {
	root := t.TempDir()
	writeBaselineHealthFile(t, filepath.Join(root, "go.mod"), "module example.com/test\n\ngo 1.22\n")
	writeBaselineHealthFile(t, filepath.Join(root, ".golangci.yml"), "version: 2\n")

	healthCommandOutput = func(ctx context.Context, dir, executable string, args ...string) (string, error) {
		return "", nil
	}
	healthLookPath = func(file string) (string, error) {
		if file == "golangci-lint" {
			return "", errors.New("not found")
		}
		return "/usr/bin/" + file, nil
	}
	t.Cleanup(resetHealthHooks)

	report, err := collectBaselineHealth(context.Background(), root)
	if err != nil {
		t.Fatalf("collectBaselineHealth() error = %v", err)
	}
	if report.Checks[2].Name != "lint" {
		t.Fatalf("lint slot = %#v", report.Checks[2])
	}
	if report.Checks[2].Status != checkSkipped {
		t.Fatalf("lint status = %#v", report.Checks[2])
	}
	if !strings.Contains(report.Checks[2].Reason, "not installed") {
		t.Fatalf("lint reason = %#v", report.Checks[2])
	}
}

func TestRunBaselineHealthPrintsHumanReadableOutput(t *testing.T) {
	root := t.TempDir()
	writeBaselineHealthFile(t, filepath.Join(root, "go.mod"), "module example.com/test\n\ngo 1.22\n")

	healthCommandOutput = func(ctx context.Context, dir, executable string, args ...string) (string, error) {
		return "all good", nil
	}
	healthLookPath = func(file string) (string, error) {
		return "", errors.New("not found")
	}
	t.Cleanup(resetHealthHooks)

	var buf bytes.Buffer
	if err := runBaselineHealth(&buf, root); err != nil {
		t.Fatalf("runBaselineHealth() error = %v", err)
	}
	output := buf.String()
	for _, needle := range []string{
		"Baseline health check",
		"build: passed (go build ./...)",
		"test: passed (go test ./...)",
		"lint: skipped - no lint configuration found",
		"typecheck: passed (go test -run ^$ ./...)",
	} {
		if !strings.Contains(output, needle) {
			t.Fatalf("expected %q in output %q", needle, output)
		}
	}
}

func resetHealthHooks() {
	healthCommandOutput = commandOutput
	healthLookPath = defaultHealthLookPath
}

func writeBaselineHealthFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
