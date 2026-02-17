package main

import "testing"

func TestDispatchPushRelayRelay(t *testing.T) {
	called := false
	exit := dispatchPushRelay("relay", relayHandlers{runRelay: func() { called = true }, runProbe: func() {}})
	if exit != 0 {
		t.Fatalf("expected exit 0, got %d", exit)
	}
	if !called {
		t.Fatalf("expected relay handler invoked")
	}
}

func TestDispatchPushRelayProbe(t *testing.T) {
	called := false
	exit := dispatchPushRelay("probe", relayHandlers{runRelay: func() {}, runProbe: func() { called = true }})
	if exit != 0 {
		t.Fatalf("expected exit 0")
	}
	if !called {
		t.Fatalf("expected probe handler invoked")
	}
}

func TestDispatchPushRelayInvalid(t *testing.T) {
	exit := dispatchPushRelay("bad", relayHandlers{runRelay: func() {}, runProbe: func() {}})
	if exit != 2 {
		t.Fatalf("expected exit 2 for invalid mode, got %d", exit)
	}
}
