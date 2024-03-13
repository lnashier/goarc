package time

import (
	"context"
	"time"
)

func SleepWithContext(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
		break
	case <-time.After(d):
		break
	}
}
