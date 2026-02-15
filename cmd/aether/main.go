package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	phase6 "github.com/aether/code_aether/pkg/phase6"
)

var (
	runMode      = flag.String("mode", "client", "runtime mode: client|relay|bootstrap")
	scenario     = flag.String("scenario", "", "optional scenario: create-server|join-deeplink")
	serverID     = flag.String("server-id", "aether-server", "server identifier for manifest scenarios")
	identity     = flag.String("identity", "aether-identity", "identity string used when signing manifests and joining")
	description  = flag.String("description", "phase6 stub server", "server description for manifest metadata")
	version      = flag.Int("version", 1, "manifest version value")
	chatEnabled  = flag.Bool("capability-chat", true, "advertise chat capability")
	voiceEnabled = flag.Bool("capability-voice", false, "advertise voice capability")
	deeplink     = flag.String("deeplink", "", "deeplink URI for join-deeplink scenario")
	seedManifest = flag.Bool("seed-manifest", false, "seed manifest store before join handshake")
)

func main() {
	flag.Parse()
	switch *runMode {
	case "client", "relay", "bootstrap":
		// valid modes maintained for compatibility.
	default:
		fmt.Fprintf(os.Stderr, "invalid --mode %q; expected client|relay|bootstrap\n", *runMode)
		os.Exit(2)
	}

	store := phase6.NewManifestStore(0)
	switch strings.ToLower(*scenario) {
	case "":
		fmt.Printf("Phase 2 foundation stub: cmd/aether mode=%s\n", *runMode)
	case "create-server":
		runCreateServer(store)
	case "join-deeplink":
		runJoinDeepLink(store)
	default:
		fmt.Fprintf(os.Stderr, "unknown scenario %q; valid scenarios: create-server, join-deeplink\n", *scenario)
		os.Exit(3)
	}
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
