package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type signup struct {
	Email string `json:"email" validate:"required"`
	Name  string `json:"name"`
	age   int    `validate:"required"` //nolint:unused // unexported: must never be inspected
}

func (s *signup) Validate(*http.Request) error { return nil }

func TestRequestDecode_ParsesBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"a@b.com"}`))
	var s signup
	if err := RequestDecode(req, &s); err != nil {
		t.Fatal(err)
	}
	if s.Email != "a@b.com" {
		t.Fatalf("Email = %q, want %q", s.Email, "a@b.com")
	}
}

func TestRequestValidate_MissingRequiredField(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	s := &signup{}
	if err := RequestValidate(req, s); err == nil {
		t.Fatal("RequestValidate() = nil, want an error for missing required Email")
	}
}

func TestRequestValidate_RequiredFieldPresent(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	s := &signup{Email: "a@b.com"}
	if err := RequestValidate(req, s); err != nil {
		t.Fatalf("RequestValidate() = %v, want nil", err)
	}
}

func TestRequestValidate_SkipsCheckOnGet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	s := &signup{} // missing Email, but GET is not checked
	if err := RequestValidate(req, s); err != nil {
		t.Fatalf("RequestValidate() = %v, want nil (GET skips required-field check)", err)
	}
}

func TestRequestValidate_UnexportedFieldNeverPanics(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	s := &signup{Email: "a@b.com"} // age is zero and unexported+required, must be ignored
	if err := RequestValidate(req, s); err != nil {
		t.Fatalf("RequestValidate() = %v, want nil (unexported fields are skipped)", err)
	}
}

func TestRequestParse_DecodeThenValidate(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":""}`))
	s := &signup{}
	if err := RequestParse(req, s); err == nil {
		t.Fatal("RequestParse() = nil, want an error for empty required Email")
	}
}

func TestRequestParse_DecodeErrorShortCircuits(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`not json`))
	s := &signup{}
	if err := RequestParse(req, s); err == nil {
		t.Fatal("RequestParse() = nil, want a JSON decode error")
	}
}

func TestRequiredFieldsError_NilPointerDoesNotPanic(t *testing.T) {
	var s *signup
	if err := requiredFieldsError(s); err == nil {
		t.Fatal("requiredFieldsError(nil) = nil, want an error")
	}
}

func TestRequiredFieldsError_NonPointerDoesNotPanic(t *testing.T) {
	if err := requiredFieldsError(signup{}); err == nil {
		t.Fatal("requiredFieldsError(non-pointer) = nil, want an error")
	}
}

func TestRequiredFieldsError_NonStructPointerDoesNotPanic(t *testing.T) {
	x := 5
	if err := requiredFieldsError(&x); err == nil {
		t.Fatal("requiredFieldsError(*int) = nil, want an error")
	}
}

func TestRequiredFieldsError_SliceOfStructPointers(t *testing.T) {
	items := []*signup{
		{Email: "a@b.com"},
		{Email: ""}, // missing
	}
	if err := requiredFieldsError(&items); err == nil {
		t.Fatal("requiredFieldsError(slice) = nil, want an error for the second element")
	}
}

func TestRequiredFieldsError_SliceOfStructValues(t *testing.T) {
	items := []signup{
		{Email: "a@b.com"},
		{Email: "c@d.com"},
	}
	if err := requiredFieldsError(&items); err != nil {
		t.Fatalf("requiredFieldsError(slice) = %v, want nil", err)
	}
}

func TestRequiredFieldsError_SliceWithNilPointerElementSkipped(t *testing.T) {
	items := []*signup{nil, {Email: "a@b.com"}}
	if err := requiredFieldsError(&items); err != nil {
		t.Fatalf("requiredFieldsError(slice with nil element) = %v, want nil (nil elements are skipped)", err)
	}
}

func TestRequiredFieldsError_ErrorNamesJSONField(t *testing.T) {
	err := requiredFieldsError(&signup{})
	if err == nil || !strings.Contains(err.Error(), "email") {
		t.Fatalf("error = %v, want it to mention the json field name %q", err, "email")
	}
}
