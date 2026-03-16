package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/node"
)

func TestBuildNodeConfigHonoursFlags(t *testing.T) {
	resetFlags(t)
	*runMode = "bootstrap"
	*dataDir = t.TempDir()
	*listenAddr = "127.0.0.1:9001"
	*bootstrapAddrs = "127.0.0.1:9000,127.0.0.1:9002"
	*manualPeers = "127.0.0.1:9003"
	*relayAddrs = "127.0.0.1:9004"
	*controlPath = filepath.Join(t.TempDir(), "control.sock")

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

func TestBuildNodeConfigMergesFileConfig(t *testing.T) {
	resetFlags(t)
	file := fileConfig{
		Mode:              "client",
		DataDir:           t.TempDir(),
		ListenAddr:        "127.0.0.1:0",
		BootstrapAddrs:    []string{"127.0.0.1:9999"},
		ControlEndpoint:   filepath.Join(t.TempDir(), "cfg.sock"),
		DiscoveryInterval: "500ms",
		HistoryLimit:      7,
	}
	raw, _ := json.Marshal(file)
	configFile := filepath.Join(t.TempDir(), "xorein.json")
	if err := os.WriteFile(configFile, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	*configPath = configFile
	*runMode = "relay"

	cfg, err := buildNodeConfig()
	if err != nil {
		t.Fatalf("buildNodeConfig() error = %v", err)
	}
	if cfg.Role != node.RoleRelay {
		t.Fatalf("override role failed: %s", cfg.Role)
	}
	if cfg.HistoryLimit != 32 {
		t.Fatalf("expected flag default history limit override, got %d", cfg.HistoryLimit)
	}
	if len(cfg.BootstrapAddrs) != 1 || cfg.BootstrapAddrs[0] != "127.0.0.1:9999" {
		t.Fatalf("bootstrap addrs = %v", cfg.BootstrapAddrs)
	}
}

func TestBuildNodeConfigRejectsLegacyScenario(t *testing.T) {
	resetFlags(t)
	*scenario = "first-contact"
	if _, err := buildNodeConfig(); err == nil {
		t.Fatal("expected legacy scenario error")
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

func resetFlags(t *testing.T) {
	t.Helper()
	*runMode = "client"
	*configPath = ""
	*dataDir = t.TempDir()
	*listenAddr = "127.0.0.1:0"
	*bootstrapAddrs = ""
	*manualPeers = ""
	*relayAddrs = ""
	*controlPath = ""
	*scenario = ""
}
