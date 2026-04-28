package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	v0_1 "github.com/aether/code_aether/pkg/v0_1"
)

var (
	listenAddr     = flag.String("listen", "127.0.0.1:0", "node listen address")
	dataDir        = flag.String("data-dir", "", "data directory (enables persistent state and control API)")
	controlAddr    = flag.String("control", "", "control API socket path (default: <data-dir>/xorein-control.sock)")
	roleFlag       = flag.String("role", "client", "node role: client, relay, bootstrap, or archivist")
	bootstrapAddrs = flag.String("bootstrap-addrs", "", "comma-separated bootstrap node multiaddrs (/ip4/.../tcp/.../p2p/<id>)")
	manualPeers    = flag.String("manual-peers", "", "comma-separated manual peer multiaddrs (never expire)")
	relayAddrs     = flag.String("relay-addrs", "", "comma-separated circuit-relay v2 server multiaddrs (client-side)")
	enableMDNS     = flag.Bool("enable-mdns", true, "enable LAN peer discovery via mDNS")
	enableNAT      = flag.Bool("enable-nat", true, "enable NAT port mapping and hole punching")
	preflight      = flag.Bool("preflight", false, "print repository snapshot and baseline health check and exit")
	repoSnapshot   = flag.Bool("repo-snapshot", false, "print a repository snapshot and exit")
	baselineHealth = flag.Bool("baseline-health", false, "run baseline build/test/lint/typecheck checks and exit")
	repoRoot       = flag.String("repo-root", ".", "repository root used with --preflight, --repo-snapshot, and --baseline-health")
)

func main() {
	// "aether serve [flags]" is an alias for running the default runtime.
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}
	flag.Parse()
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := runV01Runtime(ctx); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runV01Runtime(ctx context.Context) error {
	listen := strings.TrimSpace(*listenAddr)
	if listen == "" {
		listen = "127.0.0.1:0"
	}
	maListen := tcpAddrToMultiaddr(listen)

	rt, err := v0_1.Start(ctx, v0_1.Config{
		ListenAddr:     maListen,
		Role:           strings.TrimSpace(*roleFlag),
		DataDir:        strings.TrimSpace(*dataDir),
		ControlAddr:    strings.TrimSpace(*controlAddr),
		BootstrapAddrs: splitCSV(*bootstrapAddrs),
		ManualPeers:    splitCSV(*manualPeers),
		RelayAddrs:     splitCSV(*relayAddrs),
		EnableMDNS:     *enableMDNS,
		EnableNAT:      *enableNAT,
	})
	if err != nil {
		return fmt.Errorf("v0.1 runtime: %w", err)
	}
	defer rt.Close()

	addrs := rt.ListenAddrs()
	listenStr := ""
	if len(addrs) > 0 {
		listenStr = addrs[0]
	}
	ctrlStr := rt.ControlAddr()
	if ctrlStr != "" {
		fmt.Printf("xorein runtime ready peer_id=%s listen=%s control=%s\n", rt.PeerID(), listenStr, ctrlStr)
	} else {
		fmt.Printf("xorein runtime ready peer_id=%s listen=%s\n", rt.PeerID(), listenStr)
	}
	<-ctx.Done()
	return ctx.Err()
}

// splitCSV splits a comma-separated string into trimmed non-empty elements.
func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// tcpAddrToMultiaddr converts "host:port" to "/ip4/host/tcp/port".
func tcpAddrToMultiaddr(addr string) string {
	if strings.HasPrefix(addr, "/") {
		return addr
	}
	idx := strings.LastIndex(addr, ":")
	if idx < 0 {
		return "/ip4/" + addr + "/tcp/0"
	}
	host := addr[:idx]
	port := addr[idx+1:]
	if host == "" {
		host = "0.0.0.0"
	}
	return "/ip4/" + host + "/tcp/" + port
}
