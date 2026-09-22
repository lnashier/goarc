package health

import (
	"context"
	"testing"

	xhttp "github.com/lnashier/goarc/v2/x/http"
)

func TestController_Live_AlwaysUp(t *testing.T) {
	ctr := New()
	if ctr.Live() != "up" {
		t.Fatalf("Live() = %q, want %q", ctr.Live(), "up")
	}
	_ = ctr.Stop(context.Background())
	if ctr.Live() != "up" {
		t.Fatalf("Live() after Stop = %q, want %q (liveness stays up until the process exits)", ctr.Live(), "up")
	}
}

func TestController_Ready_BeforeStop(t *testing.T) {
	ctr := New()
	status, err := ctr.Ready()
	if err != nil {
		t.Fatalf("Ready() = (%q, %v), want nil error", status, err)
	}
	if status != "up" {
		t.Fatalf("Ready() status = %q, want %q", status, "up")
	}
}

func TestController_Ready_AfterStop(t *testing.T) {
	ctr := New()
	if err := ctr.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() = %v, want nil", err)
	}

	status, err := ctr.Ready()
	if err == nil {
		t.Fatal("Ready() after Stop = nil error, want a not-found error")
	}
	if status != "" {
		t.Fatalf("Ready() status = %q, want empty on error", status)
	}
	if s, ok := xhttp.Is4xx(err); !ok || s != 404 {
		t.Fatalf("Ready() error = %v, want a 404", err)
	}
}

func TestController_Stop_IsIdempotent(t *testing.T) {
	ctr := New()
	if err := ctr.Stop(context.Background()); err != nil {
		t.Fatalf("first Stop() = %v, want nil", err)
	}
	if err := ctr.Stop(context.Background()); err != nil {
		t.Fatalf("second Stop() = %v, want nil, not a double-close panic", err)
	}
}
