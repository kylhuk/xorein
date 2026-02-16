package phase11

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	apb "github.com/aether/code_aether/gen/go/proto"
	phase6 "github.com/aether/code_aether/pkg/phase6"
	phase7 "github.com/aether/code_aether/pkg/phase7"
	phase8 "github.com/aether/code_aether/pkg/phase8"
	phase9 "github.com/aether/code_aether/pkg/phase9"
)

const defaultTargetDuration = 5 * time.Minute

type Options struct {
	Runs           int
	OutputDir      string
	ServerIDPrefix string
	IdentityPrefix string
	TargetDuration time.Duration
	RegressionOut  string
}

type StageResult struct {
	Name      string `json:"name"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
	Duration  string `json:"duration"`
	Success   bool   `json:"success"`
	Reason    string `json:"reason,omitempty"`
	Owner     string `json:"owner,omitempty"`
}

type RunResult struct {
	RunID          int           `json:"run_id"`
	StartedAt      string        `json:"started_at"`
	EndedAt        string        `json:"ended_at"`
	Duration       string        `json:"duration"`
	DurationMillis int64         `json:"duration_ms"`
	Target         string        `json:"target"`
	TargetMet      bool          `json:"target_met"`
	Success        bool          `json:"success"`
	FailureReason  string        `json:"failure_reason,omitempty"`
	FailureOwner   string        `json:"failure_owner,omitempty"`
	Stages         []StageResult `json:"stages"`

	EventCounts      map[string]int `json:"event_counts"`
	ReasonCodeCounts map[string]int `json:"reason_code_counts"`
	FallbackCounts   map[string]int `json:"fallback_counts"`
}

type Summary struct {
	GeneratedAt      string         `json:"generated_at"`
	RunsRequested    int            `json:"runs_requested"`
	RunsCompleted    int            `json:"runs_completed"`
	PassedRuns       int            `json:"passed_runs"`
	FailedRuns       int            `json:"failed_runs"`
	PassRate         float64        `json:"pass_rate"`
	TargetDuration   string         `json:"target_duration"`
	MeanDurationMS   int64          `json:"mean_duration_ms"`
	MedianDurationMS int64          `json:"median_duration_ms"`
	EventCounts      map[string]int `json:"event_counts"`
	ReasonCodeCounts map[string]int `json:"reason_code_counts"`
	FallbackCounts   map[string]int `json:"fallback_counts"`
	Failures         []FailureTrace `json:"failures"`
}

type FailureTrace struct {
	RunID  int    `json:"run_id"`
	Reason string `json:"reason"`
	Owner  string `json:"owner"`
}

func RunFirstContact(ctx context.Context, opts Options) (*Summary, []RunResult, error) {
	normalized := normalizeOptions(opts)
	if err := os.MkdirAll(normalized.OutputDir, 0o750); err != nil {
		return nil, nil, fmt.Errorf("create output dir: %w", err)
	}
	if err := os.MkdirAll(normalized.RegressionOut, 0o750); err != nil {
		return nil, nil, fmt.Errorf("create regression output dir: %w", err)
	}

	runs := make([]RunResult, 0, normalized.Runs)
	for runID := 1; runID <= normalized.Runs; runID++ {
		run := executeRun(ctx, normalized, runID)
		runs = append(runs, run)
		if err := writeJSON(filepath.Join(normalized.OutputDir, fmt.Sprintf("run-%02d.json", runID)), run); err != nil {
			return nil, nil, fmt.Errorf("write run artifact %d: %w", runID, err)
		}
	}

	summary := summarize(normalized, runs)
	if err := writeJSON(filepath.Join(normalized.OutputDir, "summary.json"), summary); err != nil {
		return nil, nil, fmt.Errorf("write summary artifact: %w", err)
	}
	if err := writeMarkdownSummary(filepath.Join(normalized.OutputDir, "summary.md"), summary); err != nil {
		return nil, nil, fmt.Errorf("write summary markdown: %w", err)
	}
	if err := writeRegressionReport(filepath.Join(normalized.RegressionOut, "report.txt"), runs); err != nil {
		return nil, nil, fmt.Errorf("write regression report: %w", err)
	}
	if err := writeDefectStatus(filepath.Join(normalized.RegressionOut, "defects.json"), runs); err != nil {
		return nil, nil, fmt.Errorf("write defect triage: %w", err)
	}

	return &summary, runs, nil
}

func normalizeOptions(opts Options) Options {
	if opts.Runs <= 0 {
		opts.Runs = 3
	}
	if strings.TrimSpace(opts.OutputDir) == "" {
		opts.OutputDir = "artifacts/generated/first-contact"
	}
	if strings.TrimSpace(opts.ServerIDPrefix) == "" {
		opts.ServerIDPrefix = "first-contact-server"
	}
	if strings.TrimSpace(opts.IdentityPrefix) == "" {
		opts.IdentityPrefix = "first-contact-identity"
	}
	if opts.TargetDuration <= 0 {
		opts.TargetDuration = defaultTargetDuration
	}
	if strings.TrimSpace(opts.RegressionOut) == "" {
		opts.RegressionOut = "artifacts/generated/regression"
	}
	return opts
}

func executeRun(_ context.Context, opts Options, runID int) RunResult {
	start := time.Now().UTC()
	run := RunResult{
		RunID:            runID,
		StartedAt:        start.Format(time.RFC3339Nano),
		Target:           opts.TargetDuration.String(),
		Success:          true,
		Stages:           make([]StageResult, 0, 7),
		EventCounts:      map[string]int{},
		ReasonCodeCounts: map[string]int{},
		FallbackCounts:   map[string]int{},
	}

	serverID := fmt.Sprintf("%s-%02d", opts.ServerIDPrefix, runID)
	identity := fmt.Sprintf("%s-%02d", opts.IdentityPrefix, runID)

	store := phase6.NewManifestStore(5 * time.Minute)
	bootstrapper := phase7.NewBootstrapper()
	localState, err := bootstrapper.Bootstrap(phase7.ParticipantID(identity))
	if err != nil {
		failRun(&run, "bootstrap", "state_bootstrap_failed", "Core Runtime", err)
		finishRun(&run, opts.TargetDuration)
		return run
	}

	manifest := &phase6.Manifest{
		ServerID:    serverID,
		Version:     phase6.ManifestVersionV1,
		Description: "phase11 first-contact baseline",
		UpdatedAt:   time.Now().UTC(),
		Capabilities: phase6.Capabilities{
			Chat:  true,
			Voice: true,
		},
	}

	if !stage(&run, "manifest_publish", func() error {
		_, signErr := manifest.Sign(identity)
		if signErr != nil {
			return signErr
		}
		return store.Publish(manifest)
	}, "manifest_publish_failed", "Server Engineer") {
		finishRun(&run, opts.TargetDuration)
		return run
	}

	if !stage(&run, "join_deeplink", func() error {
		deeplink := fmt.Sprintf("aether://join/%s", serverID)
		link, parseErr := phase6.ParseJoinDeepLink(deeplink)
		if parseErr != nil {
			return parseErr
		}
		handshake := phase6.NewHandshakeMachine(store, identity)
		state, joinErr := handshake.Join(link)
		if joinErr != nil {
			return joinErr
		}
		if !state.ChatFlowEnabled() || !state.VoiceFlowEnabled() {
			return fmt.Errorf("membership active but chat/voice flow disabled")
		}
		return nil
	}, "join_failed", "Client Engineer") {
		finishRun(&run, opts.TargetDuration)
		return run
	}

	if !stage(&run, "relay_snapshot", func() error {
		relay, relayErr := phase9.NewService(phase9.Config{})
		if relayErr != nil {
			return relayErr
		}
		if !relay.Reserve() {
			return fmt.Errorf("reservation rejected")
		}
		relay.Release()
		s := relay.Snapshot()
		run.EventCounts["relay_snapshot"]++
		run.ReasonCodeCounts["relay_path_ok"]++
		if s.Rejected > 0 {
			maxInt := uint64(^uint(0) >> 1)
			if s.Rejected > maxInt {
				run.FallbackCounts["relay_capacity"] += int(maxInt)
			} else {
				run.FallbackCounts["relay_capacity"] += int(s.Rejected)
			}
		}
		return nil
	}, "relay_unavailable", "Network Engineer") {
		finishRun(&run, opts.TargetDuration)
		return run
	}

	if !stage(&run, "chat_roundtrip", func() error {
		pipelineA := phase7.NewPipeline(phase7.ParticipantID(identity), localState)
		remoteID := phase7.ParticipantID(identity + "-peer")
		remoteState, remoteErr := bootstrapper.Bootstrap(remoteID)
		if remoteErr != nil {
			return remoteErr
		}
		pipelineB := phase7.NewPipeline(remoteID, remoteState)

		msg, sendErr := pipelineA.Send(1, []byte("hello"))
		if sendErr != nil {
			return sendErr
		}
		_, receiveErr := pipelineB.Receive(msg, localState.MLSSecret, localState.Verifier)
		if receiveErr != nil {
			return receiveErr
		}
		run.EventCounts["chat_roundtrip"]++
		run.ReasonCodeCounts["chat_ok"]++
		return nil
	}, "chat_failed", "Messaging Engineer") {
		finishRun(&run, opts.TargetDuration)
		return run
	}

	if !stage(&run, "voice_connect", func() error {
		unixNow := time.Now().UTC().Unix()
		seed := uint64(0)
		if unixNow > 0 {
			seed = uint64(unixNow)
		}
		voice, baselineErr := phase8.NewVoicePipelineBaseline(fmt.Sprintf("voice-%s", serverID), seed)
		if baselineErr != nil {
			return baselineErr
		}
		manager, managerErr := phase8.NewVoiceSessionManager(8, func() error {
			run.FallbackCounts["voice_capacity_fallback"]++
			return nil
		})
		if managerErr != nil {
			return managerErr
		}

		peer := &localPeer{id: "peer-a"}
		joinErr := manager.Join(context.Background(), voice.GetSessionId(), voice.GetCodecProfile(), voice.GetTransportProfile(), peer)
		if joinErr != nil {
			return joinErr
		}
		if manager.ParticipantCount() != 1 {
			return fmt.Errorf("unexpected participant count: %d", manager.ParticipantCount())
		}
		leaveErr := manager.Leave(context.Background(), voice.GetSessionId(), peer.ID())
		if leaveErr != nil {
			return leaveErr
		}
		run.EventCounts["voice_connect"]++
		run.ReasonCodeCounts["voice_ok"]++
		return nil
	}, "voice_failed", "Realtime Engineer") {
		finishRun(&run, opts.TargetDuration)
		return run
	}

	finishRun(&run, opts.TargetDuration)
	return run
}

func stage(run *RunResult, name string, fn func() error, reasonCode, owner string) bool {
	started := time.Now().UTC()
	err := fn()
	ended := time.Now().UTC()
	res := StageResult{
		Name:      name,
		StartedAt: started.Format(time.RFC3339Nano),
		EndedAt:   ended.Format(time.RFC3339Nano),
		Duration:  ended.Sub(started).String(),
		Success:   err == nil,
	}
	if err != nil {
		res.Reason = err.Error()
		res.Owner = owner
		failRun(run, name, reasonCode, owner, err)
	}
	run.Stages = append(run.Stages, res)
	return err == nil
}

func failRun(run *RunResult, stageName, reasonCode, owner string, err error) {
	run.Success = false
	run.FailureReason = fmt.Sprintf("%s: %v", stageName, err)
	run.FailureOwner = owner
	run.ReasonCodeCounts[reasonCode]++
	run.EventCounts["failure"]++
}

func finishRun(run *RunResult, target time.Duration) {
	end := time.Now().UTC()
	started, parseErr := time.Parse(time.RFC3339Nano, run.StartedAt)
	if parseErr != nil {
		started = end
	}
	d := end.Sub(started)
	run.EndedAt = end.Format(time.RFC3339Nano)
	run.Duration = d.String()
	run.DurationMillis = d.Milliseconds()
	run.TargetMet = d <= target && run.Success
}

func summarize(opts Options, runs []RunResult) Summary {
	summary := Summary{
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339Nano),
		RunsRequested:    opts.Runs,
		RunsCompleted:    len(runs),
		TargetDuration:   opts.TargetDuration.String(),
		EventCounts:      map[string]int{},
		ReasonCodeCounts: map[string]int{},
		FallbackCounts:   map[string]int{},
		Failures:         make([]FailureTrace, 0),
	}

	durations := make([]int64, 0, len(runs))
	var total int64
	for _, run := range runs {
		durations = append(durations, run.DurationMillis)
		total += run.DurationMillis
		if run.Success && run.TargetMet {
			summary.PassedRuns++
		} else {
			summary.FailedRuns++
			summary.Failures = append(summary.Failures, FailureTrace{RunID: run.RunID, Reason: run.FailureReason, Owner: run.FailureOwner})
		}
		mergeCounts(summary.EventCounts, run.EventCounts)
		mergeCounts(summary.ReasonCodeCounts, run.ReasonCodeCounts)
		mergeCounts(summary.FallbackCounts, run.FallbackCounts)
	}

	if len(runs) > 0 {
		summary.PassRate = float64(summary.PassedRuns) / float64(len(runs))
		summary.MeanDurationMS = total / int64(len(runs))
		sortedDurations := append([]int64(nil), durations...)
		sort.Slice(sortedDurations, func(i, j int) bool { return sortedDurations[i] < sortedDurations[j] })
		mid := len(sortedDurations) / 2
		if len(sortedDurations)%2 == 0 {
			summary.MedianDurationMS = (sortedDurations[mid-1] + sortedDurations[mid]) / 2
		} else {
			summary.MedianDurationMS = sortedDurations[mid]
		}
	}

	return summary
}

func mergeCounts(dst, src map[string]int) {
	for k, v := range src {
		dst[k] += v
	}
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}

func writeMarkdownSummary(path string, summary Summary) error {
	var b strings.Builder
	b.WriteString("# P11-T1 First-Contact Baseline Summary\n\n")
	b.WriteString(fmt.Sprintf("- Generated: %s\n", summary.GeneratedAt))
	b.WriteString(fmt.Sprintf("- Runs: %d/%d completed\n", summary.RunsCompleted, summary.RunsRequested))
	b.WriteString(fmt.Sprintf("- Passed: %d\n", summary.PassedRuns))
	b.WriteString(fmt.Sprintf("- Failed: %d\n", summary.FailedRuns))
	b.WriteString(fmt.Sprintf("- Pass rate: %.2f\n", summary.PassRate))
	b.WriteString(fmt.Sprintf("- Target duration: %s\n", summary.TargetDuration))
	b.WriteString(fmt.Sprintf("- Mean duration (ms): %d\n", summary.MeanDurationMS))
	b.WriteString(fmt.Sprintf("- Median duration (ms): %d\n", summary.MedianDurationMS))
	b.WriteString("\n## Reason-code counts\n")
	appendCountList(&b, summary.ReasonCodeCounts)
	b.WriteString("\n## Event counts\n")
	appendCountList(&b, summary.EventCounts)
	b.WriteString("\n## Fallback frequencies\n")
	appendCountList(&b, summary.FallbackCounts)
	b.WriteString("\n## Failures\n")
	if len(summary.Failures) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, f := range summary.Failures {
			b.WriteString(fmt.Sprintf("- run %d: %s (owner: %s)\n", f.RunID, f.Reason, f.Owner))
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o600)
}

func appendCountList(b *strings.Builder, counts map[string]int) {
	if len(counts) == 0 {
		b.WriteString("- none\n")
		return
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString(fmt.Sprintf("- %s: %d\n", k, counts[k]))
	}
}

type DefectStatus struct {
	ID          string `json:"id"`
	RunID       int    `json:"run_id"`
	Severity    string `json:"severity"`
	Status      string `json:"status"`
	Owner       string `json:"owner"`
	ReasonCode  string `json:"reason_code"`
	Description string `json:"description"`
}

func writeRegressionReport(path string, runs []RunResult) error {
	var b strings.Builder
	b.WriteString("P11-T2 integrated regression report\n")
	b.WriteString(fmt.Sprintf("generated_at: %s\n", time.Now().UTC().Format(time.RFC3339Nano)))
	b.WriteString(fmt.Sprintf("runs: %d\n\n", len(runs)))

	b.WriteString("per-run matrix:\n")
	for _, run := range runs {
		status := "PASS"
		if !run.Success {
			status = "FAIL"
		}
		b.WriteString(fmt.Sprintf("- run=%02d status=%s target_met=%t duration=%s\n", run.RunID, status, run.TargetMet, run.Duration))

		if len(run.EventCounts) > 0 {
			b.WriteString("  event_counts:")
			for _, kv := range sortedPairs(run.EventCounts) {
				b.WriteString(fmt.Sprintf(" %s=%d", kv.Key, kv.Val))
			}
			b.WriteByte('\n')
		}
		if len(run.ReasonCodeCounts) > 0 {
			b.WriteString("  reason_codes:")
			for _, kv := range sortedPairs(run.ReasonCodeCounts) {
				b.WriteString(fmt.Sprintf(" %s=%d", kv.Key, kv.Val))
			}
			b.WriteByte('\n')
		}
		if len(run.FallbackCounts) > 0 {
			b.WriteString("  fallback_counts:")
			for _, kv := range sortedPairs(run.FallbackCounts) {
				b.WriteString(fmt.Sprintf(" %s=%d", kv.Key, kv.Val))
			}
			b.WriteByte('\n')
		}
		if !run.Success {
			b.WriteString(fmt.Sprintf("  failure: %s (owner=%s)\n", run.FailureReason, run.FailureOwner))
		}
	}

	return os.WriteFile(path, []byte(b.String()), 0o600)
}

func writeDefectStatus(path string, runs []RunResult) error {
	defects := make([]DefectStatus, 0)
	for _, run := range runs {
		if run.Success {
			continue
		}
		reasonCode := dominantReason(run.ReasonCodeCounts)
		defects = append(defects, DefectStatus{
			ID:          fmt.Sprintf("P11-T2-RUN-%02d", run.RunID),
			RunID:       run.RunID,
			Severity:    "P0",
			Status:      "open",
			Owner:       run.FailureOwner,
			ReasonCode:  reasonCode,
			Description: run.FailureReason,
		})
	}

	if len(defects) == 0 {
		defects = append(defects, DefectStatus{
			ID:          "P11-T2-REGRESSION-CLEAN",
			RunID:       0,
			Severity:    "none",
			Status:      "closed",
			Owner:       "qa",
			ReasonCode:  "none",
			Description: "No integrated-regression defects observed in this run set.",
		})
	}

	return writeJSON(path, defects)
}

type pair struct {
	Key string
	Val int
}

func sortedPairs(counts map[string]int) []pair {
	pairs := make([]pair, 0, len(counts))
	for k, v := range counts {
		pairs = append(pairs, pair{Key: k, Val: v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Val == pairs[j].Val {
			return pairs[i].Key < pairs[j].Key
		}
		return pairs[i].Val > pairs[j].Val
	})
	return pairs
}

func dominantReason(counts map[string]int) string {
	pairs := sortedPairs(counts)
	if len(pairs) == 0 {
		return "unknown"
	}
	return pairs[0].Key
}

type localPeer struct{ id string }

func (p *localPeer) ID() string { return p.id }

func (p *localPeer) Connect(context.Context, string, *apb.VoiceCodecProfile, *apb.VoiceTransportProfile) error {
	return nil
}

func (p *localPeer) Disconnect(context.Context, string) error { return nil }
