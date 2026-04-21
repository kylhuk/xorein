package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/node"
)

func TestBuildNodeConfigHonoursFlags(t *testing.T) {
	resetFlags(t)
	setFlag(t, "mode", "bootstrap")
	setFlag(t, "data-dir", filepath.Join(os.TempDir(), "xorein"))
	setFlag(t, "listen", "127.0.0.1:9001")
	setFlag(t, "bootstrap-addrs", "127.0.0.1:9000,127.0.0.1:9002")
	setFlag(t, "manual-peers", "127.0.0.1:9003")
	setFlag(t, "relay-addrs", "127.0.0.1:9004")
	setFlag(t, "control", filepath.Join(t.TempDir(), "control.sock"))

	cfg, err := buildNodeConfig()
	if err != nil {
		t.Fatalf("buildNodeConfig() error = %v", err)
	}
	if cfg.Role != node.RoleBootstrap {
		t.Fatalf("role = %s", cfg.Role)
	}
	if len(cfg.BootstrapAddrs) != 2 || len(cfg.ManualPeers) != 1 || len(cfg.RelayAddrs) != 1 {
		t.Fatalf("unexpected addresses: %+v", cfg)
	}
}

func TestBuildNodeConfigHonoursConfigFileDefaults(t *testing.T) {
	resetFlags(t)
	dataDir := t.TempDir()
	controlPath := filepath.Join(t.TempDir(), "cfg.sock")
	file := fileConfig{
		Mode:              "client",
		DataDir:           dataDir,
		ListenAddr:        "127.0.0.1:9100",
		BootstrapAddrs:    []string{"127.0.0.1:9999"},
		ManualPeers:       []string{"127.0.0.1:9998"},
		RelayAddrs:        []string{"127.0.0.1:9997"},
		ControlEndpoint:   controlPath,
		DiscoveryInterval: "500ms",
		HistoryLimit:      7,
	}
	raw, _ := json.Marshal(file)
	configFile := filepath.Join(t.TempDir(), "xorein.json")
	if err := os.WriteFile(configFile, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	*configPath = configFile

	cfg, err := buildNodeConfig()
	if err != nil {
		t.Fatalf("buildNodeConfig() error = %v", err)
	}
	if cfg.Role != node.RoleClient {
		t.Fatalf("role = %s", cfg.Role)
	}
	if cfg.DataDir != dataDir {
		t.Fatalf("data dir = %s want %s", cfg.DataDir, dataDir)
	}
	if cfg.ListenAddr != file.ListenAddr {
		t.Fatalf("listen addr = %s want %s", cfg.ListenAddr, file.ListenAddr)
	}
	if len(cfg.BootstrapAddrs) != 1 || cfg.BootstrapAddrs[0] != file.BootstrapAddrs[0] {
		t.Fatalf("bootstrap addrs = %v", cfg.BootstrapAddrs)
	}
	if len(cfg.ManualPeers) != 1 || cfg.ManualPeers[0] != file.ManualPeers[0] {
		t.Fatalf("manual peers = %v", cfg.ManualPeers)
	}
	if len(cfg.RelayAddrs) != 1 || cfg.RelayAddrs[0] != file.RelayAddrs[0] {
		t.Fatalf("relay addrs = %v", cfg.RelayAddrs)
	}
	if cfg.ControlEndpoint != controlPath {
		t.Fatalf("control endpoint = %s want %s", cfg.ControlEndpoint, controlPath)
	}
	if cfg.DiscoveryInterval != 500*time.Millisecond {
		t.Fatalf("discovery interval = %s want 500ms", cfg.DiscoveryInterval)
	}
	if cfg.HistoryLimit != 7 {
		t.Fatalf("history limit = %d want 7", cfg.HistoryLimit)
	}
}

func TestBuildNodeConfigAcceptsArchivistMode(t *testing.T) {
	resetFlags(t)
	setFlag(t, "mode", "archivist")
	setFlag(t, "data-dir", t.TempDir())

	cfg, err := buildNodeConfig()
	if err != nil {
		t.Fatalf("buildNodeConfig() error = %v", err)
	}
	if cfg.Role != node.RoleArchivist {
		t.Fatalf("role = %s", cfg.Role)
	}
}

func TestBuildNodeConfigRejectsInvalidMode(t *testing.T) {
	resetFlags(t)
	setFlag(t, "mode", "definitely-not-a-mode")
	setFlag(t, "data-dir", t.TempDir())
	if _, err := buildNodeConfig(); err == nil {
		t.Fatal("expected invalid mode error")
	}
}

func TestBuildNodeConfigRejectsInvalidConfigFileMode(t *testing.T) {
	resetFlags(t)
	file := fileConfig{Mode: "definitely-not-a-mode", DataDir: t.TempDir()}
	raw, _ := json.Marshal(file)
	configFile := filepath.Join(t.TempDir(), "xorein.json")
	if err := os.WriteFile(configFile, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	setFlag(t, "config", configFile)
	if _, err := buildNodeConfig(); err == nil {
		t.Fatal("expected invalid config file mode error")
	}
}

func TestBuildNodeConfigRejectsLegacyScenario(t *testing.T) {
	resetFlags(t)
	setFlag(t, "scenario", "first-contact")
	if _, err := buildNodeConfig(); err == nil {
		t.Fatal("expected legacy scenario error")
	}
}

func TestBuildNodeConfigAllowsExplicitDefaultOverride(t *testing.T) {
	resetFlags(t)
	file := fileConfig{Mode: "relay", DataDir: t.TempDir()}
	raw, _ := json.Marshal(file)
	configFile := filepath.Join(t.TempDir(), "xorein.json")
	if err := os.WriteFile(configFile, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	setFlag(t, "config", configFile)
	setFlag(t, "mode", "client")

	cfg, err := buildNodeConfig()
	if err != nil {
		t.Fatalf("buildNodeConfig() error = %v", err)
	}
	if cfg.Role != node.RoleClient {
		t.Fatalf("role = %s want %s", cfg.Role, node.RoleClient)
	}
}

func TestRunRuntimeStartsService(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := node.Config{Role: node.RoleClient, DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0", DiscoveryInterval: 50 * time.Millisecond}
	errCh := make(chan error, 1)
	go func() { errCh <- runRuntime(ctx, cfg) }()
	time.Sleep(150 * time.Millisecond)
	cancel()
	if err := <-errCh; err == nil {
		t.Fatal("expected context cancellation")
	}
}

func TestRunRuntimeRequiresValidNetworkRuntimeMode(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := runRuntime(ctx, node.Config{Role: node.Role("invalid"), DataDir: t.TempDir(), ListenAddr: "127.0.0.1:0"})
	if err == nil {
		t.Fatal("expected invalid runtime mode error")
	}
}

func resetFlags(t *testing.T) {
	t.Helper()
	explicitCLIFlags = map[string]bool{}
	*runMode = "client"
	*configPath = ""
	*dataDir = filepath.Join(os.TempDir(), "xorein")
	*listenAddr = "127.0.0.1:0"
	*bootstrapAddrs = ""
	*manualPeers = ""
	*relayAddrs = ""
	*controlPath = ""
	*scenario = ""
	*preflight = false
	*repoSnapshot = false
	*baselineHealth = false
	*repoRoot = "."
}

func setFlag(t *testing.T, name, value string) {
	t.Helper()
	if err := flag.CommandLine.Set(name, value); err != nil {
		t.Fatalf("Set(%s) error = %v", name, err)
	}
	explicitCLIFlags[name] = true
}
