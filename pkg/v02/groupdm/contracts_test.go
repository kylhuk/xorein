package groupdm

import "testing"

func TestValidateMemberCap(t *testing.T) {
	status := ValidateMemberCap(39, 1)
	if !status.Allowed || !status.Warning {
		t.Fatalf("expected warning allow, got %+v", status)
	}
	rejected := ValidateMemberCap(50, 1)
	if rejected.Allowed {
		t.Fatalf("expected cap rejection, got %+v", rejected)
	}
	if rejected.Reason != ReasonMemberCapExceeded {
		t.Fatalf("reason=%q want %q", rejected.Reason, ReasonMemberCapExceeded)
	}
}

func TestNormalizeMembershipEventsStableOrder(t *testing.T) {
	in := []MembershipEvent{
		{EventID: "b", Sequence: 3},
		{EventID: "a", Sequence: 3},
		{EventID: "c", Sequence: 1},
	}
	out := NormalizeMembershipEvents(in)
	if out[0].EventID != "c" || out[1].EventID != "a" || out[2].EventID != "b" {
		t.Fatalf("unexpected order: %#v", out)
	}
}

func TestLifecycleTransitions(t *testing.T) {
	result := EvaluateMemberTransition(MemberStateNone, EventInvite)
	if result.Next != MemberStateInvited || result.Reason != ReasonTransitionApplied {
		t.Fatalf("invite transition mismatch: %+v", result)
	}
	reject := EvaluateMemberTransition(MemberStateNone, EventLeave)
	if reject.Next != MemberStateNone || reject.Reason != ReasonInvalidTransition {
		t.Fatalf("invalid transition mismatch: %+v", reject)
	}
}

func TestValidateSenderKeyEnvelope(t *testing.T) {
	authorized := map[string]struct{}{"alice": {}}
	base := SenderKeyEnvelope{
		EnvelopeID:          "env-1",
		GroupID:             "g1",
		SenderID:            "alice",
		RecipientID:         "bob",
		Epoch:               1,
		Nonce:               "n1",
		Ciphertext:          []byte{1},
		SignatureAlgorithm:  "ed25519",
		Signature:           []byte{2},
		ReplayProtectionTag: "rp1",
	}

	valid := ValidateSenderKeyEnvelope(base, authorized)
	if !valid.Accepted || valid.Reason != ReasonSenderKeyAccepted {
		t.Fatalf("valid envelope mismatch: %+v", valid)
	}

	unauthorized := base
	unauthorized.SenderID = "mallory"
	denied := ValidateSenderKeyEnvelope(unauthorized, authorized)
	if denied.Accepted || denied.Reason != ReasonSenderKeyUnauthorizedSender {
		t.Fatalf("unauthorized envelope mismatch: %+v", denied)
	}

	missingSig := base
	missingSig.Signature = nil
	denied = ValidateSenderKeyEnvelope(missingSig, authorized)
	if denied.Accepted || denied.Reason != ReasonSenderKeyMissingSignature {
		t.Fatalf("missing signature mismatch: %+v", denied)
	}

	missingReplay := base
	missingReplay.ReplayProtectionTag = ""
	denied = ValidateSenderKeyEnvelope(missingReplay, authorized)
	if denied.Accepted || denied.Reason != ReasonSenderKeyReplayUnprotected {
		t.Fatalf("missing replay protection mismatch: %+v", denied)
	}
}

func TestEvaluateSenderKeyBootstrapRetry(t *testing.T) {
	retry := EvaluateSenderKeyBootstrapRetry(0, 3, true)
	if retry.Accepted || !retry.Retry || retry.Reason != ReasonSenderKeyRetryScheduled {
		t.Fatalf("retry scheduled mismatch: %+v", retry)
	}
	exhausted := EvaluateSenderKeyBootstrapRetry(2, 3, true)
	if exhausted.Accepted || exhausted.Retry || exhausted.Reason != ReasonSenderKeyRetryExhausted {
		t.Fatalf("retry exhausted mismatch: %+v", exhausted)
	}
	terminal := EvaluateSenderKeyBootstrapRetry(0, 3, false)
	if terminal.Accepted || terminal.Retry || terminal.Reason != ReasonSenderKeyRetryTerminalFailure {
		t.Fatalf("retry terminal mismatch: %+v", terminal)
	}
}

func TestEvaluateRekeyTriggerAndRejoin(t *testing.T) {
	memberRemoved := EvaluateRekeyTrigger(RekeyTriggerMemberRemoved)
	if !memberRemoved.Mandatory || memberRemoved.Reason != ReasonRekeyRequiredMembershipChange {
		t.Fatalf("member removed rekey mismatch: %+v", memberRemoved)
	}
	rejoinGate := EvaluateRekeyTrigger(RekeyTriggerRejoinAfterDrop)
	if !rejoinGate.Mandatory || rejoinGate.RejoinAllowed || rejoinGate.Reason != ReasonRekeyRequiredRejoinGate {
		t.Fatalf("rejoin gate mismatch: %+v", rejoinGate)
	}
	if CanRejoinAfterRemoval(5, 5, true) {
		t.Fatal("expected no rejoin when epoch unchanged")
	}
	if !CanRejoinAfterRemoval(5, 6, true) {
		t.Fatal("expected rejoin when epoch advanced and rekey completed")
	}
	if CanRejoinAfterRemoval(5, 6, false) {
		t.Fatal("expected rejoin blocked without completed rekey")
	}
}

func TestSelectGroupTransport(t *testing.T) {
	direct := SelectGroupTransport(true, true, true)
	if direct.Path != GroupTransportPathDirect || direct.Reason != ReasonGroupTransportDirectAvailable {
		t.Fatalf("direct route mismatch: %+v", direct)
	}
	offline := SelectGroupTransport(false, true, true)
	if offline.Path != GroupTransportPathOffline || offline.Reason != ReasonGroupTransportOfflineFallback {
		t.Fatalf("offline route mismatch: %+v", offline)
	}
	reject := SelectGroupTransport(false, false, false)
	if reject.Path != GroupTransportPathReject || reject.Reason != ReasonGroupTransportRejected {
		t.Fatalf("reject route mismatch: %+v", reject)
	}
}

func TestResolveHistorySync(t *testing.T) {
	locked := ResolveHistorySync(100, 1, 200, 200, 0, false)
	if !locked.Locked || locked.FromSeq != 100 || locked.ToSeq != 200 || locked.Reason != ReasonHistoryLocked || !locked.FromJoinTime {
		t.Fatalf("locked history mismatch: %+v", locked)
	}

	bounded := ResolveHistorySync(100, 100, 300, 300, 50, true)
	if bounded.Locked || bounded.FromSeq != 251 || bounded.ToSeq != 300 || bounded.Reason != ReasonHistoryWindowBound {
		t.Fatalf("bounded history mismatch: %+v", bounded)
	}

	invalid := ResolveHistorySync(100, 10, 5, 10, 0, true)
	if invalid.Reason != ReasonHistoryInvalid {
		t.Fatalf("invalid history mismatch: %+v", invalid)
	}
}

func TestEvaluateGrowthAndConvertPlan(t *testing.T) {
	warn := EvaluateGrowth(39, 1)
	if !warn.Allowed || !warn.Warning || warn.Reason != ReasonGrowthWarning {
		t.Fatalf("growth warning mismatch: %+v", warn)
	}

	convert := EvaluateGrowth(50, 1)
	if convert.Allowed || !convert.ConvertToServer || convert.HistoryTransferable || convert.Reason != ReasonGrowthConvertRequired || convert.DisclosureCode != HistoryDisclosureCodeNotTransferable {
		t.Fatalf("growth conversion mismatch: %+v", convert)
	}

	plan := BuildConvertPlan()
	if !plan.CreateServer || !plan.CreateInitialChannel || !plan.MigrateMemberList || !plan.PostSystemNotice || plan.HistoryTransferable || plan.DisclosureCode != HistoryDisclosureCodeNotTransferable {
		t.Fatalf("convert plan mismatch: %+v", plan)
	}
}
