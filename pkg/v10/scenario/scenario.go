package scenario

import (
	"fmt"
	"strings"

	"github.com/aether/code_aether/pkg/v10/conformance"
	"github.com/aether/code_aether/pkg/v10/docs"
	"github.com/aether/code_aether/pkg/v10/governance"
	"github.com/aether/code_aether/pkg/v10/network"
	"github.com/aether/code_aether/pkg/v10/publication"
	"github.com/aether/code_aether/pkg/v10/relay"
	"github.com/aether/code_aether/pkg/v10/release"
	"github.com/aether/code_aether/pkg/v10/repro"
	"github.com/aether/code_aether/pkg/v10/security"
	"github.com/aether/code_aether/pkg/v10/store"
	"github.com/aether/code_aether/pkg/v10/website"
)

// RunGenesisScenario deterministically validates key v1.0 invariants.
func RunGenesisScenario() error {
	var problems []string

	if len(security.AssetInventory()) < 3 {
		problems = append(problems, "security asset inventory incomplete")
	}
	if len(security.ThreatModel()) < 3 {
		problems = append(problems, "threat model underspecified")
	}
	if len(security.EngagementCriteria()) < 3 {
		problems = append(problems, "engagement criteria missing entries")
	}
	if len(security.FindingLifecycle()) < 3 {
		problems = append(problems, "finding lifecycle lacks classes")
	}

	nodes := store.BootstrapNodes(10)
	if len(nodes) < 10 {
		problems = append(problems, "bootstrap nodes insufficient")
	}
	topology := network.Topology(nodes)
	if len(topology) != len(nodes) {
		problems = append(problems, "topology map mismatch")
	}
	if len(network.ContinuityPlan()) != 4 {
		problems = append(problems, "continuity plan incomplete")
	}

	if len(publication.SectionMap()) == 0 {
		problems = append(problems, "spec sections missing")
	}
	if len(publication.SectionList()) != 3 {
		problems = append(problems, "spec section list incomplete")
	}

	claims := website.LandingClaims()
	if len(claims) < 3 {
		problems = append(problems, "landing claims insufficient")
	}

	relScores := relay.ReliabilityScores()
	if len(relScores) < 3 {
		problems = append(problems, "relay reliability underspecified")
	}
	abuse := relay.AbuseResponseClass()
	if abuse["high"] == "" {
		problems = append(problems, "abuse response missing high severity")
	}

	releaseChecklist := release.ManifestChecklist()
	if len(releaseChecklist) == 0 {
		problems = append(problems, "release checklist empty")
	}
	compliance := release.DistributionCompliance()
	if len(compliance) < 4 {
		problems = append(problems, "distribution compliance missing entries")
	}

	docsChecklist := docs.UserGuideChecklist()
	if len(docsChecklist) != 3 {
		problems = append(problems, "user guide checklist missing sections")
	}
	adminChecklist := docs.AdminGuideChecklist()
	if len(adminChecklist) != 3 {
		problems = append(problems, "admin guide checklist missing chapters")
	}
	bodyGuide := docs.DeveloperGuideChecklist()
	if len(bodyGuide) != 4 {
		problems = append(problems, "developer checklist incomplete")
	}

	naming := governance.NamingGovernance()
	if naming["client"] != "Harmolyn" {
		problems = append(problems, "naming governance mismatch")
	}
	if naming["backend"] != "xorein" {
		problems = append(problems, "backend name mismatch")
	}
	if len(governance.AdditiveChecklist()) < 3 {
		problems = append(problems, "additive checklist shallow")
	}
	if !strings.Contains(governance.MajorPathTriggerClassifier(true, "multistream"), "major-path") {
		problems = append(problems, "major-path classifier missing")
	}

	pins := repro.BuildPins()
	if len(pins) == 0 {
		problems = append(problems, "build pins empty")
	}
	steps := repro.VerificationSteps()
	if len(steps) == 0 {
		problems = append(problems, "verification steps missing")
	}

	checklist := conformance.Checklist{
		"VA-G1": true,
		"VA-G2": true,
		"VA-G3": true,
		"VA-G4": true,
		"VA-G5": true,
		"VA-G6": true,
	}
	result, err := conformance.ValidateChecklist("V10-G0", checklist)
	if err != nil {
		return fmt.Errorf("gate validation failure: %w", err)
	}
	if !result.Ready {
		problems = append(problems, fmt.Sprintf("gate %s not ready", result.Gate.ID))
	}

	summary := fmt.Sprintf("nodes=%d topology=%d relay=%d claims=%d pins=%d", len(nodes), len(topology), len(relScores), len(claims), len(pins))
	if len(problems) > 0 {
		return fmt.Errorf("v1.0 genesis failed: %s | %s", strings.Join(problems, "; "), summary)
	}
	return nil
}
