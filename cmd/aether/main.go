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

	"github.com/aether/code_aether/pkg/network"
	"github.com/aether/code_aether/pkg/node"
)

var (
	runRuntimeFn        = runRuntime
	runRepoSnapshotFn   = runRepoSnapshot
	runBaselineHealthFn = runBaselineHealth
	explicitCLIFlags    = map[string]bool{}

	runMode        = flag.String("mode", "client", "runtime mode: client|relay|bootstrap|archivist")
	configPath     = flag.String("config", "", "optional JSON config file")
	dataDir        = flag.String("data-dir", filepath.Join(os.TempDir(), "xorein"), "persistent data directory")
	listenAddr     = flag.String("listen", "127.0.0.1:0", "node listen address")
	bootstrapAddrs = flag.String("bootstrap-addrs", "", "comma-separated bootstrap node addresses")
	manualPeers    = flag.String("manual-peers", "", "comma-separated manual peer addresses")
	relayAddrs     = flag.String("relay-addrs", "", "comma-separated relay addresses")
	controlPath    = flag.String("control", "", "control socket path or local endpoint")
	scenario       = flag.String("scenario", "", "legacy flag retained for compatibility; must be empty")
	preflight      = flag.Bool("preflight", false, "print repository snapshot and baseline health check and exit")
	repoSnapshot   = flag.Bool("repo-snapshot", false, "print a repository snapshot and exit")
	baselineHealth = flag.Bool("baseline-health", false, "run baseline build/test/lint/typecheck checks and exit")
	repoRoot       = flag.String("repo-root", ".", "repository root used with --preflight, --repo-snapshot, and --baseline-health")
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
	captureExplicitCLIFlags()
	if *preflight {
		if err := runPreflight(os.Stdout, strings.TrimSpace(*repoRoot)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *repoSnapshot {
		if err := runRepoSnapshot(os.Stdout, strings.TrimSpace(*repoRoot)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *baselineHealth {
		if err := runBaselineHealth(os.Stdout, strings.TrimSpace(*repoRoot)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
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
	if mode := strings.TrimSpace(*runMode); !node.Role(mode).Valid() {
		return node.Config{}, fmt.Errorf("invalid --mode %q; expected client|relay|bootstrap|archivist", mode)
	}
	defaults := defaultNodeConfig()
	cfg := defaults
	if path := strings.TrimSpace(*configPath); path != "" {
		loaded, err := loadFileConfig(path)
		if err != nil {
			return node.Config{}, err
		}
		cfg = mergeConfig(cfg, loaded)
	}
	cfg = mergeConfig(cfg, explicitCLIConfig())
	if !cfg.Role.Valid() {
		return node.Config{}, fmt.Errorf("invalid --mode %q; expected client|relay|bootstrap|archivist", cfg.Role)
	}
	if cfg.DataDir == "" {
		return node.Config{}, fmt.Errorf("data dir is required")
	}
	return cfg, nil
}

func defaultNodeConfig() node.Config {
	return node.Config{
		Role:              node.RoleClient,
		DataDir:           filepath.Join(os.TempDir(), "xorein"),
		ListenAddr:        "127.0.0.1:0",
		DiscoveryInterval: 250 * time.Millisecond,
		HistoryLimit:      32,
	}
}

func explicitCLIConfig() node.Config {
	cfg := node.Config{}
	if flagWasExplicitlySet("mode") {
		cfg.Role = node.Role(strings.TrimSpace(*runMode))
	}
	if flagWasExplicitlySet("data-dir") {
		cfg.DataDir = strings.TrimSpace(*dataDir)
	}
	if flagWasExplicitlySet("listen") {
		cfg.ListenAddr = strings.TrimSpace(*listenAddr)
	}
	if flagWasExplicitlySet("bootstrap-addrs") {
		cfg.BootstrapAddrs = splitCSV(*bootstrapAddrs)
	}
	if flagWasExplicitlySet("manual-peers") {
		cfg.ManualPeers = splitCSV(*manualPeers)
	}
	if flagWasExplicitlySet("relay-addrs") {
		cfg.RelayAddrs = splitCSV(*relayAddrs)
	}
	if flagWasExplicitlySet("control") {
		cfg.ControlEndpoint = strings.TrimSpace(*controlPath)
	}
	return cfg
}

func captureExplicitCLIFlags() {
	explicitCLIFlags = map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		explicitCLIFlags[f.Name] = true
	})
}

func flagWasExplicitlySet(name string) bool {
	return explicitCLIFlags[name]
}

func runRuntime(ctx context.Context, cfg node.Config) error {
	runtime, err := network.NewP2PRuntime(network.Config{Mode: network.Mode(cfg.Role), ListenAddr: cfg.ListenAddr})
	if err != nil {
		return err
	}
	service, err := node.NewService(cfg, node.WithPeerRuntime(runtime))
	if err != nil {
		return err
	}
	runtime.SetHandler(service)
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
	if file.Mode != "" && !cfg.Role.Valid() {
		return node.Config{}, fmt.Errorf("invalid mode in config %q; expected client|relay|bootstrap|archivist", file.Mode)
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
