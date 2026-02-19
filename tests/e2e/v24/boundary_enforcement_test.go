package v24

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBoundaryEnforcement(t *testing.T) {
	t.Run("passesClean", func(t *testing.T) {
		tmp := t.TempDir()
		writeGoFile(t, tmp, "cmd/xorein/main.go", `package main

func main() {}
`)
		writeGoFile(t, tmp, "cmd/harmolyn/main.go", `package main

func main() {}
`)

		out, err := runBoundaryScript(t, tmp)
		if err != nil {
			t.Fatalf("expected success but script failed: %v\n%s", err, out)
		}
	})

	t.Run("rejectsGioImports", func(t *testing.T) {
		tmp := t.TempDir()
		writeGoFile(t, tmp, "cmd/xorein/uses_gio.go", `package main

import "gioui.org/app"

func unused() {}
`)
		writeGoFile(t, tmp, "cmd/harmolyn/main.go", `package main

func main() {}
`)

		out, err := runBoundaryScript(t, tmp)
		if err == nil {
			t.Fatalf("expected Gio violation to fail but script succeeded")
		}
		if !strings.Contains(out, "ST1 (Gio import)") {
			t.Fatalf("missing Gio failure detail in output:\n%s", out)
		}
	})

	t.Run("rejectsProtocolImports", func(t *testing.T) {
		tmp := t.TempDir()
		writeGoFile(t, tmp, "cmd/xorein/main.go", `package main

func main() {}
`)
		writeGoFile(t, tmp, "cmd/harmolyn/uses_runtime.go", `package main

import "github.com/aether/code_aether/pkg/v23/security"

func main() {}
`)

		out, err := runBoundaryScript(t, tmp)
		if err == nil {
			t.Fatalf("expected protocol runtime violation but script succeeded")
		}
		if !strings.Contains(out, "ST2 (protocol runtime import)") {
			t.Fatalf("missing protocol failure detail in output:\n%s", out)
		}
	})
}

func runBoundaryScript(t *testing.T, root string) (string, error) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	script := filepath.Clean(filepath.Join(wd, "..", "..", "..", "scripts", "ci", "enforce-boundaries.sh"))
	cmd := exec.Command("bash", script)
	cmd.Env = append(os.Environ(), fmt.Sprintf("CHECK_ROOT=%s", root))
	out, runErr := cmd.CombinedOutput()
	return string(out), runErr
}

func writeGoFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
