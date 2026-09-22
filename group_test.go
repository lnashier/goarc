package goarc

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// recSvc is a test double that satisfies the full Service contract: Start
// blocks until either ctx is done or Stop is called (whichever comes
// first), and Stop is safe to call before Start, concurrently, or more
// than once.
type recSvc struct {
	mu       sync.Mutex
	startCnt int
	stopCnt  int
	startErr error // if set, Start returns this immediately instead of blocking
	stopErr  error

	done     chan struct{}
	stopOnce sync.Once
}

func newRecSvc() *recSvc {
	return &recSvc{done: make(chan struct{})}
}

func (s *recSvc) Start(ctx context.Context) error {
	s.mu.Lock()
	s.startCnt++
	err := s.startErr
	s.mu.Unlock()

	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
	case <-s.done:
	}
	return nil
}

func (s *recSvc) Stop(ctx context.Context) error {
	s.mu.Lock()
	s.stopCnt++
	err := s.stopErr
	s.mu.Unlock()

	s.stopOnce.Do(func() { close(s.done) })
	return err
}

func (s *recSvc) startCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.startCnt
}

func (s *recSvc) stopCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopCnt
}

func TestGroup_ImplementsService(t *testing.T) {
	var _ Service = NewGroup()
}

func TestGroup_Empty_StartBlocksUntilCtxDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	g := NewGroup()

	errCh := make(chan error, 1)
	go func() { errCh <- g.Start(ctx) }()

	select {
	case err := <-errCh:
		t.Fatalf("Start returned early with %v before ctx was canceled", err)
	case <-time.After(20 * time.Millisecond):
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Start() = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Start did not return after ctx was canceled")
	}
}

func TestGroup_StartStop_AllMembersRunUntilCanceled(t *testing.T) {
	a, b, c := newRecSvc(), newRecSvc(), newRecSvc()
	g := NewGroup().Add("a", a).Add("b", b).Add("c", c)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- g.Start(ctx) }()

	waitFor(t, func() bool {
		return a.startCount() == 1 && b.startCount() == 1 && c.startCount() == 1
	})

	cancel()
	if err := waitErr(t, errCh); err != nil {
		t.Fatalf("Start() = %v, want nil", err)
	}
}

func TestGroup_Start_OneMemberFailureCascadesStop(t *testing.T) {
	failing := &recSvc{startErr: errors.New("boom"), done: make(chan struct{})}
	sibling := newRecSvc()
	g := NewGroup().Add("failing", failing).Add("sibling", sibling)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := g.Start(ctx)
	if err == nil {
		t.Fatal("Start() = nil, want an error naming the failing member")
	}
	if !strings.Contains(err.Error(), "failing") {
		t.Fatalf("Start() error = %q, want it to name %q", err.Error(), "failing")
	}
	if !errors.Is(err, failing.startErr) {
		t.Fatalf("Start() error does not wrap the original member error: %v", err)
	}
	// The sibling should have been canceled and returned too, without
	// Start ever needing to wait for the outer ctx.
	if got := sibling.startCount(); got != 1 {
		t.Fatalf("sibling.startCount() = %d, want 1", got)
	}
}

func TestGroup_Stop_SafeBeforeStart(t *testing.T) {
	a, b := newRecSvc(), newRecSvc()
	g := NewGroup().Add("a", a).Add("b", b)

	if err := g.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() before Start = %v, want nil", err)
	}
	if a.stopCount() != 1 || b.stopCount() != 1 {
		t.Fatalf("stopCount = (%d, %d), want (1, 1)", a.stopCount(), b.stopCount())
	}

	// Start should now return promptly since Stop already fired.
	errCh := make(chan error, 1)
	go func() { errCh <- g.Start(context.Background()) }()
	if err := waitErr(t, errCh); err != nil {
		t.Fatalf("Start() after prior Stop = %v, want nil", err)
	}
}

func TestGroup_Stop_IdempotentAndConcurrent(t *testing.T) {
	a := newRecSvc()
	g := NewGroup().Add("a", a)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := g.Stop(context.Background()); err != nil {
				t.Errorf("Stop() = %v, want nil", err)
			}
		}()
	}
	wg.Wait()

	if got := a.stopCount(); got != 5 {
		t.Fatalf("member Stop was called %d times, want 5 (each Stop call must reach the member)", got)
	}
}

func TestGroup_Stop_ErrorsJoinedWithMemberNames(t *testing.T) {
	failA := &recSvc{stopErr: errors.New("a broke"), done: make(chan struct{})}
	failB := &recSvc{stopErr: errors.New("b broke"), done: make(chan struct{})}
	ok := newRecSvc()
	g := NewGroup().Add("svc-a", failA).Add("svc-b", failB).Add("svc-ok", ok)

	err := g.Stop(context.Background())
	if err == nil {
		t.Fatal("Stop() = nil, want a joined error")
	}
	for _, want := range []string{"svc-a", "a broke", "svc-b", "b broke"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Stop() error = %q, want it to contain %q", err.Error(), want)
		}
	}
	if !errors.Is(err, failA.stopErr) || !errors.Is(err, failB.stopErr) {
		t.Fatalf("Stop() error does not wrap both member errors: %v", err)
	}
}

func TestGroup_Nested(t *testing.T) {
	inner := newRecSvc()
	innerGroup := NewGroup().Add("inner", inner)
	outer := NewGroup().Add("group", innerGroup)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- outer.Start(ctx) }()

	waitFor(t, func() bool { return inner.startCount() == 1 })

	cancel()
	if err := waitErr(t, errCh); err != nil {
		t.Fatalf("outer.Start() = %v, want nil", err)
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition not met before deadline")
}

func waitErr(t *testing.T, errCh <-chan error) error {
	t.Helper()
	select {
	case err := <-errCh:
		return err
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for result")
		return nil
	}
}
