package v24

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v24/daemon"
	"github.com/aether/code_aether/pkg/v24/daemon/doctor"
	"github.com/aether/code_aether/pkg/v24/harmolyn/attach"
	"github.com/aether/code_aether/pkg/v24/harmolyn/journeys"
	"github.com/aether/code_aether/pkg/v24/localapi"
)

func TestScenarioDualClientsParallelReadsAndSerializedWrites(t *testing.T) {
	ctx := context.Background()
	state := newScenarioState()
	client := &scenarioLocalClient{
		state:        state,
		readRelease:  state.readRelease,
		writeRelease: state.writeRelease,
	}

	provider := &fixedAttachProvider{token: localapi.SessionToken{Value: "shared-token", ExpiresAt: time.Now().Add(time.Hour)}}
	coordA := journeys.NewCoordinator(provider, client)
	coordB := journeys.NewCoordinator(provider, client)

	var readWG sync.WaitGroup
	readWG.Add(2)
	for i := 0; i < 2; i++ {
		coord := coordA
		if i == 1 {
			coord = coordB
		}
		testIndex := i
		go func() {
			defer readWG.Done()
			if _, err := coord.Read(ctx, fmt.Sprintf("cursor-%d", testIndex)); err != nil {
				t.Errorf("read failed: %v", err)
			}
		}()
	}

	state.waitReadStarts(2)
	close(state.readRelease)
	readWG.Wait()

	if got := state.maxConcurrentReads(); got < 2 {
		t.Fatalf("expected parallel reads, got max concurrent reads=%d", got)
	}

	var writeWG sync.WaitGroup
	writeWG.Add(2)
	for i := 0; i < 2; i++ {
		coord := coordA
		if i == 1 {
			coord = coordB
		}
		payload := fmt.Sprintf("payload-%d", i)
		go func() {
			defer writeWG.Done()
			if err := coord.Send(ctx, payload); err != nil {
				t.Errorf("send failed: %v", err)
			}
		}()
	}

	close(state.writeRelease)
	writeWG.Wait()

	if got := state.maxConcurrentWrites(); got != 1 {
		t.Fatalf("expected serialized writes, got max concurrent writes=%d", got)
	}
}

func TestScenarioDaemonCrashMidCallReattach(t *testing.T) {
	ctx := context.Background()
	coordinator := &attach.Coordinator{
		Probe: &runningProbe{},
		Handshaker: &sequenceHandshaker{
			tokens: []localapi.SessionToken{
				{Value: "token-1", ExpiresAt: time.Now().Add(time.Minute)},
				{Value: "token-2", ExpiresAt: time.Now().Add(time.Minute)},
			},
		},
		Store: attach.NewMemoryTokenStore(),
	}

	client := &crashyLocalClient{crashes: 1}
	journey := journeys.NewCoordinator(coordinator, client)

	if err := journey.Send(ctx, "first message"); err == nil {
		t.Fatalf("expected write failure during crash")
	} else if !errors.Is(err, errDaemonCrashed) {
		t.Fatalf("expected crash failure, got %v", err)
	}

	initial, ok := coordinator.Store.Load()
	if !ok || initial.Value != "token-1" {
		t.Fatalf("unexpected initial token: %+v", initial)
	}

	recovery := coordinator.Reattach(ctx)
	if !recovery.Attached {
		t.Fatalf("expected successful reattach")
	}
	if recovery.Failure != nil {
		t.Fatalf("unexpected reattach failure: %v", recovery.Failure)
	}

	after, ok := coordinator.Store.Load()
	if !ok || after.Value != "token-2" {
		t.Fatalf("expected token rotation after reattach, got %v", after)
	}

	if err := journey.Send(ctx, "after recovery"); err != nil {
		t.Fatalf("send failed after reattach: %v", err)
	}
}

func TestScenarioStaleSocketAutoRepair(t *testing.T) {
	tmp := t.TempDir()
	socketPath := filepath.Join(tmp, "daemon.sock")
	lockPath := filepath.Join(tmp, "daemon.lock")
	if err := os.WriteFile(socketPath, []byte(""), 0o600); err != nil {
		t.Fatalf("seed stale socket: %v", err)
	}

	mgr := daemon.NewLockManager(lockPath)
	reporter := doctor.New(mgr, socketPath)

	stale := reporter.Run()
	if !stale.StaleSocket {
		t.Fatalf("expected stale socket report: %+v", stale)
	}
	if stale.HealthState != doctor.HealthStateStaleSocket {
		t.Fatalf("expected stale socket health state, got %s", stale.HealthState)
	}
	if stale.NextAction != "remove stale socket and restart daemon" {
		t.Fatalf("unexpected stale next action: %s", stale.NextAction)
	}

	if err := os.Remove(socketPath); err != nil {
		t.Fatalf("remove stale socket: %v", err)
	}

	missing := reporter.Run()
	if missing.HealthState != doctor.HealthStateMissingSocket {
		t.Fatalf("expected missing socket state, got %s", missing.HealthState)
	}
	if missing.NextAction != "start daemon to create socket" {
		t.Fatalf("unexpected missing next action: %s", missing.NextAction)
	}

	if err := os.WriteFile(socketPath, []byte(""), 0o600); err != nil {
		t.Fatalf("recreate repaired socket: %v", err)
	}
	lock, err := mgr.Acquire("repair-client", daemon.StateStarting)
	if err != nil {
		t.Fatalf("acquire start lock: %v", err)
	}

	recovered := reporter.Run()
	if recovered.HealthState != doctor.HealthStateRunning {
		t.Fatalf("expected repaired health state, got %s", recovered.HealthState)
	}
	if recovered.LockActive != true {
		t.Fatalf("expected active lock after repair")
	}
	if err := lock.Release(daemon.StateRunning); err != nil {
		t.Fatalf("release running lock: %v", err)
	}
	if recovered.NextAction != "lock held by repair-client" {
		t.Fatalf("unexpected post-repair next action: %s", recovered.NextAction)
	}
	if recovered.HealthSummary != "daemon lock held by repair-client" {
		t.Fatalf("unexpected health summary: %s", recovered.HealthSummary)
	}
}

type scenarioState struct {
	mu             sync.Mutex
	readActive     int
	maxReadActive  int
	writeActive    int
	maxWriteActive int
	readStarted    chan struct{}
	writeStarted   chan struct{}
	readRelease    chan struct{}
	writeRelease   chan struct{}
}

func newScenarioState() *scenarioState {
	return &scenarioState{
		readStarted:  make(chan struct{}, 2),
		writeStarted: make(chan struct{}, 2),
		readRelease:  make(chan struct{}),
		writeRelease: make(chan struct{}),
	}
}

func (s *scenarioState) waitReadStarts(count int) {
	for i := 0; i < count; i++ {
		<-s.readStarted
	}
}

func (s *scenarioState) maxConcurrentReads() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.maxReadActive
}

func (s *scenarioState) maxConcurrentWrites() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.maxWriteActive
}

func (s *scenarioState) noteReadStart() {
	s.mu.Lock()
	s.readActive++
	if s.readActive > s.maxReadActive {
		s.maxReadActive = s.readActive
	}
	s.mu.Unlock()
	s.readStarted <- struct{}{}
}

func (s *scenarioState) noteReadEnd() {
	s.mu.Lock()
	s.readActive--
	s.mu.Unlock()
}

func (s *scenarioState) noteWriteStart() {
	s.mu.Lock()
	s.writeActive++
	if s.writeActive > s.maxWriteActive {
		s.maxWriteActive = s.writeActive
	}
	s.mu.Unlock()
	s.writeStarted <- struct{}{}
}

func (s *scenarioState) noteWriteEnd() {
	s.mu.Lock()
	s.writeActive--
	s.mu.Unlock()
}

type scenarioLocalClient struct {
	state        *scenarioState
	readRelease  chan struct{}
	writeRelease chan struct{}
	writeMu      sync.Mutex
}

func (c *scenarioLocalClient) Send(ctx context.Context, token localapi.SessionToken, payload string) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	c.state.noteWriteStart()
	defer c.state.noteWriteEnd()
	<-c.writeRelease
	return nil
}

func (c *scenarioLocalClient) Read(ctx context.Context, token localapi.SessionToken, cursor string) (string, error) {
	c.state.noteReadStart()
	defer c.state.noteReadEnd()
	<-c.readRelease
	return fmt.Sprintf("read:%s", cursor), nil
}

func (c *scenarioLocalClient) Search(ctx context.Context, token localapi.SessionToken, query string) ([]string, error) {
	return []string{"match"}, nil
}

func (c *scenarioLocalClient) FetchHistory(ctx context.Context, token localapi.SessionToken, req journeys.HistoryRequest) ([]string, error) {
	return []string{"history"}, nil
}

func (c *scenarioLocalClient) MediaControl(ctx context.Context, token localapi.SessionToken, action string) error {
	return nil
}

type fixedAttachProvider struct {
	token localapi.SessionToken
}

func (f *fixedAttachProvider) Attach(ctx context.Context) attach.AttachResult {
	return attach.AttachResult{Attached: true, Token: f.token}
}

type sequenceHandshaker struct {
	tokens []localapi.SessionToken
	calls  int
	mu     sync.Mutex
}

func (s *sequenceHandshaker) Handshake(ctx context.Context) (localapi.SessionToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.calls
	s.calls++
	if idx < len(s.tokens) {
		return s.tokens[idx], nil
	}
	return localapi.SessionToken{Value: fmt.Sprintf("token-%d", idx+1), ExpiresAt: time.Now().Add(time.Minute)}, nil
}

type crashyLocalClient struct {
	mu      sync.Mutex
	crashes int
}

func (c *crashyLocalClient) Send(ctx context.Context, token localapi.SessionToken, payload string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.crashes > 0 {
		c.crashes--
		return errDaemonCrashed
	}
	return nil
}

func (c *crashyLocalClient) Read(ctx context.Context, token localapi.SessionToken, cursor string) (string, error) {
	return "read", nil
}

func (c *crashyLocalClient) Search(ctx context.Context, token localapi.SessionToken, query string) ([]string, error) {
	return []string{"result"}, nil
}

func (c *crashyLocalClient) FetchHistory(ctx context.Context, token localapi.SessionToken, req journeys.HistoryRequest) ([]string, error) {
	return []string{"history"}, nil
}

func (c *crashyLocalClient) MediaControl(ctx context.Context, token localapi.SessionToken, action string) error {
	return nil
}

type runningProbe struct{}

func (r *runningProbe) Probe(ctx context.Context) (attach.DaemonStatus, error) {
	return attach.DaemonStatus{Running: true}, nil
}

func (r *runningProbe) WaitReady(ctx context.Context) error {
	return nil
}

var errDaemonCrashed = errors.New("daemon crashed during call")
