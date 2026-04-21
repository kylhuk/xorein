package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectRepoSnapshotReportsGitState(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	repoDir := initTestRepo(t)
	writeTestFile(t, repoDir, "tracked.txt", "one\ntwo\n")
	gitRun(t, repoDir, "add", "tracked.txt")
	gitRun(t, repoDir, "commit", "-m", "initial commit")

	writeTestFile(t, repoDir, "second.txt", "committed\n")
	gitRun(t, repoDir, "add", "second.txt")
	gitRun(t, repoDir, "commit", "-m", "second commit")

	writeTestFile(t, repoDir, "tracked.txt", "one\ntwo\nthree\n")
	writeTestFile(t, repoDir, "second.txt", "committed\nchanged\n")
	writeTestFile(t, repoDir, "untracked.txt", "scratch\n")

	snapshot, err := collectRepoSnapshot(repoDir)
	if err != nil {
		t.Fatalf("collectRepoSnapshot() error = %v", err)
	}
	if !snapshot.GitAvailable || !snapshot.RepoDetected {
		t.Fatalf("expected git repo detection, got %+v", snapshot)
	}
	if snapshot.Branch != "main" {
		t.Fatalf("branch = %q want %q", snapshot.Branch, "main")
	}
	if snapshot.CommitHash == "" || snapshot.CommitSummary == "" {
		t.Fatalf("expected latest commit details, got %+v", snapshot)
	}
	if !containsString(snapshot.DirtyFiles, "tracked.txt") || !containsString(snapshot.DirtyFiles, "second.txt") {
		t.Fatalf("dirty files = %#v", snapshot.DirtyFiles)
	}
	if !containsString(snapshot.UntrackedFiles, "untracked.txt") {
		t.Fatalf("untracked files = %#v", snapshot.UntrackedFiles)
	}
	if len(snapshot.RecentCommits) < 2 {
		t.Fatalf("recent commits = %#v", snapshot.RecentCommits)
	}
}

func TestRunRepoSnapshotReportsNonGitDirectory(t *testing.T) {
	var buf bytes.Buffer
	if err := runRepoSnapshot(&buf, t.TempDir()); err != nil {
		t.Fatalf("runRepoSnapshot() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Repository snapshot") {
		t.Fatalf("expected header in output, got %q", output)
	}
	if !strings.Contains(output, "repo: not a git repository") && !strings.Contains(output, "git: unavailable") {
		t.Fatalf("expected graceful non-git report, got %q", output)
	}
}

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init", "-b", "main")
	gitRun(t, dir, "config", "user.name", "Test User")
	gitRun(t, dir, "config", "user.email", "test@example.com")
	return dir
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s error = %v\n%s", strings.Join(args, " "), err, output)
	}
}

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", name, err)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
