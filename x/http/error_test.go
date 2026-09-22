package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConvertError_Nil(t *testing.T) {
	if got := ConvertError(nil); got != nil {
		t.Fatalf("ConvertError(nil) = %v, want nil", got)
	}
}

func TestConvertError_PassesThroughError(t *testing.T) {
	e := NotFound(nil).(*Error)
	if got := ConvertError(e); got != e {
		t.Fatalf("ConvertError(*Error) = %v, want the same instance back", got)
	}
}

func TestConvertError_WrapsPlainError(t *testing.T) {
	got := ConvertError(errors.New("boom"))
	if got.Status != http.StatusInternalServerError {
		t.Fatalf("Status = %d, want 500", got.Status)
	}
}

func TestConvertError_UnwrapsWrappedError(t *testing.T) {
	inner := NotFound(nil).(*Error)
	wrapped := fmt.Errorf("context: %w", inner)

	got := ConvertError(wrapped)
	if got != inner {
		t.Fatalf("ConvertError(wrapped) = %v, want the original *Error unwrapped via errors.As", got)
	}
}

func TestError_Error_WithAndWithoutCause(t *testing.T) {
	withCause := NewError(400, "bad", errors.New("root cause"))
	if withCause.Error() != "400 bad caused by root cause" {
		t.Fatalf("Error() = %q", withCause.Error())
	}

	withoutCause := NewError(400, "bad", nil)
	if withoutCause.Error() != "400 bad" {
		t.Fatalf("Error() = %q", withoutCause.Error())
	}
}

func TestError_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	e := NewError(400, "bad", cause)
	if !errors.Is(e, cause) {
		t.Fatal("errors.Is(e, cause) = false, want true via Unwrap")
	}
}

func TestNewErrorf_FormatsMessage(t *testing.T) {
	e := NewErrorf(http.StatusBadRequest, nil, "invalid id %d", 7)
	if e.Message != "invalid id 7" {
		t.Fatalf("Message = %q, want %q", e.Message, "invalid id 7")
	}
}

func TestError_String(t *testing.T) {
	e := NewError(400, "bad", nil)
	if e.String() != e.Error() {
		t.Fatalf("String() = %q, want it to match Error() = %q", e.String(), e.Error())
	}
}

func TestError_WriteJSON(t *testing.T) {
	e := NewError(http.StatusTeapot, "teapot", nil)
	rec := httptest.NewRecorder()
	if _, err := e.WriteJSON(rec); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want 418", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=UTF-8" {
		t.Fatalf("Content-Type = %q", ct)
	}
}

func TestError_WriteText(t *testing.T) {
	e := NewError(http.StatusTeapot, "teapot", nil)
	rec := httptest.NewRecorder()
	if _, err := e.WriteText(rec); err != nil {
		t.Fatal(err)
	}
	if rec.Body.String() != "teapot" {
		t.Fatalf("body = %q, want %q", rec.Body.String(), "teapot")
	}
}

func TestIs4xx(t *testing.T) {
	if status, ok := Is4xx(NotFound(nil)); !ok || status != http.StatusNotFound {
		t.Fatalf("Is4xx(NotFound) = (%d, %v), want (404, true)", status, ok)
	}
	if status, ok := Is4xx(Internal(nil)); ok {
		t.Fatalf("Is4xx(Internal) = (%d, %v), want ok=false for a 5xx", status, ok)
	}
	if _, ok := Is4xx(errors.New("plain")); ok {
		t.Fatal("Is4xx(plain error) = ok=true, want false")
	}
	wrapped := fmt.Errorf("context: %w", BadRequest(nil))
	if status, ok := Is4xx(wrapped); !ok || status != http.StatusBadRequest {
		t.Fatalf("Is4xx(wrapped) = (%d, %v), want (400, true) via errors.As", status, ok)
	}
}

func TestConstructors_SetExpectedStatus(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"NotFound", NotFound(nil), http.StatusNotFound},
		{"NotFoundf", NotFoundf(nil, "id %d", 1), http.StatusNotFound},
		{"Conflict", Conflict(nil), http.StatusConflict},
		{"Conflictf", Conflictf(nil, "dup %d", 1), http.StatusConflict},
		{"BadRequest", BadRequest(nil), http.StatusBadRequest},
		{"BadRequestf", BadRequestf(nil, "bad %d", 1), http.StatusBadRequest},
		{"UnprocessableEntity", UnprocessableEntity(nil), http.StatusUnprocessableEntity},
		{"UnprocessableEntityf", UnprocessableEntityf(nil, "x %d", 1), http.StatusUnprocessableEntity},
		{"PreconditionFailed", PreconditionFailed(nil), http.StatusPreconditionFailed},
		{"PreconditionFailedf", PreconditionFailedf(nil, "x %d", 1), http.StatusPreconditionFailed},
		{"PreconditionRequired", PreconditionRequired(nil), http.StatusPreconditionRequired},
		{"PreconditionRequiredf", PreconditionRequiredf(nil, "x %d", 1), http.StatusPreconditionRequired},
		{"Internal", Internal(nil), http.StatusInternalServerError},
		{"Internalf", Internalf(nil, "x %d", 1), http.StatusInternalServerError},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := c.err.(*Error)
			if e.Status != c.want {
				t.Fatalf("Status = %d, want %d", e.Status, c.want)
			}
		})
	}
}
