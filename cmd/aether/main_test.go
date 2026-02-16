package main

import (
	"testing"

	phase6 "github.com/aether/code_aether/pkg/phase6"
)

func TestDispatchScenarioFirstContactInvokesHandler(t *testing.T) {
	called := false
	exitCode := dispatchScenario("client", "first-contact", phase6.NewManifestStore(0), scenarioHandlers{
		runCreateServer: func(*phase6.ManifestStore) {},
		runJoinDeepLink: func(*phase6.ManifestStore) {},
		runFirstContact: func() {
			called = true
		},
		runRelayMode: func() {},
	})

	if exitCode != 0 {
		t.Fatalf("dispatchScenario() exit code = %d, want 0", exitCode)
	}
	if !called {
		t.Fatalf("expected first-contact handler to be invoked")
	}
}

func TestDispatchScenarioUnknownScenarioReturnsCode3(t *testing.T) {
	exitCode := dispatchScenario("client", "not-a-scenario", phase6.NewManifestStore(0), scenarioHandlers{
		runCreateServer: func(*phase6.ManifestStore) {},
		runJoinDeepLink: func(*phase6.ManifestStore) {},
		runFirstContact: func() {},
		runRelayMode:    func() {},
	})

	if exitCode != 3 {
		t.Fatalf("dispatchScenario() exit code = %d, want 3", exitCode)
	}
}

func TestDispatchScenarioRelayModeWithEmptyScenarioInvokesRelayHandler(t *testing.T) {
	called := false
	exitCode := dispatchScenario("relay", "", phase6.NewManifestStore(0), scenarioHandlers{
		runCreateServer: func(*phase6.ManifestStore) {},
		runJoinDeepLink: func(*phase6.ManifestStore) {},
		runFirstContact: func() {},
		runRelayMode: func() {
			called = true
		},
	})

	if exitCode != 0 {
		t.Fatalf("dispatchScenario() exit code = %d, want 0", exitCode)
	}
	if !called {
		t.Fatalf("expected relay mode handler to be invoked")
	}
}
