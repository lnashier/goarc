package goarc

import (
	"context"
	"errors"
	"testing"
)

func TestFunc_Start(t *testing.T) {
	want := errors.New("boom")
	f := Func{StartFunc: func(ctx context.Context) error { return want }}

	if got := f.Start(context.Background()); !errors.Is(got, want) {
		t.Fatalf("Start() = %v, want %v", got, want)
	}
}

func TestFunc_Stop_NilIsNoOp(t *testing.T) {
	f := Func{StartFunc: func(context.Context) error { return nil }}

	if err := f.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() with nil StopFunc = %v, want nil", err)
	}
}

func TestFunc_Stop_CallsStopFunc(t *testing.T) {
	want := errors.New("stop failed")
	called := false
	f := Func{
		StartFunc: func(context.Context) error { return nil },
		StopFunc: func(ctx context.Context) error {
			called = true
			return want
		},
	}

	if got := f.Stop(context.Background()); !errors.Is(got, want) {
		t.Fatalf("Stop() = %v, want %v", got, want)
	}
	if !called {
		t.Fatal("StopFunc was not called")
	}
}

func TestFunc_ImplementsService(t *testing.T) {
	var _ Service = Func{}
}
