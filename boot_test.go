package goarc

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestRun_ContextCancelTriggersGracefulStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	started := make(chan struct{})
	svc := Func{
		StartFunc: func(ctx context.Context) error {
			close(started)
			<-ctx.Done()
			return nil
		},
	}

	errCh := make(chan error, 1)
	go func() { errCh <- Run(svc, Context(ctx)) }()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("Start was never called")
	}

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Run() = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not return after ctx was canceled")
	}
}

func TestRun_AlwaysCallsStopEvenWhenStartFailsImmediately(t *testing.T) {
	startErr := errors.New("bind failed")
	stopCalled := false
	svc := Func{
		StartFunc: func(context.Context) error { return startErr },
		StopFunc: func(context.Context) error {
			stopCalled = true
			return nil
		},
	}

	err := Run(svc, Context(context.Background()))
	if !errors.Is(err, startErr) {
		t.Fatalf("Run() = %v, want it to wrap %v", err, startErr)
	}
	if !stopCalled {
		t.Fatal("Stop was not called after Start failed — cleanup was skipped")
	}
}

func TestRun_JoinsStartAndStopErrors(t *testing.T) {
	startErr := errors.New("start failed")
	stopErr := errors.New("stop failed")
	svc := Func{
		StartFunc: func(context.Context) error { return startErr },
		StopFunc:  func(context.Context) error { return stopErr },
	}

	err := Run(svc, Context(context.Background()))
	if !errors.Is(err, startErr) || !errors.Is(err, stopErr) {
		t.Fatalf("Run() = %v, want it to wrap both %v and %v", err, startErr, stopErr)
	}
}

func TestRun_HooksAreObservabilityOnly(t *testing.T) {
	startErr := errors.New("start failed")
	var gotStart, gotStop error
	svc := Func{
		StartFunc: func(context.Context) error { return startErr },
	}

	err := Run(svc,
		Context(context.Background()),
		OnStart(func(e error) { gotStart = e }),
		OnStop(func(e error) { gotStop = e }),
	)

	if !errors.Is(gotStart, startErr) {
		t.Fatalf("OnStart hook got %v, want %v", gotStart, startErr)
	}
	if gotStop != nil {
		t.Fatalf("OnStop hook got %v, want nil (StopFunc unset)", gotStop)
	}
	// The hooks ran and observed the errors, but Run's own return value is
	// still the source of truth for both phases.
	if !errors.Is(err, startErr) {
		t.Fatalf("Run() = %v, want it to still wrap %v despite the hooks", err, startErr)
	}
}

func TestRun_ShutdownTimeoutBoundsStopContext(t *testing.T) {
	var gotDeadline bool
	var remaining time.Duration
	svc := Func{
		StartFunc: func(ctx context.Context) error { <-ctx.Done(); return nil },
		StopFunc: func(ctx context.Context) error {
			dl, ok := ctx.Deadline()
			gotDeadline = ok
			remaining = time.Until(dl)
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already canceled: Start returns immediately, Stop runs right after

	if err := Run(svc, Context(ctx), ShutdownTimeout(5*time.Second)); err != nil {
		t.Fatalf("Run() = %v, want nil", err)
	}
	if !gotDeadline {
		t.Fatal("Stop's ctx had no deadline, want one derived from ShutdownTimeout")
	}
	if remaining <= 0 || remaining > 5*time.Second {
		t.Fatalf("Stop's ctx deadline had %v remaining, want (0, 5s]", remaining)
	}
}

func TestRun_DefaultShutdownTimeoutIsTenSeconds(t *testing.T) {
	if defaultBootOpts.shutdownTimeout != 10*time.Second {
		t.Fatalf("default shutdownTimeout = %v, want 10s", defaultBootOpts.shutdownTimeout)
	}
}

func TestUp_ReturnsNormallyOnSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	svc := Func{StartFunc: func(ctx context.Context) error { <-ctx.Done(); return nil }}

	done := make(chan struct{})
	go func() {
		Up(svc, Context(ctx))
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Up did not return")
	}
}

// TestUp_ExitsNonZeroOnError re-execs this same test as a subprocess (the
// standard pattern for exercising a code path that calls os.Exit) and
// checks that it exits with status 1 when Start fails.
func TestUp_ExitsNonZeroOnError(t *testing.T) {
	if os.Getenv("GOARC_TEST_UP_SUBPROCESS") == "1" {
		Up(Func{StartFunc: func(context.Context) error { return errors.New("boom") }})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestUp_ExitsNonZeroOnError")
	cmd.Env = append(os.Environ(), "GOARC_TEST_UP_SUBPROCESS=1")
	err := cmd.Run()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("subprocess error = %v, want *exec.ExitError", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Fatalf("subprocess exit code = %d, want 1", exitErr.ExitCode())
	}
}
