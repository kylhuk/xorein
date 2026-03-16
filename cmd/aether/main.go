package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/aether/code_aether/pkg/node"
)

var (
	runRuntimeFn = runRuntime

	runMode        = flag.String("mode", "client", "runtime mode: client|relay|bootstrap")
	configPath     = flag.String("config", "", "optional JSON config file")
	dataDir        = flag.String("data-dir", filepath.Join(os.TempDir(), "xorein"), "persistent data directory")
	listenAddr     = flag.String("listen", "127.0.0.1:0", "node listen address")
	bootstrapAddrs = flag.String("bootstrap-addrs", "", "comma-separated bootstrap node addresses")
	manualPeers    = flag.String("manual-peers", "", "comma-separated manual peer addresses")
	relayAddrs     = flag.String("relay-addrs", "", "comma-separated relay addresses")
	controlPath    = flag.String("control", "", "control socket path or local endpoint")
	scenario       = flag.String("scenario", "", "legacy flag retained for compatibility; must be empty")
)

type fileConfig struct {
	Mode              string   `json:"mode"`
	DataDir           string   `json:"data_dir"`
	ListenAddr        string   `json:"listen_addr"`
	BootstrapAddrs    []string `json:"bootstrap_addrs"`
	ManualPeers       []string `json:"manual_peers"`
	RelayAddrs        []string `json:"relay_addrs"`
	ControlEndpoint   string   `json:"control_endpoint"`
	DiscoveryInterval string   `json:"discovery_interval"`
	HistoryLimit      int      `json:"history_limit"`
}

func main() {
	flag.Parse()
	cfg, err := buildNodeConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := runRuntimeFn(ctx, cfg); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildNodeConfig() (node.Config, error) {
	if strings.TrimSpace(*scenario) != "" {
		return node.Config{}, fmt.Errorf("legacy scenario mode is removed from normal execution path")
	}
	cfg := node.Config{
		Role:              node.Role(strings.TrimSpace(*runMode)),
		DataDir:           strings.TrimSpace(*dataDir),
		ListenAddr:        strings.TrimSpace(*listenAddr),
		BootstrapAddrs:    splitCSV(*bootstrapAddrs),
		ManualPeers:       splitCSV(*manualPeers),
		RelayAddrs:        splitCSV(*relayAddrs),
		ControlEndpoint:   strings.TrimSpace(*controlPath),
		DiscoveryInterval: 250 * time.Millisecond,
		HistoryLimit:      32,
	}
	if path := strings.TrimSpace(*configPath); path != "" {
		loaded, err := loadFileConfig(path)
		if err != nil {
			return node.Config{}, err
		}
		cfg = mergeConfig(loaded, cfg)
	}
	if !cfg.Role.Valid() {
		return node.Config{}, fmt.Errorf("invalid --mode %q; expected client|relay|bootstrap", cfg.Role)
	}
	if cfg.DataDir == "" {
		return node.Config{}, fmt.Errorf("data dir is required")
	}
	return cfg, nil
}

func runRuntime(ctx context.Context, cfg node.Config) error {
	service, err := node.NewService(cfg)
	if err != nil {
		return err
	}
	if err := service.Start(ctx); err != nil {
		return err
	}
	defer service.Close()
	snapshot := service.Snapshot()
	fmt.Printf("xorein runtime ready role=%s peer_id=%s listen=%s control=%s\n", snapshot.Role, snapshot.PeerID, first(snapshot.ListenAddresses), snapshot.ControlEndpoint)
	<-ctx.Done()
	return ctx.Err()
}

func loadFileConfig(path string) (node.Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return node.Config{}, err
	}
	var file fileConfig
	if err := json.Unmarshal(raw, &file); err != nil {
		return node.Config{}, err
	}
	cfg := node.Config{
		Role:            node.Role(strings.TrimSpace(file.Mode)),
		DataDir:         strings.TrimSpace(file.DataDir),
		ListenAddr:      strings.TrimSpace(file.ListenAddr),
		BootstrapAddrs:  append([]string(nil), file.BootstrapAddrs...),
		ManualPeers:     append([]string(nil), file.ManualPeers...),
		RelayAddrs:      append([]string(nil), file.RelayAddrs...),
		ControlEndpoint: strings.TrimSpace(file.ControlEndpoint),
		HistoryLimit:    file.HistoryLimit,
	}
	if cfg.HistoryLimit == 0 {
		cfg.HistoryLimit = 32
	}
	if strings.TrimSpace(file.DiscoveryInterval) != "" {
		d, err := time.ParseDuration(file.DiscoveryInterval)
		if err != nil {
			return node.Config{}, err
		}
		cfg.DiscoveryInterval = d
	}
	return cfg, nil
}

func mergeConfig(base, override node.Config) node.Config {
	out := base
	if override.Role.Valid() {
		out.Role = override.Role
	}
	if override.DataDir != "" {
		out.DataDir = override.DataDir
	}
	if override.ListenAddr != "" {
		out.ListenAddr = override.ListenAddr
	}
	if len(override.BootstrapAddrs) > 0 {
		out.BootstrapAddrs = override.BootstrapAddrs
	}
	if len(override.ManualPeers) > 0 {
		out.ManualPeers = override.ManualPeers
	}
	if len(override.RelayAddrs) > 0 {
		out.RelayAddrs = override.RelayAddrs
	}
	if override.ControlEndpoint != "" {
		out.ControlEndpoint = override.ControlEndpoint
	}
	if override.DiscoveryInterval > 0 {
		out.DiscoveryInterval = override.DiscoveryInterval
	}
	if override.HistoryLimit > 0 {
		out.HistoryLimit = override.HistoryLimit
	}
	return out
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
