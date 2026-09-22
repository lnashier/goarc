package time

import (
	"context"
	"testing"
	"time"
)

func TestSleepWithContext_ReturnsAfterDuration(t *testing.T) {
	start := time.Now()
	SleepWithContext(context.Background(), 20*time.Millisecond)
	if elapsed := time.Since(start); elapsed < 20*time.Millisecond {
		t.Fatalf("returned after %v, want at least 20ms", elapsed)
	}
}

func TestSleepWithContext_ReturnsEarlyOnCtxDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	SleepWithContext(ctx, time.Second)
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Fatalf("returned after %v, want it to return promptly once ctx was already done", elapsed)
	}
}
