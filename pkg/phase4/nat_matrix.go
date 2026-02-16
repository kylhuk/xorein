package phase4

import "fmt"

// NATTopology classifies a home-network environment for traversal planning.
type NATTopology string

const (
	NATTopologySameLAN            NATTopology = "same-lan"
	NATTopologyFullCone           NATTopology = "full-cone"
	NATTopologyRestrictedCone     NATTopology = "restricted-cone"
	NATTopologyPortRestrictedCone NATTopology = "port-restricted-cone"
	NATTopologySymmetric          NATTopology = "symmetric"
	NATTopologyDoubleNAT          NATTopology = "double-nat"
	NATTopologyCarrierGradeNAT    NATTopology = "carrier-grade-nat"
)

// MitigationPriority captures urgency for high-risk NAT mitigations.
type MitigationPriority string

const (
	MitigationPriorityP0 MitigationPriority = "P0"
	MitigationPriorityP1 MitigationPriority = "P1"
	MitigationPriorityP2 MitigationPriority = "P2"
)

// NATMitigation ties a high-risk outcome to a concrete action and validation linkage.
type NATMitigation struct {
	ID            string
	Priority      MitigationPriority
	Action        string
	Diagnostic    ConnectivityReasonCode
	ValidationRef string
}

// NATScenarioEntry defines expected traversal behavior for a target NAT environment.
type NATScenarioEntry struct {
	ID                string
	Topology          NATTopology
	Conditions        string
	TraversalSequence []TraversalStage
	HighRisk          bool
	RiskReason        string
	Mitigations       []NATMitigation
}

// DefaultNATScenarioMatrix returns the v0.1 NAT environment matrix linked to fallback behavior.
func DefaultNATScenarioMatrix() []NATScenarioEntry {
	return []NATScenarioEntry{
		{
			ID:                "nat-same-lan-mdns",
			Topology:          NATTopologySameLAN,
			Conditions:        "both peers on same subnet with multicast available",
			TraversalSequence: []TraversalStage{TraversalStageDirect},
			HighRisk:          false,
		},
		{
			ID:                "nat-consumer-full-cone",
			Topology:          NATTopologyFullCone,
			Conditions:        "single router with permissive endpoint mapping",
			TraversalSequence: []TraversalStage{TraversalStageDirect, TraversalStageAutoNAT},
			HighRisk:          false,
		},
		{
			ID:                "nat-port-restricted-with-upnp-off",
			Topology:          NATTopologyPortRestrictedCone,
			Conditions:        "endpoint-dependent filtering with no automatic port mapping",
			TraversalSequence: []TraversalStage{TraversalStageDirect, TraversalStageAutoNAT, TraversalStageHolePunch},
			HighRisk:          false,
		},
		{
			ID:                "nat-symmetric-hard-fallback",
			Topology:          NATTopologySymmetric,
			Conditions:        "strict endpoint-dependent mapping and filtering",
			TraversalSequence: []TraversalStage{TraversalStageDirect, TraversalStageAutoNAT, TraversalStageHolePunch, TraversalStageRelay},
			HighRisk:          true,
			RiskReason:        "direct and hole-punch paths frequently fail, relay pressure increases",
			Mitigations: []NATMitigation{
				{
					ID:            "mitigate-relay-capacity-alert",
					Priority:      MitigationPriorityP0,
					Action:        "surface relay fallback activation in client diagnostics and trigger relay-capacity alerting",
					Diagnostic:    ReasonRelayFallbackActive,
					ValidationRef: "TestTraversalRunnerRelayReservationLifecycle/active_reservation",
				},
				{
					ID:            "mitigate-relay-reservation-failure",
					Priority:      MitigationPriorityP0,
					Action:        "treat relay reservation failure as explicit degraded terminal state with actionable retry messaging",
					Diagnostic:    ReasonRelayReservationFailed,
					ValidationRef: "TestTraversalRunnerRelayReservationLifecycle/failed_reservation",
				},
			},
		},
		{
			ID:                "nat-double-nat-isp-router",
			Topology:          NATTopologyDoubleNAT,
			Conditions:        "home router behind upstream managed gateway",
			TraversalSequence: []TraversalStage{TraversalStageDirect, TraversalStageAutoNAT, TraversalStageHolePunch, TraversalStageRelay},
			HighRisk:          true,
			RiskReason:        "inbound reachability and hole punching are unstable under layered mappings",
			Mitigations: []NATMitigation{
				{
					ID:            "mitigate-timeout-observability",
					Priority:      MitigationPriorityP1,
					Action:        "track timeout reason-code frequency to separate temporary packet loss from persistent topology limits",
					Diagnostic:    ReasonTraversalTimeout,
					ValidationRef: "TestTraversalRunnerStageTimeoutEvent",
				},
				{
					ID:            "mitigate-recovery-signal",
					Priority:      MitigationPriorityP1,
					Action:        "capture recovery transition after transient failures to validate fallback chain resilience",
					Diagnostic:    ReasonRecoveryTriggered,
					ValidationRef: "TestTraversalRunnerFallbackOrder",
				},
			},
		},
		{
			ID:                "nat-cgnat-mobile-broadband",
			Topology:          NATTopologyCarrierGradeNAT,
			Conditions:        "carrier-managed NAT with shared public address space",
			TraversalSequence: []TraversalStage{TraversalStageDirect, TraversalStageAutoNAT, TraversalStageHolePunch, TraversalStageRelay},
			HighRisk:          true,
			RiskReason:        "peer-to-peer reachability is often blocked; relay dependency can become chronic",
			Mitigations: []NATMitigation{
				{
					ID:            "mitigate-relay-first-degradation",
					Priority:      MitigationPriorityP0,
					Action:        "escalate to relay fallback quickly when direct + hole punch fail and keep reason-code trail intact",
					Diagnostic:    ReasonRelayFallbackActive,
					ValidationRef: "TestP4T8TraversalFailureObservations",
				},
				{
					ID:            "mitigate-stage-failure-taxonomy",
					Priority:      MitigationPriorityP1,
					Action:        "require normalized per-stage failure reason-codes for support triage and QoL no-limbo messaging",
					Diagnostic:    ReasonHolePunchFailure,
					ValidationRef: "TestP4T8TraversalFailureObservations",
				},
			},
		},
	}
}

// NATScenarioByID returns one scenario by ID.
func NATScenarioByID(id string) (NATScenarioEntry, bool) {
	for _, scenario := range DefaultNATScenarioMatrix() {
		if scenario.ID == id {
			return scenario, true
		}
	}
	return NATScenarioEntry{}, false
}

// ValidateNATScenarioMatrix verifies matrix entries meet v0.1 acceptance constraints.
func ValidateNATScenarioMatrix(matrix []NATScenarioEntry) error {
	if len(matrix) == 0 {
		return fmt.Errorf("nat scenario matrix is empty")
	}
	for _, scenario := range matrix {
		if scenario.ID == "" {
			return fmt.Errorf("nat scenario has empty id")
		}
		if len(scenario.TraversalSequence) == 0 {
			return fmt.Errorf("nat scenario %s has no traversal sequence", scenario.ID)
		}
		if scenario.TraversalSequence[0] != TraversalStageDirect {
			return fmt.Errorf("nat scenario %s must start with direct traversal", scenario.ID)
		}
		if scenario.HighRisk {
			if scenario.RiskReason == "" {
				return fmt.Errorf("nat scenario %s is high risk but missing risk reason", scenario.ID)
			}
			if len(scenario.Mitigations) == 0 {
				return fmt.Errorf("nat scenario %s is high risk but has no mitigations", scenario.ID)
			}
			for _, mitigation := range scenario.Mitigations {
				if mitigation.ID == "" || mitigation.Action == "" || mitigation.ValidationRef == "" {
					return fmt.Errorf("nat scenario %s has incomplete mitigation metadata", scenario.ID)
				}
			}
		}
	}
	return nil
}
