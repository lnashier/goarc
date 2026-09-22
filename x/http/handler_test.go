package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONHandler_Success(t *testing.T) {
	h := JSONHandler(func(*http.Request) (any, error) {
		return map[string]string{"hello": "world"}, nil
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["hello"] != "world" {
		t.Fatalf("body = %v, want hello=world", body)
	}
}

func TestJSONHandler_HandlerErrorWritesErrorStatus(t *testing.T) {
	h := JSONHandler(func(*http.Request) (any, error) {
		return nil, NotFound(errors.New("missing"))
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestJSONHandler_MarshalErrorDoesNotWrite200(t *testing.T) {
	// A channel cannot be marshaled to JSON; the handler must catch this
	// before it has already committed a 200 status.
	h := JSONHandler(func(*http.Request) (any, error) {
		return make(chan int), nil
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code == http.StatusOK {
		t.Fatalf("status = 200, want a non-2xx status for an unmarshalable response")
	}
	if rec.Body.Len() == 0 {
		t.Fatal("body is empty, want an error body describing the marshal failure")
	}
}

func TestTextHandler_Success(t *testing.T) {
	h := TextHandler(func(*http.Request) (string, error) {
		return "hello", nil
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "hello" {
		t.Fatalf("body = %q, want %q", rec.Body.String(), "hello")
	}
}

func TestTextHandler_Error(t *testing.T) {
	h := TextHandler(func(*http.Request) (string, error) {
		return "", errors.New("boom")
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

func TestXMLHandler_Success(t *testing.T) {
	type payload struct {
		Name string `xml:"name"`
	}
	h := XMLHandler(func(*http.Request) (any, error) {
		return payload{Name: "world"}, nil
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/xml; charset=UTF-8" {
		t.Fatalf("Content-Type = %q, want application/xml", ct)
	}
}

func TestXMLHandler_Error(t *testing.T) {
	h := XMLHandler(func(*http.Request) (any, error) {
		return nil, errors.New("boom")
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

func TestXMLHandler_MarshalErrorDoesNotWrite200(t *testing.T) {
	h := XMLHandler(func(*http.Request) (any, error) {
		return make(chan int), nil
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code == http.StatusOK {
		t.Fatalf("status = 200, want a non-2xx status for an unmarshalable response")
	}
}
