package buildinfo_test

import (
	"context"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"testing"
	"time"

	goarchttp "github.com/lnashier/goarc/v2/http"
	"github.com/lnashier/goarc/v2/x/buildinfo"
)

func TestApp_RegistersBuildinfoEndpoint(t *testing.T) {
	svc := goarchttp.NewService(goarchttp.ServicePort(0), goarchttp.App(buildinfo.App))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- svc.Start(ctx) }()

	addr := waitForAddr(t, svc)

	resp, err := nethttp.Get(fmt.Sprintf("http://%s/buildinfo", addr))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != nethttp.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["version"]; !ok {
		t.Fatalf("response missing version key: %v", body)
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
