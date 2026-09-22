package health_test

import (
	"context"
	"fmt"
	nethttp "net/http"
	"testing"
	"time"

	goarchttp "github.com/lnashier/goarc/v2/http"
	"github.com/lnashier/goarc/v2/x/health"
	xhttp "github.com/lnashier/goarc/v2/x/http"
)

func TestApp_WiresAliveAndReady(t *testing.T) {
	svc := goarchttp.NewService(goarchttp.ServicePort(0), goarchttp.App(health.App))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- svc.Start(ctx) }()

	addr := waitForAddr(t, svc)
	base := fmt.Sprintf("http://%s", addr)

	assertStatus(t, base+"/alive", nethttp.StatusOK)
	assertStatus(t, base+"/ready", nethttp.StatusOK)
}

// TestController_Ready_FlipsAfterStopWhileServiceKeepsServing wires the
// same routes health.App does, but keeps a direct reference to the
// Controller so it can be stopped independently of the whole http.Service
// — App itself has no way to observe this intermediate state via HTTP,
// since stopping the Service also shuts down its listener.
func TestController_Ready_FlipsAfterStopWhileServiceKeepsServing(t *testing.T) {
	svc := goarchttp.NewService(goarchttp.ServicePort(0))
	ctr := health.New()
	svc.Component(ctr)
	svc.Register("/alive", nethttp.MethodGet, xhttp.TextHandler(func(*nethttp.Request) (string, error) {
		return ctr.Live(), nil
	}))
	svc.Register("/ready", nethttp.MethodGet, xhttp.TextHandler(func(*nethttp.Request) (string, error) {
		return ctr.Ready()
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- svc.Start(ctx) }()

	addr := waitForAddr(t, svc)
	base := fmt.Sprintf("http://%s", addr)

	assertStatus(t, base+"/ready", nethttp.StatusOK)

	if err := ctr.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() = %v, want nil", err)
	}

	assertStatus(t, base+"/alive", nethttp.StatusOK)
	assertStatus(t, base+"/ready", nethttp.StatusNotFound)
}

func assertStatus(t *testing.T, url string, want int) {
	t.Helper()
	resp, err := nethttp.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != want {
		t.Fatalf("GET %s = %d, want %d", url, resp.StatusCode, want)
	}
}

func waitForAddr(t *testing.T, svc *goarchttp.Service) string {
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
