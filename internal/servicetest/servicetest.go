// Package servicetest runs a shared conformance suite against any
// goarc.Service implementation in this module (http.Service, grpc.Service,
// cli.Service, ...), checking the lifecycle contract documented on
// goarc.Service: Stop is safe before Start, concurrently, or more than
// once, and Start self-terminates when its context is done.
package servicetest

import (
	"context"
	"testing"
	"time"

	"github.com/lnashier/goarc/v2"
)

// Conformance runs the suite as subtests of t. newService must return a
// fresh, not-yet-started Service on every call, configured so that
// starting it (e.g. binding a listener) cannot collide across subtests —
// for a network service, that typically means an OS-assigned port.
func Conformance(t *testing.T, newService func() goarc.Service) {
	t.Helper()

	t.Run("StopBeforeStartIsSafe", func(t *testing.T) {
		svc := newService()
		if err := svc.Stop(context.Background()); err != nil {
			t.Fatalf("Stop() before Start = %v, want nil", err)
		}
	})

	t.Run("StopIsIdempotent", func(t *testing.T) {
		svc := newService()
		first := svc.Stop(context.Background())
		second := svc.Stop(context.Background())
		if second != first {
			t.Fatalf("second Stop() = %v, want the same result as the first Stop() = %v", second, first)
		}
	})

	t.Run("StartSelfTerminatesOnCtxDone", func(t *testing.T) {
		svc := newService()
		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan error, 1)
		go func() { done <- svc.Start(ctx) }()

		time.Sleep(20 * time.Millisecond) // let Start actually begin
		cancel()

		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("Start() = %v, want nil once ctx is canceled", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Start did not return within 5s of ctx being canceled")
		}
	})

	t.Run("ExplicitStopUnblocksStart", func(t *testing.T) {
		svc := newService()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		done := make(chan error, 1)
		go func() { done <- svc.Start(ctx) }()

		time.Sleep(20 * time.Millisecond)
		if err := svc.Stop(context.Background()); err != nil {
			t.Fatalf("Stop() = %v, want nil", err)
		}

		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("Start() = %v, want nil once Stop is called directly", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Start did not return within 5s of Stop being called")
		}
	})

	t.Run("StartAfterPriorStopReturnsPromptly", func(t *testing.T) {
		svc := newService()
		if err := svc.Stop(context.Background()); err != nil {
			t.Fatalf("Stop() before Start = %v, want nil", err)
		}

		done := make(chan error, 1)
		go func() { done <- svc.Start(context.Background()) }()

		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("Start() after a prior Stop = %v, want nil", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Start did not return promptly after a Stop that happened before it")
		}
	})
}
