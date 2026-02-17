package push

import (
	"context"
	"testing"
)

func TestServiceRegisterForward(t *testing.T) {
	adapter := &MockAdapter{}
	svc := NewService(adapter)
	if err := svc.Register(context.Background(), Registration{DeviceID: "", ClientID: ""}); err == nil {
		t.Fatalf("expected error for invalid registration")
	}
	reg := Registration{DeviceID: "device-1", ClientID: "client-1", Token: "token"}
	if err := svc.Register(context.Background(), reg); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if err := svc.Register(context.Background(), reg); err != nil {
		t.Fatalf("duplicate register should be idempotent, got %v", err)
	}
	env := BuildPayload("payload-1", "device-1", []byte("data"), map[string]string{AuthMetadataKey: "token"})
	if err := svc.Forward(context.Background(), env); err != nil {
		t.Fatalf("forward failed: %v", err)
	}
	if len(adapter.Sent) != 1 {
		t.Fatalf("expected one forwarded envelope")
	}
	if adapter.Sent[0].Recipient != "device-1" {
		t.Fatalf("unexpected recipient")
	}
}

func TestBuildPayloadCopiesMetadata(t *testing.T) {
	meta := map[string]string{"k": "v"}
	env := BuildPayload("payload-2", "recipient", []byte("data"), meta)
	meta["k"] = "changed"
	if env.Metadata["k"] == "changed" {
		t.Fatalf("expected metadata copy, but got mutated entry")
	}
	if env.CreatedAt.IsZero() {
		t.Fatalf("expected non-zero timestamp")
	}
}

func TestServiceRegisterUnauthorized(t *testing.T) {
	svc := NewService(&MockAdapter{})
	if err := svc.Register(context.Background(), Registration{DeviceID: "device-2", ClientID: "client-2"}); err == nil || err.Error() != "push.registration.unauthorized" {
		t.Fatalf("expected unauthorized registration error, got %v", err)
	}
}

func TestServiceForwardUnknownRecipient(t *testing.T) {
	adapter := &MockAdapter{}
	svc := NewService(adapter)
	if err := svc.Register(context.Background(), Registration{DeviceID: "device-3", ClientID: "client-3", Token: "token"}); err != nil {
		t.Fatalf("setup registration failed: %v", err)
	}
	env := BuildPayload("env-known", "unknown-device", []byte("payload"), nil)
	if err := svc.Forward(context.Background(), env); err == nil || err.Error() != "push.recipient.unknown" {
		t.Fatalf("expected unknown recipient error, got %v", err)
	}
}

func TestServiceForwardUnauthorized(t *testing.T) {
	adapter := &MockAdapter{}
	svc := NewService(adapter)
	if err := svc.Register(context.Background(), Registration{DeviceID: "device-unauth", ClientID: "client-unauth", Token: "token"}); err != nil {
		t.Fatalf("setup registration failed: %v", err)
	}
	env := BuildPayload("env-unauth", "device-unauth", []byte("payload"), map[string]string{"meta": "value"})
	if err := svc.Forward(context.Background(), env); err == nil || err.Error() != "push.payload.unauthorized" {
		t.Fatalf("expected unauthorized payload error, got %v", err)
	}
}

func TestServiceForwardPayloadTooLarge(t *testing.T) {
	adapter := &MockAdapter{}
	svc := NewService(adapter)
	if err := svc.Register(context.Background(), Registration{DeviceID: "device-4", ClientID: "client-4", Token: "token"}); err != nil {
		t.Fatalf("setup registration failed: %v", err)
	}
	payload := BuildPayload("env-large", "device-4", make([]byte, MaxPayloadBytes+1), map[string]string{AuthMetadataKey: "token"})
	if err := svc.Forward(context.Background(), payload); err == nil || err.Error() != "push.payload.too_large" {
		t.Fatalf("expected payload too large error, got %v", err)
	}
}

func TestServiceForwardIdempotent(t *testing.T) {
	adapter := &MockAdapter{}
	svc := NewService(adapter)
	if err := svc.Register(context.Background(), Registration{DeviceID: "device-5", ClientID: "client-5", Token: "token"}); err != nil {
		t.Fatalf("setup registration failed: %v", err)
	}
	env := BuildPayload("env-dup", "device-5", []byte("payload"), map[string]string{"meta": "value", AuthMetadataKey: "token"})
	if err := svc.Forward(context.Background(), env); err != nil {
		t.Fatalf("forward failed: %v", err)
	}
	if err := svc.Forward(context.Background(), env); err != nil {
		t.Fatalf("duplicate forward unexpected error: %v", err)
	}
	if len(adapter.Sent) != 1 {
		t.Fatalf("expected single send despite duplicates, got %d", len(adapter.Sent))
	}
}
