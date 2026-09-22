package cli

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lnashier/goarc/v2"
	"github.com/lnashier/goarc/v2/internal/servicetest"
)

func TestService_ImplementsGoarcService(t *testing.T) {
	var _ goarc.Service = NewService()
}

func TestService_Conformance(t *testing.T) {
	servicetest.Conformance(t, func() goarc.Service {
		svc := NewService(ServiceArgs([]string{"wait"}))
		svc.Register("wait", func(ctx context.Context, args []string) error {
			<-ctx.Done()
			return nil
		})
		return svc
	})
}

func TestService_Register_RunsAndReceivesArgs(t *testing.T) {
	var gotArgs []string
	svc := NewService(ServiceArgs([]string{"echo", "hello", "world"}))
	svc.Register("echo", func(ctx context.Context, args []string) error {
		gotArgs = args
		return nil
	})

	if err := svc.Start(context.Background()); err != nil {
		t.Fatalf("Start() = %v, want nil", err)
	}
	if len(gotArgs) != 2 || gotArgs[0] != "hello" || gotArgs[1] != "world" {
		t.Fatalf("args = %v, want [hello world]", gotArgs)
	}
}

func TestService_Register_PropagatesRunnerError(t *testing.T) {
	want := errors.New("command failed")
	svc := NewService(ServiceArgs([]string{"fail"}))
	svc.Register("fail", func(ctx context.Context, args []string) error {
		return want
	})

	if err := svc.Start(context.Background()); !errors.Is(err, want) {
		t.Fatalf("Start() = %v, want it to wrap %v", err, want)
	}
}

func TestService_NoCommandSelected_ReturnsDefaultError(t *testing.T) {
	svc := NewService(ServiceArgs([]string{}))
	if err := svc.Start(context.Background()); err == nil {
		t.Fatal("Start() = nil, want an error when no subcommand is selected")
	}
}

func TestService_CtxCancel_CancelsRunningCommand(t *testing.T) {
	svc := NewService(ServiceArgs([]string{"wait"}))
	svc.Register("wait", func(ctx context.Context, args []string) error {
		<-ctx.Done()
		return ctx.Err()
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- svc.Start(ctx) }()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Start() = %v, want it to wrap context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after ctx was canceled")
	}
}

func TestService_Stop_CancelsRunningCommandIndependentlyOfCtx(t *testing.T) {
	svc := NewService(ServiceArgs([]string{"wait"}))
	svc.Register("wait", func(ctx context.Context, args []string) error {
		<-ctx.Done()
		return nil
	})

	// Start with a ctx that never gets canceled on its own: only Stop
	// should be able to unblock the running command.
	done := make(chan error, 1)
	go func() { done <- svc.Start(context.Background()) }()

	time.Sleep(20 * time.Millisecond)
	if err := svc.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() = %v, want nil", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start() = %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after Stop was called")
	}
}

func TestServiceName_IsApplied(t *testing.T) {
	var got string
	NewService(ServiceName("my-cli"), App(func(s *Service) error {
		got = s.opts.name
		return nil
	}))
	if got != "my-cli" {
		t.Fatalf("opts.name = %q, want %q", got, "my-cli")
	}
}

func TestService_App_ErrorPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewService() with a failing App did not panic")
		}
	}()
	NewService(App(func(*Service) error {
		return errors.New("bad config")
	}))
}
