package discordshim

import "testing"

func TestCoverageReportCalculations(t *testing.T) {
	report := CoverageFor(
		[]RestEndpoint{RestChannelsMessages, RestChannelsReactions},
		[]GatewayIntent{IntentMessages},
	)

	if report.RestCoverage != 2.0/3.0 {
		t.Fatalf("rest coverage => %f", report.RestCoverage)
	}

	if report.GatewayCoverage != 0.5 {
		t.Fatalf("gateway coverage => %f", report.GatewayCoverage)
	}

	combined := CombinedScore(report)
	if combined != (report.RestCoverage+report.GatewayCoverage)/2 {
		t.Fatalf("combined score mismatch: %f", combined)
	}

	if MeetsTarget(combined) {
		t.Fatalf("combined score %f should be below target %.2f", combined, CoverageTarget)
	}

	fullReport := CoverageFor(RequiredRestEndpoints, RequiredGatewayIntents)
	if !MeetsTarget(CombinedScore(fullReport)) {
		t.Fatal("full coverage should meet the target")
	}
}

func TestSubsetMappingBehaviors(t *testing.T) {
	covered := []RestEndpoint{RestChannelsMessages, RestGatewayBot}
	restScore := restScore(covered)
	if restScore != 2.0/3.0 {
		t.Fatalf("rest score => %f", restScore)
	}

	gateway := []GatewayIntent{IntentMessages}
	if gatewayScore(gateway) != 0.5 {
		t.Fatalf("gateway score => %f", gatewayScore(gateway))
	}
}
