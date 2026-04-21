package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type repositorySnapshot struct {
	Root           string
	GitAvailable   bool
	RepoDetected   bool
	Branch         string
	CommitHash     string
	CommitSummary  string
	DirtyFiles     []string
	UntrackedFiles []string
	RecentCommits  []string
}

func runRepoSnapshot(w io.Writer, root string) error {
	snapshot, err := collectRepoSnapshot(root)
	if err != nil {
		return err
	}
	printRepoSnapshot(w, snapshot)
	return nil
}

func collectRepoSnapshot(root string) (repositorySnapshot, error) {
	if strings.TrimSpace(root) == "" {
		root = "."
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return repositorySnapshot{}, err
	}
	if _, err := os.Stat(absRoot); err != nil {
		return repositorySnapshot{}, err
	}
	snapshot := repositorySnapshot{Root: absRoot}
	if _, err := exec.LookPath("git"); err != nil {
		return snapshot, nil
	}
	snapshot.GitAvailable = true
	if _, err := gitOutput(absRoot, "rev-parse", "--show-toplevel"); err != nil {
		return snapshot, nil
	}
	snapshot.RepoDetected = true
	if branch, err := gitOutput(absRoot, "branch", "--show-current"); err == nil {
		snapshot.Branch = strings.TrimSpace(branch)
	}
	if latest, err := gitOutput(absRoot, "log", "-1", "--pretty=format:%H%x09%s"); err == nil {
		parts := strings.SplitN(strings.TrimSpace(latest), "\t", 2)
		if len(parts) > 0 {
			snapshot.CommitHash = parts[0]
		}
		if len(parts) == 2 {
			snapshot.CommitSummary = parts[1]
		}
	}
	if status, err := gitOutput(absRoot, "status", "--porcelain=v1", "--untracked-files=all"); err == nil {
		snapshot.DirtyFiles, snapshot.UntrackedFiles = parsePorcelainStatus(status)
	}
	if history, err := gitOutput(absRoot, "log", "--oneline", "-5"); err == nil {
		for _, line := range strings.Split(strings.TrimSpace(history), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				snapshot.RecentCommits = append(snapshot.RecentCommits, line)
			}
		}
	}
	return snapshot, nil
}

func gitOutput(root string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func parsePorcelainStatus(raw string) (dirty []string, untracked []string) {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || len(line) < 3 {
			continue
		}
		path := strings.TrimSpace(line[3:])
		if strings.HasPrefix(line, "??") {
			untracked = append(untracked, path)
			continue
		}
		dirty = append(dirty, path)
	}
	return dirty, untracked
}

func printRepoSnapshot(w io.Writer, snapshot repositorySnapshot) {
	_, _ = fmt.Fprintln(w, "Repository snapshot")
	_, _ = fmt.Fprintf(w, "root: %s\n", snapshot.Root)
	if !snapshot.GitAvailable {
		_, _ = fmt.Fprintln(w, "git: unavailable")
		return
	}
	_, _ = fmt.Fprintln(w, "git: available")
	if !snapshot.RepoDetected {
		_, _ = fmt.Fprintln(w, "repo: not a git repository")
		return
	}
	_, _ = fmt.Fprintln(w, "repo: detected")
	_, _ = fmt.Fprintf(w, "branch: %s\n", blankIfEmpty(snapshot.Branch, "(detached or unknown)"))
	latest := strings.TrimSpace(snapshot.CommitHash + " " + snapshot.CommitSummary)
	_, _ = fmt.Fprintf(w, "latest commit: %s\n", blankIfEmpty(latest, "(none)"))
	printList(w, "dirty files", snapshot.DirtyFiles, "clean")
	printList(w, "untracked files", snapshot.UntrackedFiles, "none")
	printList(w, "recent commits", snapshot.RecentCommits, "none")
}

func printList(w io.Writer, label string, items []string, empty string) {
	if len(items) == 0 {
		_, _ = fmt.Fprintf(w, "%s: %s\n", label, empty)
		return
	}
	_, _ = fmt.Fprintf(w, "%s:\n", label)
	for _, item := range items {
		_, _ = fmt.Fprintf(w, "  - %s\n", item)
	}
}

func blankIfEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
