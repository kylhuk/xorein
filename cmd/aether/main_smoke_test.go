package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestAetherRuntimeSmokeHarness(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not available")
	}
	packageDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	binaryPath := buildAetherBinary(t, packageDir)
	proc := startAetherProcess(t, binaryPath)
	defer proc.stop(t)

	if proc.listenAddr == "" {
		t.Fatal("readiness line did not include listen address")
	}
	t.Logf("v0.1 runtime ready at listen=%s", proc.listenAddr)
}

func TestAetherServeSubcommandAlias(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not available")
	}
	packageDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	binaryPath := buildAetherBinary(t, packageDir)
	proc := startAetherProcessWithArgs(t, binaryPath, "serve", "--listen", "127.0.0.1:0")
	defer proc.stop(t)
	if proc.listenAddr == "" {
		t.Fatal("serve alias: listenAddr not captured from readiness line")
	}
}

type startedAetherProcess struct {
	cmd        *exec.Cmd
	cancel     context.CancelFunc
	listenAddr string

	mu     sync.Mutex
	buffer bytes.Buffer
	waitCh chan error
}

func buildAetherBinary(t *testing.T, packageDir string) string {
	t.Helper()
	binaryPath := filepath.Join(t.TempDir(), "aether-smoke")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = packageDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build cmd/aether error = %v\n%s", err, output)
	}
	return binaryPath
}

func startAetherProcess(t *testing.T, binaryPath string) *startedAetherProcess {
	t.Helper()
	return startAetherProcessWithArgs(t, binaryPath, "--listen", "127.0.0.1:0")
}

func startAetherProcessWithArgs(t *testing.T, binaryPath string, args ...string) *startedAetherProcess {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, binaryPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatalf("StdoutPipe() error = %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		t.Fatalf("StderrPipe() error = %v", err)
	}
	proc := &startedAetherProcess{cmd: cmd, cancel: cancel, waitCh: make(chan error, 1)}
	readyCh := make(chan string, 1)

	proc.captureOutput(stdout, readyCh)
	proc.captureOutput(stderr, nil)

	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("Start() error = %v", err)
	}
	go func() {
		proc.waitCh <- cmd.Wait()
	}()

	select {
	case listenAddr := <-readyCh:
		proc.listenAddr = listenAddr
		return proc
	case err := <-proc.waitCh:
		cancel()
		if err == nil {
			t.Fatalf("aether process exited before readiness\nprocess output:\n%s", proc.output())
		}
		t.Fatalf("aether process exited before readiness: %v\nprocess output:\n%s", err, proc.output())
	case <-time.After(10 * time.Second):
		proc.stop(t)
		t.Fatalf("timed out waiting for readiness line\nprocess output:\n%s", proc.output())
		return nil
	}
	return nil
}

func (p *startedAetherProcess) captureOutput(r io.ReadCloser, readyCh chan<- string) {
	go func() {
		defer r.Close()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			p.appendOutput(line)
			if readyCh != nil {
				if listenAddr, ok := readinessListenAddress(line); ok {
					select {
					case readyCh <- listenAddr:
					default:
					}
				}
			}
		}
		if err := scanner.Err(); err != nil {
			p.appendOutput(fmt.Sprintf("scanner error: %v", err))
		}
	}()
}

func (p *startedAetherProcess) appendOutput(line string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.buffer.WriteString(line)
	p.buffer.WriteByte('\n')
}

func (p *startedAetherProcess) output() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.buffer.String()
}

func (p *startedAetherProcess) stop(t *testing.T) {
	t.Helper()
	if p == nil || p.cmd == nil || p.cmd.Process == nil {
		return
	}
	_ = p.cmd.Process.Signal(syscall.SIGINT)
	select {
	case err := <-p.waitCh:
		p.cancel()
		if err != nil && !isExpectedShutdown(err) {
			t.Logf("aether shutdown: %v\nprocess output:\n%s", err, p.output())
		}
	case <-time.After(5 * time.Second):
		p.cancel()
		select {
		case err := <-p.waitCh:
			if err != nil && !isExpectedShutdown(err) {
				t.Logf("aether forced shutdown: %v", err)
			}
		case <-time.After(2 * time.Second):
			_ = p.cmd.Process.Kill()
			<-p.waitCh
		}
	}
}

func isExpectedShutdown(err error) bool {
	if err == nil {
		return true
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return false
	}
	return errors.Is(err, context.Canceled)
}

func readinessListenAddress(line string) (string, bool) {
	if !strings.Contains(line, "xorein runtime ready") {
		return "", false
	}
	for _, field := range strings.Fields(line) {
		if strings.HasPrefix(field, "listen=") {
			listenAddr := strings.TrimSpace(strings.TrimPrefix(field, "listen="))
			if listenAddr != "" {
				return listenAddr, true
			}
		}
	}
	return "", false
}
