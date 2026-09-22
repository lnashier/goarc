package time

import (
	"context"
	"time"
)

// SleepWithContext sleeps for d, or until ctx is done, whichever comes first.
func SleepWithContext(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
