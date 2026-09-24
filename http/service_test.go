package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	nethttp "net/http"
	"strings"
	"sync"
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
		return NewService(ServicePort(0))
	})
}

func TestService_RegisterAndServe(t *testing.T) {
	svc := NewService(ServicePort(0))
	svc.Register("/hello", nethttp.MethodGet, nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.WriteHeader(nethttp.StatusOK)
		_, _ = w.Write([]byte("world"))
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- svc.Start(ctx) }()

	addr := waitForAddr(t, svc)

	resp, err := nethttp.Get(fmt.Sprintf("http://%s/hello", addr))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "world" {
		t.Fatalf("body = %q, want %q", body, "world")
	}

	cancel()
	if err := waitDone(t, done); err != nil {
		t.Fatalf("Start() = %v, want nil", err)
	}
}

func TestService_RegisterPanicsOnMalformedPath(t *testing.T) {
	svc := NewService(ServicePort(0))

	defer func() {
		if recover() == nil {
			t.Fatal("Register() with an unbalanced route pattern did not panic")
		}
	}()
	svc.Register("/x/{id:[0-9", nethttp.MethodGet, nethttp.HandlerFunc(func(nethttp.ResponseWriter, *nethttp.Request) {}))
}

func TestService_App_ConfiguresService(t *testing.T) {
	called := false
	svc := NewService(ServicePort(0), App(func(s *Service) error {
		called = true
		s.Register("/from-app", nethttp.MethodGet, nethttp.HandlerFunc(func(nethttp.ResponseWriter, *nethttp.Request) {}))
		return nil
	}))
	if !called {
		t.Fatal("App func was never called by NewService")
	}
	_ = svc
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

type recComponent struct {
	mu      sync.Mutex
	calls   int
	stopErr error
}

func (c *recComponent) Stop(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	return c.stopErr
}

func (c *recComponent) callCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

func TestService_Stop_CallsComponents(t *testing.T) {
	svc := NewService(ServicePort(0))
	comp := &recComponent{}
	svc.Component(comp)

	if err := svc.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() = %v, want nil", err)
	}
	if comp.callCount() != 1 {
		t.Fatalf("component Stop called %d times, want 1", comp.callCount())
	}
}

func TestService_Stop_JoinsComponentErrors(t *testing.T) {
	svc := NewService(ServicePort(0))
	svc.Component(&recComponent{stopErr: errors.New("comp-a failed")})
	svc.Component(&recComponent{stopErr: errors.New("comp-b failed")})

	err := svc.Stop(context.Background())
	if err == nil {
		t.Fatal("Stop() = nil, want a joined error from both failing components")
	}
	if !strings.Contains(err.Error(), "comp-a failed") || !strings.Contains(err.Error(), "comp-b failed") {
		t.Fatalf("Stop() error = %v, want it to mention both component failures", err)
	}
}

func TestService_CtxCancel_StopsComponentsWithinGracetime(t *testing.T) {
	svc := NewService(ServicePort(0), ServiceShutdownGracetime(2*time.Second))
	comp := &recComponent{}
	svc.Component(comp)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- svc.Start(ctx) }()

	waitForAddr(t, svc)
	cancel()

	if err := waitDone(t, done); err != nil {
		t.Fatalf("Start() = %v, want nil", err)
	}
	if comp.callCount() != 1 {
		t.Fatalf("component Stop called %d times after ctx cancellation, want 1", comp.callCount())
	}
}

func TestServiceName_IsApplied(t *testing.T) {
	var got string
	NewService(ServiceName("my-svc"), App(func(s *Service) error {
		got = s.opts.name
		return nil
	}))
	if got != "my-svc" {
		t.Fatalf("opts.name = %q, want %q", got, "my-svc")
	}
}

func TestService_Addr_EmptyBeforeStart(t *testing.T) {
	svc := NewService(ServicePort(0))
	if svc.Addr() != "" {
		t.Fatalf("Addr() before Start = %q, want empty", svc.Addr())
	}
}

func waitForAddr(t *testing.T, svc *Service) string {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if addr := svc.Addr(); addr != "" {
			return addr
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("Addr() never became non-empty after Start")
	return ""
}

func waitDone(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return in time")
		return nil
	}
}
