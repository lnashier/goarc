package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewRequest_SetsContentLength(t *testing.T) {
	c := NewClient()
	body := bytes.NewReader([]byte("hello world"))

	req, err := c.NewRequest(context.Background(), http.MethodPost, "/x", http.Header{}, body)
	if err != nil {
		t.Fatal(err)
	}
	if req.ContentLength != int64(len("hello world")) {
		t.Fatalf("ContentLength = %d, want %d", req.ContentLength, len("hello world"))
	}
}

type failingReadSeeker struct{}

func (failingReadSeeker) Read([]byte) (int, error)       { return 0, errors.New("read failed") }
func (failingReadSeeker) Seek(int64, int) (int64, error) { return 0, nil }

func TestNewRequest_PropagatesBodyMeasureError(t *testing.T) {
	c := NewClient()
	_, err := c.NewRequest(context.Background(), http.MethodPost, "/x", http.Header{}, failingReadSeeker{})
	if err == nil {
		t.Fatal("NewRequest() = nil error, want one wrapping the read failure")
	}
}

func TestClient_Do_SuccessNoRetry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient(WithHost(srv.URL))
	resp, err := c.Get(context.Background(), "/", http.Header{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

func TestClient_Do_RetriesOn5xxThenSucceeds(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient(WithHost(srv.URL))
	retry := &Retry{
		WaitMin: time.Millisecond,
		WaitMax: 5 * time.Millisecond,
		Max:     3,
		Policy:  DefaultRetryPolicy,
		Backoff: DefaultBackoffPolicy,
		OnTry:   func(time.Duration, int, error) {},
	}

	resp, err := c.Get(context.Background(), "/", http.Header{}, retry)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want 200", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("server was called %d times, want 3", got)
	}
}

func TestClient_Do_GivesUpAfterMaxRetries(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := NewClient(WithHost(srv.URL))
	retry := &Retry{
		WaitMin: time.Millisecond,
		WaitMax: 2 * time.Millisecond,
		Max:     2,
		Policy:  DefaultRetryPolicy,
		Backoff: DefaultBackoffPolicy,
		OnTry:   func(time.Duration, int, error) {},
	}

	_, err := c.Get(context.Background(), "/", http.Header{}, retry)
	if err == nil {
		t.Fatal("Get() = nil error, want one after exhausting retries")
	}
	if got := atomic.LoadInt32(&calls); got != 3 { // 1 initial + 2 retries
		t.Fatalf("server was called %d times, want 3", got)
	}
}

func TestClient_Do_NoRetryByDefault(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := NewClient(WithHost(srv.URL))
	resp, err := c.Get(context.Background(), "/", http.Header{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("StatusCode = %d, want 503", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("server was called %d times, want 1 (nil Retry means no retries)", got)
	}
}

type result struct {
	Name string `json:"name"`
}

func TestDoDecoded_SuccessOn2xxVariants(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusCreated, http.StatusAccepted} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_ = json.NewEncoder(w).Encode(result{Name: "ok"})
			}))
			defer srv.Close()

			c := NewClient(WithHost(srv.URL))
			var got result
			resp, err := c.GetDecoded(context.Background(), "/", http.Header{}, &got, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = resp.Body.Close() }()
			if got.Name != "ok" {
				t.Fatalf("Name = %q, want %q", got.Name, "ok")
			}
		})
	}
}

func TestDoDecoded_NoContentIsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(WithHost(srv.URL))
	var got result
	resp, err := c.GetDecoded(context.Background(), "/", http.Header{}, &got, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
}

func TestDoDecoded_EmptyBodyErrorStatusIsStillAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound) // no body written
	}))
	defer srv.Close()

	c := NewClient(WithHost(srv.URL))
	var got result
	_, err := c.GetDecoded(context.Background(), "/", http.Header{}, &got, nil)
	if err == nil {
		t.Fatal("GetDecoded() = nil error, want an error for a 404 even with an empty body")
	}
	status, ok := Is4xx(err)
	if !ok || status != http.StatusNotFound {
		t.Fatalf("Is4xx(err) = (%d, %v), want (404, true)", status, ok)
	}
}

func TestDoDecoded_DecodeErrorIsWrapped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := NewClient(WithHost(srv.URL))
	var got result
	_, err := c.GetDecoded(context.Background(), "/", http.Header{}, &got, nil)
	if err == nil {
		t.Fatal("GetDecoded() = nil error, want a decode error")
	}
}

func TestPostEncodedDecoded_RoundTrips(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in result
		_ = json.NewDecoder(r.Body).Decode(&in)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result{Name: "echo:" + in.Name})
	}))
	defer srv.Close()

	c := NewClient(WithHost(srv.URL))
	var got result
	resp, err := c.PostEncodedDecoded(context.Background(), "/", http.Header{}, result{Name: "hi"}, &got, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if got.Name != "echo:hi" {
		t.Fatalf("Name = %q, want %q", got.Name, "echo:hi")
	}
}

func TestClient_Do_RewindsBodyOnRetry(t *testing.T) {
	var bodies []string
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient(WithHost(srv.URL))
	retry := &Retry{
		WaitMin: time.Millisecond,
		WaitMax: 2 * time.Millisecond,
		Max:     1,
		Policy:  DefaultRetryPolicy,
		Backoff: DefaultBackoffPolicy,
		OnTry:   func(time.Duration, int, error) {},
	}

	resp, err := c.Post(context.Background(), "/", http.Header{}, bytes.NewReader([]byte("payload")), retry)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if len(bodies) != 2 || bodies[0] != "payload" || bodies[1] != "payload" {
		t.Fatalf("bodies seen by server = %v, want [\"payload\" \"payload\"]", bodies)
	}
}

func TestClient_Do_RetryWaitHonorsContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := NewClient(WithHost(srv.URL))
	retry := &Retry{
		WaitMin: time.Hour, // would hang the test if the wait ignored ctx
		WaitMax: time.Hour,
		Max:     3,
		Policy:  DefaultRetryPolicy,
		Backoff: func(min, max time.Duration, attempt int, resp *http.Response) time.Duration { return time.Hour },
		OnTry:   func(time.Duration, int, error) {},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := c.Get(ctx, "/", http.Header{}, retry)
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("Get() took %v, want it to return soon after ctx expiry", elapsed)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Get() error = %v, want it to wrap context.DeadlineExceeded", err)
	}
}
