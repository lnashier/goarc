package http

import (
	"net/http"
	"testing"
)

func TestClientOpts_EncoderDecoderOverride(t *testing.T) {
	encCalled, decCalled := false, false
	c := NewClient(
		WithEncoder(func(v any) ([]byte, error) {
			encCalled = true
			return []byte("encoded"), nil
		}),
		WithDecoder(func(b []byte, v any) error {
			decCalled = true
			return nil
		}),
	)

	if _, err := c.encoder("x"); err != nil {
		t.Fatal(err)
	}
	if err := c.decoder([]byte("x"), nil); err != nil {
		t.Fatal(err)
	}
	if !encCalled || !decCalled {
		t.Fatalf("encCalled=%v decCalled=%v, want both true", encCalled, decCalled)
	}
}

func TestClientOpts_TransportOverride(t *testing.T) {
	tr := &http.Transport{}
	c := NewClient(WithTransport(tr))
	if c.hc.Transport != tr {
		t.Fatal("WithTransport did not set the client's Transport")
	}
}
