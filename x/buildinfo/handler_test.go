package buildinfo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandler_ServeHTTP_DefaultReport(t *testing.T) {
	h := New()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/buildinfo", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=UTF-8" {
		t.Fatalf("Content-Type = %q", ct)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"host", "version", "hash", "startTime", "uptime"} {
		if _, ok := body[key]; !ok {
			t.Errorf("report missing key %q: %v", key, body)
		}
	}
	if body["version"] != Version {
		t.Errorf("version = %v, want %v", body["version"], Version)
	}
}

func TestHandler_StartTime_IsUTCRFC3339(t *testing.T) {
	h := New()
	report := h.report()

	ts, ok := report[KeyStartTime].(string)
	if !ok {
		t.Fatalf("startTime = %v (%T), want a string", report[KeyStartTime], report[KeyStartTime])
	}
	if !strings.HasSuffix(ts, "Z") {
		t.Fatalf("startTime = %q, want it to end in Z (UTC)", ts)
	}
	if _, err := time.Parse(time.RFC3339, ts); err != nil {
		t.Fatalf("startTime = %q is not valid RFC3339: %v", ts, err)
	}
}

func TestHandler_CustomReporter_OverridesKeys(t *testing.T) {
	h := New(func() Report {
		return Report{
			KeyVersion: "custom-version",
			"extra":    "value",
		}
	})

	report := h.report()
	if report[KeyVersion] != "custom-version" {
		t.Fatalf("version = %v, want it overridden by the custom Reporter", report[KeyVersion])
	}
	if report["extra"] != "value" {
		t.Fatalf("extra = %v, want %q", report["extra"], "value")
	}
	// Built-in keys not touched by the custom Reporter survive.
	if _, ok := report[KeyHost]; !ok {
		t.Fatal("host key missing after custom Reporter ran")
	}
}

func TestHandler_NilReporterReturn_DoesNotPanic(t *testing.T) {
	h := New(func() Report { return nil })
	report := h.report()
	if report[KeyVersion] != Version {
		t.Fatalf("version = %v, want default %v when custom Reporter returns nil", report[KeyVersion], Version)
	}
}

func TestHandler_Uptime_Increases(t *testing.T) {
	h := New()
	first := h.report()[KeyUptime].(int64)

	time.Sleep(1100 * time.Millisecond)

	second := h.report()[KeyUptime].(int64)
	if second <= first {
		t.Fatalf("uptime did not increase: first=%d second=%d", first, second)
	}
}
