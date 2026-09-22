package http

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestDefaultRetryPolicy(t *testing.T) {
	cases := []struct {
		name      string
		resp      *http.Response
		err       error
		wantRetry bool
	}{
		{"connection error", nil, errors.New("dial failed"), true},
		{"zero status", &http.Response{StatusCode: 0}, nil, true},
		{"500", &http.Response{StatusCode: 500}, nil, true},
		{"503", &http.Response{StatusCode: 503}, nil, true},
		{"200", &http.Response{StatusCode: 200}, nil, false},
		{"404", &http.Response{StatusCode: 404}, nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			retry, _ := DefaultRetryPolicy(c.resp, c.err)
			if retry != c.wantRetry {
				t.Fatalf("DefaultRetryPolicy() = %v, want %v", retry, c.wantRetry)
			}
		})
	}
}

func TestDefaultBackoffPolicy_ExponentialWithinBounds(t *testing.T) {
	min := 1 * time.Second
	max := 30 * time.Second

	prev := time.Duration(0)
	for attempt := 0; attempt < 4; attempt++ {
		d := DefaultBackoffPolicy(min, max, attempt, nil)
		if d < prev {
			t.Fatalf("attempt %d: backoff %v is less than previous %v, want non-decreasing", attempt, d, prev)
		}
		if d > max {
			t.Fatalf("attempt %d: backoff %v exceeds max %v", attempt, d, max)
		}
		prev = d
	}
}

func TestDefaultBackoffPolicy_SaturatesAtMax(t *testing.T) {
	min := 1 * time.Second
	max := 10 * time.Second

	d := DefaultBackoffPolicy(min, max, 20, nil) // 2^20 * 1s is enormous
	if d != max {
		t.Fatalf("backoff = %v, want it saturated at max %v", d, max)
	}
}

func TestDefaultRetry_HasSaneDefaults(t *testing.T) {
	r := DefaultRetry()
	if r.Max != 3 {
		t.Fatalf("Max = %d, want 3", r.Max)
	}
	if r.WaitMin <= 0 || r.WaitMax <= r.WaitMin {
		t.Fatalf("WaitMin=%v WaitMax=%v, want 0 < WaitMin < WaitMax", r.WaitMin, r.WaitMax)
	}
	if r.Policy == nil || r.Backoff == nil || r.OnTry == nil {
		t.Fatal("DefaultRetry() left a nil Policy/Backoff/OnTry")
	}
}

func TestNoRetry_NeverRetries(t *testing.T) {
	r := noRetry()
	retry, err := r.Policy(&http.Response{StatusCode: 503}, nil)
	if retry {
		t.Fatal("noRetry().Policy() = retry true, want false")
	}
	if err != nil {
		t.Fatalf("noRetry().Policy() = err %v, want nil for a nil input error", err)
	}
}
