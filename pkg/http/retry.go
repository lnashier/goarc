package http

import (
	"math"
	"net/http"
	"time"
)

type Retry struct {
	WaitMin time.Duration
	WaitMax time.Duration
	Max     int
	Policy  RetryPolicy
	Backoff BackoffPolicy
	OnTry   func(time.Duration, int, error)
}

func DefaultRetry() *Retry {
	return &Retry{
		WaitMin: 2 * time.Second,
		WaitMax: 32 * time.Second,
		Max:     3,
		Policy:  DefaultRetryPolicy,
		Backoff: DefaultBackoffPolicy,
		OnTry:   func(time.Duration, int, error) {},
	}
}

func noRetry() *Retry {
	return &Retry{
		Policy: func(_ *http.Response, err error) (bool, error) {
			return false, err
		},
	}
}

// RetryPolicy specifies a policy for handling retries. It is called following each request with the response and error
// values returned by the http.Client. If returns false; the client stops retrying and returns the response to the
// caller. If returns an error; that error value is returned in lieu of the error from the request.
type RetryPolicy func(resp *http.Response, err error) (bool, error)

// DefaultRetryPolicy provides a default callback for RetryPolicy, which will retry on connection errors and server errors.
func DefaultRetryPolicy(resp *http.Response, err error) (bool, error) {
	if err != nil {
		return true, err
	}
	// Check the response code. Retry on 500-range responses to allow the server to recover
	if resp.StatusCode == 0 || resp.StatusCode >= 500 {
		return true, nil
	}
	return false, nil
}

// BackoffPolicy specifies a policy for how long to wait between retries. It is called following after a failing request to
// determine the amount of time that should pass before trying again.
type BackoffPolicy func(min, max time.Duration, attempt int, resp *http.Response) time.Duration

// DefaultBackoffPolicy provides a default callback for BackoffPolicy, which will perform exponential backoff based on the attempt
// number and limited by the provided minimum and maximum durations.
func DefaultBackoffPolicy(min, max time.Duration, attempt int, _ *http.Response) time.Duration {
	m := math.Pow(2, float64(attempt)) * float64(min)
	d := time.Duration(m)
	if float64(d) != m || d > max {
		d = max
	}
	return d
}
