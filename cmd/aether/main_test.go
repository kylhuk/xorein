package main

import (
	"errors"
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

func TestDispatchScenarioV09ForgeSuccessInvokesHandler(t *testing.T) {
	t.Parallel()
	called := false
	exitCode := dispatchScenario("client", "v09-forge", phase6.NewManifestStore(0), scenarioHandlers{
		runCreateServer: func(*phase6.ManifestStore) {},
		runJoinDeepLink: func(*phase6.ManifestStore) {},
		runFirstContact: func() {},
		runRelayMode:    func() {},
		runV08Echo:      func() error { return nil },
		runV09Forge: func() error {
			called = true
			return nil
		},
	})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	if !called {
		t.Fatalf("expected v09 forge handler to run")
	}
}

func TestDispatchScenarioV09ForgeFailureReportsNonZero(t *testing.T) {
	t.Parallel()
	handlerErr := errors.New("forge boom")
	exitCode := dispatchScenario("client", "v09-forge", phase6.NewManifestStore(0), scenarioHandlers{
		runCreateServer: func(*phase6.ManifestStore) {},
		runJoinDeepLink: func(*phase6.ManifestStore) {},
		runFirstContact: func() {},
		runRelayMode:    func() {},
		runV08Echo:      func() error { return nil },
		runV09Forge: func() error {
			return handlerErr
		},
	})
	if exitCode != 7 {
		t.Fatalf("expected exit 7 on failure, got %d", exitCode)
	}
}

func TestDispatchScenarioV10GenesisSuccess(t *testing.T) {
	t.Parallel()
	called := false
	exitCode := dispatchScenario("client", "v10-genesis", phase6.NewManifestStore(0), scenarioHandlers{
		runCreateServer: func(*phase6.ManifestStore) {},
		runJoinDeepLink: func(*phase6.ManifestStore) {},
		runFirstContact: func() {},
		runRelayMode:    func() {},
		runV08Echo:      func() error { return nil },
		runV09Forge:     func() error { return nil },
		runV10Genesis: func() error {
			called = true
			return nil
		},
	})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	if !called {
		t.Fatal("expected v10 genesis handler to run")
	}
}

func TestDispatchScenarioV10GenesisFailure(t *testing.T) {
	t.Parallel()
	handlerErr := errors.New("genesis boom")
	exitCode := dispatchScenario("client", "v10-genesis", phase6.NewManifestStore(0), scenarioHandlers{
		runCreateServer: func(*phase6.ManifestStore) {},
		runJoinDeepLink: func(*phase6.ManifestStore) {},
		runFirstContact: func() {},
		runRelayMode:    func() {},
		runV08Echo:      func() error { return nil },
		runV09Forge:     func() error { return nil },
		runV10Genesis: func() error {
			return handlerErr
		},
	})
	if exitCode != 8 {
		t.Fatalf("expected exit 8 on failure, got %d", exitCode)
	}
}
