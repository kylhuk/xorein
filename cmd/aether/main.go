package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	phase11 "github.com/aether/code_aether/pkg/phase11"
	phase6 "github.com/aether/code_aether/pkg/phase6"
	phase9 "github.com/aether/code_aether/pkg/phase9"
	v08scenario "github.com/aether/code_aether/pkg/v08/scenario"
	v09scenario "github.com/aether/code_aether/pkg/v09/scenario"
	v10scenario "github.com/aether/code_aether/pkg/v10/scenario"
	relaypolicy "github.com/aether/code_aether/pkg/v11/relaypolicy"
)

var (
	dispatchScenarioFn = dispatchScenario

	runMode              = flag.String("mode", "client", "runtime mode: client|relay|bootstrap")
	scenario             = flag.String("scenario", "", "optional scenario: create-server|join-deeplink|first-contact|v08-echo|v09-forge|v10-genesis")
	firstContactRuns     = flag.Int("first-contact-runs", 3, "number of repeated first-contact runs")
	firstContactOut      = flag.String("first-contact-output", "artifacts/generated/first-contact", "output directory for first-contact scenario artifacts")
	firstContactGoal     = flag.Duration("first-contact-target", 5*time.Minute, "target duration for each first-contact run")
	serverID             = flag.String("server-id", "aether-server", "server identifier for manifest scenarios")
	identity             = flag.String("identity", "aether-identity", "identity string used when signing manifests and joining")
	description          = flag.String("description", "phase6 stub server", "server description for manifest metadata")
	version              = flag.Int("version", 1, "manifest version value")
	chatEnabled          = flag.Bool("capability-chat", true, "advertise chat capability")
	voiceEnabled         = flag.Bool("capability-voice", false, "advertise voice capability")
	deeplink             = flag.String("deeplink", "", "deeplink URI for join-deeplink scenario")
	seedManifest         = flag.Bool("seed-manifest", false, "seed manifest store before join handshake")
	relayListen          = flag.String("relay-listen", "0.0.0.0:4001", "relay listen address host:port")
	relayStore           = flag.String("relay-store", "./artifacts/generated/relay-store", "relay store-and-forward data directory")
	relayPersistenceMode = flag.String("relay-persistence-mode", "session-metadata", "relay persistence intent (none|session-metadata|transient-metadata|durable-message-body|attachment-payload|media-frame-archive)")
	relayHealth          = flag.Duration("relay-health-interval", 30*time.Second, "relay health status emission interval")
	relayReservations    = flag.Int("relay-reservation-limit", 256, "maximum concurrent relay reservations")
	relaySessionTTL      = flag.Duration("relay-session-timeout", 2*time.Minute, "maximum relay session lifetime")
	relayMaxBytesSec     = flag.Int64("relay-max-bytes-per-sec", 1_000_000, "per-session relay bandwidth budget in bytes/sec")
	profile              = flag.String("profile", "default", "operator profile for demos")
)

type scenarioHandlers struct {
	runCreateServer func(*phase6.ManifestStore)
	runJoinDeepLink func(*phase6.ManifestStore)
	runFirstContact func()
	runRelayMode    func()
	runV08Echo      func() error
	runV09Forge     func() error
	runV10Genesis   func() error
}

func defaultScenarioHandlers() scenarioHandlers {
	return scenarioHandlers{
		runCreateServer: runCreateServer,
		runJoinDeepLink: runJoinDeepLink,
		runFirstContact: runFirstContactScenario,
		runRelayMode:    runRelayMode,
		runV08Echo:      runV08EchoScenario,
		runV09Forge:     runV09ForgeScenario,
		runV10Genesis:   runV10GenesisScenario,
	}
}

func main() {
	flag.Parse()
	store := phase6.NewManifestStore(0)
	exitCode := dispatchScenarioFn(*runMode, *scenario, store, defaultScenarioHandlers())
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func dispatchScenario(mode string, scenario string, store *phase6.ManifestStore, handlers scenarioHandlers) int {
	switch mode {
	case "client", "relay", "bootstrap":
		// valid modes maintained for compatibility.
	default:
		fmt.Fprintf(os.Stderr, "invalid --mode %q; expected client|relay|bootstrap\n", mode)
		return 2
	}

	switch strings.ToLower(scenario) {
	case "":
		if mode == "relay" {
			handlers.runRelayMode()
			return 0
		}
		fmt.Printf("Phase 2 foundation stub: cmd/aether mode=%s profile=%s\n", mode, *profile)
		return 0
	case "create-server":
		handlers.runCreateServer(store)
		return 0
	case "join-deeplink":
		handlers.runJoinDeepLink(store)
		return 0
	case "first-contact":
		handlers.runFirstContact()
		return 0
	case "v08-echo":
		if err := handlers.runV08Echo(); err != nil {
			fmt.Fprintf(os.Stderr, "v0.8 echo: FAIL: %v\n", err)
			return 6
		}
		fmt.Println("v0.8 echo: PASS")
		return 0
	case "v09-forge":
		if err := handlers.runV09Forge(); err != nil {
			fmt.Fprintf(os.Stderr, "v0.9 forge: FAIL: %v\n", err)
			return 7
		}
		fmt.Println("v0.9 forge: PASS")
		return 0
	case "v10-genesis":
		if err := handlers.runV10Genesis(); err != nil {
			fmt.Fprintf(os.Stderr, "v1.0 genesis: FAIL: %v\n", err)
			return 8
		}
		fmt.Println("v1.0 genesis: PASS")
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown scenario %q; valid scenarios: create-server, join-deeplink, first-contact, v08-echo, v09-forge, v10-genesis\n", scenario)
		return 3
	}
}

func runFirstContactScenario() {
	summary, runs, err := phase11.RunFirstContact(context.Background(), phase11.Options{
		Runs:           *firstContactRuns,
		OutputDir:      *firstContactOut,
		ServerIDPrefix: *serverID,
		IdentityPrefix: *identity,
		TargetDuration: *firstContactGoal,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "first-contact scenario failed: %v\n", err)
		os.Exit(15)
	}

	fmt.Printf("First-contact baseline generated: runs=%d passed=%d failed=%d pass_rate=%.2f output=%s\n",
		summary.RunsCompleted,
		summary.PassedRuns,
		summary.FailedRuns,
		summary.PassRate,
		*firstContactOut,
	)
	fmt.Printf("Duration metrics: target=%s mean_ms=%d median_ms=%d\n", summary.TargetDuration, summary.MeanDurationMS, summary.MedianDurationMS)
	for _, run := range runs {
		fmt.Printf("Run %02d success=%t target_met=%t duration=%s\n", run.RunID, run.Success, run.TargetMet, run.Duration)
		if !run.Success {
			fmt.Printf("  failure: %s (owner: %s)\n", run.FailureReason, run.FailureOwner)
		}
	}
}

func validateRelayPersistenceMode() error {
	mode, err := relaypolicy.ParsePersistenceMode(*relayPersistenceMode)
	if err != nil {
		return err
	}
	return relaypolicy.ValidateMode(mode)
}

func runRelayMode() {
	if err := validateRelayPersistenceMode(); err != nil {
		fmt.Fprintf(os.Stderr, "relay persistence policy violation: %v\n", err)
		os.Exit(16)
	}

	if strings.TrimSpace(*relayListen) == "" {
		fmt.Fprintln(os.Stderr, "invalid relay configuration: --relay-listen must be non-empty host:port")
		os.Exit(11)
	}
	if strings.TrimSpace(*relayStore) == "" {
		fmt.Fprintln(os.Stderr, "invalid relay configuration: --relay-store must be non-empty path")
		os.Exit(12)
	}
	if *relayHealth <= 0 {
		fmt.Fprintln(os.Stderr, "invalid relay configuration: --relay-health-interval must be greater than 0")
		os.Exit(13)
	}

	service, err := phase9.NewService(phase9.Config{
		ReservationLimit: *relayReservations,
		SessionTimeout:   *relaySessionTTL,
		MaxBytesPerSec:   *relayMaxBytesSec,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid relay policy: %v\n", err)
		os.Exit(14)
	}

	startedAt := time.Now().UTC()
	snapshot := service.Snapshot()
	fmt.Printf("Relay runtime active mode=relay listen=%s store=%s profile=%s\n", *relayListen, *relayStore, *profile)
	fmt.Printf("Relay policy: reservation_limit=%d session_timeout=%s max_bytes_per_sec=%d active=%d rejected=%d timed_out=%d established=%d\n",
		snapshot.ReservationLimit,
		snapshot.SessionTimeout,
		snapshot.MaxBytesPerSec,
		snapshot.Active,
		snapshot.Rejected,
		snapshot.TimedOut,
		snapshot.Established,
	)
	fmt.Printf("Relay health status: state=ready started_at=%s next_health_in=%s\n", startedAt.Format(time.RFC3339Nano), relayHealth.String())
}

func runCreateServer(store *phase6.ManifestStore) {
	manifest := &phase6.Manifest{
		ServerID:    *serverID,
		Version:     *version,
		Description: *description,
		UpdatedAt:   time.Now().UTC(),
		Capabilities: phase6.Capabilities{
			Chat:  *chatEnabled,
			Voice: *voiceEnabled,
		},
	}

	sig, err := manifest.Sign(*identity)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to sign manifest: %v\n", err)
		os.Exit(4)
	}

	if err := store.Publish(manifest); err != nil {
		fmt.Fprintf(os.Stderr, "failed to publish manifest: %v\n", err)
		os.Exit(5)
	}

	state := phase6.NewServerState(manifest)
	state.AddMetadata("cli-scenario", "create-server")

	fmt.Printf("Created server manifest for %s\n", manifest.ServerID)
	fmt.Printf("Signed at %s with signature %s\n", manifest.UpdatedAt.Format(time.RFC3339Nano), sig)
	fmt.Printf("Local state metadata: %+v\n", state.LocalMetadata)
}

func runV08EchoScenario() error {
	return v08scenario.RunEchoContracts()
}

func runV09ForgeScenario() error {
	return v09scenario.RunForgeScenario()
}

func runV10GenesisScenario() error {
	return v10scenario.RunGenesisScenario()
}

func runJoinDeepLink(store *phase6.ManifestStore) {
	if *deeplink == "" {
		fmt.Fprintln(os.Stderr, "--deeplink is required for join-deeplink scenario")
		os.Exit(6)
	}

	link, err := phase6.ParseJoinDeepLink(*deeplink)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse deeplink: %v\n", err)
		os.Exit(7)
	}

	if *seedManifest {
		seed := &phase6.Manifest{
			ServerID:    link.ServerID,
			Version:     *version,
			Description: *description,
			UpdatedAt:   time.Now().UTC(),
			Capabilities: phase6.Capabilities{
				Chat:  *chatEnabled,
				Voice: *voiceEnabled,
			},
		}
		_, seedSignErr := seed.Sign(*identity)
		if seedSignErr != nil {
			fmt.Fprintf(os.Stderr, "failed to sign seed manifest: %v\n", seedSignErr)
			os.Exit(8)
		}
		publishErr := store.Publish(seed)
		if publishErr != nil {
			fmt.Fprintf(os.Stderr, "failed to publish seed manifest: %v\n", publishErr)
			os.Exit(9)
		}
	}

	handshake := phase6.NewHandshakeMachine(store, *identity)
	state, err := handshake.Join(link)
	if err != nil {
		fmt.Fprintf(os.Stderr, "join handshake failed: %v\n", err)
		os.Exit(10)
	}

	fmt.Printf("Handshake succeeded for %s\n", state.ServerID)
	fmt.Printf("Membership status: %s, chat enabled: %t, voice enabled: %t\n", state.Status, state.ChatReady, state.VoiceReady)
	fmt.Printf("Last handshake: %s, retries: %d\n", state.LastHandshake.Format(time.RFC3339Nano), state.RetryCount)
}
