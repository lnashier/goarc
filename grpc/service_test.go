package grpc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

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

func TestService_RegisterServiceAndServe(t *testing.T) {
	svc := NewService(ServicePort(0))
	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(svc, healthSrv)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- svc.Start(ctx) }()

	addr := waitForAddr(t, svc)

	conn, err := googlegrpc.NewClient(addr, googlegrpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	client := healthpb.NewHealthClient(conn)
	reqCtx, reqCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer reqCancel()

	resp, err := client.Check(reqCtx, &healthpb.HealthCheckRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != healthpb.HealthCheckResponse_SERVING {
		t.Fatalf("status = %v, want SERVING", resp.Status)
	}

	cancel()
	if err := waitDone(t, done); err != nil {
		t.Fatalf("Start() = %v, want nil", err)
	}
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

func TestService_Stop_ForcesClosedWhenGracefulStopExceedsDeadline(t *testing.T) {
	svc := NewService(ServicePort(0))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- svc.Start(ctx) }()
	addr := waitForAddr(t, svc)

	// Open a long-lived streaming-ish connection to keep GracefulStop from
	// finishing on its own: a plain client connection is enough to make
	// GracefulStop wait, since it waits for all client connections/RPCs to
	// finish or be explicitly closed.
	conn, err := googlegrpc.NewClient(addr, googlegrpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer stopCancel()

	start := time.Now()
	_ = svc.Stop(stopCtx)
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		t.Fatalf("Stop() took %v, want it bounded by the ~100ms stop deadline (GracefulStop must not hang indefinitely)", elapsed)
	}

	cancel()
	_ = waitDone(t, done)
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

func TestService_App_ErrorPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewService() with a failing App did not panic")
		}
	}()
	NewService(App(func(*Service) error {
		return fmt.Errorf("bad config")
	}))
}

func TestServiceNetwork_IsApplied(t *testing.T) {
	var got string
	NewService(ServiceNetwork("tcp4"), App(func(s *Service) error {
		got = s.opts.network
		return nil
	}))
	if got != "tcp4" {
		t.Fatalf("opts.network = %q, want %q", got, "tcp4")
	}
}

func TestService_Addr_EmptyBeforeStart(t *testing.T) {
	svc := NewService(ServicePort(0))
	if svc.Addr() != "" {
		t.Fatalf("Addr() before Start = %q, want empty", svc.Addr())
	}
}
