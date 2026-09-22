package goarc

import "context"

// Service is a component with an explicit Start/Stop lifecycle.
//
// Start must block until the service stops on its own, fails, or ctx is
// done. When ctx is done, Start must begin shutting down on its own and
// return — it cannot depend on an external call to Stop to unblock it.
//
// Stop must be safe to call at any time: before Start, concurrently with a
// running Start, after Start has already returned, or more than once.
// Implementations typically route both the internal ctx-cancellation path
// and the exported Stop method through the same idempotent shutdown logic,
// for example guarded by a sync.Once.
type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Func adapts plain functions into a Service.
//
// StartFunc is required and is called by Start. StopFunc is optional; a nil
// StopFunc makes Stop a no-op, which is convenient for a StartFunc that
// already returns promptly once ctx is done and has nothing else to
// release.
type Func struct {
	StartFunc func(ctx context.Context) error
	StopFunc  func(ctx context.Context) error
}

// Start calls StartFunc.
func (f Func) Start(ctx context.Context) error {
	return f.StartFunc(ctx)
}

// Stop calls StopFunc, if set.
func (f Func) Stop(ctx context.Context) error {
	if f.StopFunc == nil {
		return nil
	}
	return f.StopFunc(ctx)
}
