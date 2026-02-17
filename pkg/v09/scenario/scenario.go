package scenario

import (
	"fmt"
	"strings"

	"github.com/aether/code_aether/pkg/v09/conformance"
	"github.com/aether/code_aether/pkg/v09/governance"
	"github.com/aether/code_aether/pkg/v09/ipfs"
	"github.com/aether/code_aether/pkg/v09/mobile"
	"github.com/aether/code_aether/pkg/v09/relay"
	"github.com/aether/code_aether/pkg/v09/scale"
	"github.com/aether/code_aether/pkg/v09/sfu"
)

// RunForgeScenario exercises the v0.9 helpers deterministically.
func RunForgeScenario() error {
	var problems []string

	meta := ipfs.ContentMeta{Name: "forge-manifest", SizeBytes: 1_200_000_000, Owner: "server-owner", CreatedAt: 1680000000}
	address := ipfs.ContentAddress(meta)
	if address == "" {
		problems = append(problems, "empty content address")
	}
	retention := ipfs.ClassifyRetention(meta.SizeBytes, true)
	if retention != ipfs.RetentionDurable {
		problems = append(problems, "unexpected retention state")
	}
	degraded := ipfs.DegradedBehavior(true, false)
	if degraded.Level == "" {
		problems = append(problems, "invalid degraded outcome")
	}

	plan := scale.HierarchyPlan(1200)
	if len(plan) == 0 {
		problems = append(problems, "empty hierarchy plan")
	}
	lazy := scale.LazyLoadPlan(1200)
	if len(lazy) < 3 {
		problems = append(problems, "lazy load plan underspecified")
	}
	mode := scale.NextSecurityMode(scale.ModeStandard, 1250)
	guidance := scale.ShardingPlan(1250, mode)
	if guidance.ShardCount < 2 {
		problems = append(problems, "sharding guidance lacks shards")
	}

	tiers := sfu.CascadingTiers(250)
	decision := sfu.SelectForwardingPath(tiers, 120)
	failover := sfu.FailoverDecision(true, decision.Priority == "degraded")
	if decision.Path == "" || failover.Action == "" {
		problems = append(problems, "sfu plans incomplete")
	}

	envelope := relay.CapacityEnvelope{Limit: 1200, Control: 120, ActiveMedia: 420, StoreForward: 180, BulkSync: 80}
	util := envelope.UtilizationPercent()
	if util > 100 {
		problems = append(problems, "relay utilization > 100%")
	}
	priorities := relay.OverloadPolicy(util)
	if len(priorities) == 0 {
		problems = append(problems, "overload policy empty")
	}
	recovery := relay.RecoveryPlan(util)
	if recovery.NextStep == "" {
		problems = append(problems, "relay recovery not specified")
	}

	budget := mobile.BackgroundBudget(45, 65)
	wake := mobile.EvaluateWakePolicy(65, true)
	decisionText := mobile.BatteryOptimizationDecision(45, true)
	if string(budget) == "" || (!wake.AllowWake && wake.SuppressionReason == "") {
		problems = append(problems, "mobile decision invalid")
	}
	if decisionText == "" {
		problems = append(problems, "mobile optimization decision missing")
	}

	checklist := conformance.Checklist{"VA-G1": true, "VA-G2": true}
	result, err := conformance.ValidateChecklist("V9-G0", checklist)
	if err != nil {
		return fmt.Errorf("gate validation failure: %w", err)
	}
	if !result.Ready {
		problems = append(problems, fmt.Sprintf("gate %s not ready", result.Gate.ID))
	}

	additive := governance.AdditiveChecklist()
	if len(additive) == 0 {
		problems = append(problems, "additive checklist empty")
	}
	classifier := governance.MajorPathTriggerClassifier(true, "multistream")
	if !strings.Contains(classifier, "AEP") {
		problems = append(problems, "major-path classifier missing AEP reference")
	}
	license := governance.LicensingStatus("AGPL", "CC-BY-SA")
	if license == "" {
		problems = append(problems, "licensing status missing")
	}

	summary := fmt.Sprintf("address=%s retention=%s degraded=%s mode=%s shards=%d util=%d budget=%s wake=%t(%s)",
		address, retention, degraded.Level, mode, guidance.ShardCount, util, budget, wake.AllowWake, wake.SuppressionReason)
	if len(problems) > 0 {
		return fmt.Errorf("forge scenario failed: %s | %s", strings.Join(problems, "; "), summary)
	}
	if summary == "" {
		return fmt.Errorf("forge scenario could not build summary")
	}
	_ = decision
	_ = failover
	_ = priorities
	_ = recovery
	_ = decisionText
	_ = additive
	_ = classifier
	_ = license
	return nil
}
