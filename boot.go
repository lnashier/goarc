package goarc

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run starts s, blocks until it stops — via ctx being done, an OS shutdown
// signal, or Start returning on its own — then stops it and returns any
// error from either phase, joined together with errors.Join.
//
// Run always calls Stop after Start returns, regardless of why Start
// returned, so a Service's cleanup runs even if Start failed for a reason
// unrelated to shutdown (e.g. a port already in use).
//
// Run never calls os.Exit. Use Run directly from tests, or when embedding a
// Service inside a program that wants to control its own exit behavior; use
// Up for the common case of a standalone process.
//
// Run listens for these signals and treats any of them the same as ctx
// being canceled:
//
//	syscall.SIGINT
//	syscall.SIGTERM
//	syscall.SIGQUIT
//	syscall.SIGABRT
func Run(s Service, opt ...BootOpt) error {
	opts := defaultBootOpts
	opts.apply(opt...)

	ctx, cancel := signal.NotifyContext(opts.ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
	defer cancel()

	startErr := s.Start(ctx)
	opts.onStart(startErr)

	stopCtx, stopCancel := context.WithTimeout(context.Background(), opts.shutdownTimeout)
	defer stopCancel()
	stopErr := s.Stop(stopCtx)
	opts.onStop(stopErr)

	return errors.Join(startErr, stopErr)
}

// Up manages the lifecycle of a standalone process: it calls Run and exits
// with a non-zero status code if Run returns an error from either the
// Start or the Stop phase.
//
// OnStart and OnStop are pure observability hooks — they do not affect
// whether Up exits, which is instead decided once, from Run's combined
// result, after both phases have had a chance to run.
func Up(s Service, opt ...BootOpt) {
	if err := Run(s, opt...); err != nil {
		os.Exit(1)
	}
}

type BootOpt func(*bootOpts)

type bootOpts struct {
	ctx             context.Context
	shutdownTimeout time.Duration
	onStart         func(error)
	onStop          func(error)
}

func (b *bootOpts) apply(opt ...BootOpt) {
	for _, o := range opt {
		o(b)
	}
}

var defaultBootOpts = bootOpts{
	// An empty Context. It is never canceled, has no values, and has no
	// deadline, other than whatever ShutdownTimeout applies to Stop.
	ctx:             context.Background(),
	shutdownTimeout: 10 * time.Second,
	onStart:         func(error) {},
	onStop:          func(error) {},
}

// Context sets the parent context for Run/Up. Canceling it triggers the
// same graceful-shutdown path as an OS signal — useful for tests, and for
// embedding Run/Up inside a program that already manages its own shutdown
// trigger without sending itself a real OS signal.
func Context(ctx context.Context) BootOpt {
	return func(b *bootOpts) {
		b.ctx = ctx
	}
}

// ShutdownTimeout bounds how long Stop is given to complete once Start has
// returned. It is the outer deadline for the whole shutdown; a Service's
// own internal shutdown pacing, if any, still applies within this budget.
//
// The default is 10 seconds.
func ShutdownTimeout(d time.Duration) BootOpt {
	return func(b *bootOpts) {
		b.shutdownTimeout = d
	}
}

// OnStart registers a callback invoked with the error returned by Start
// (nil on success), before Stop runs. It is an observability hook only; it
// does not affect Run's return value or whether Up exits.
func OnStart(f func(error)) BootOpt {
	return func(b *bootOpts) {
		b.onStart = f
	}
}

// OnStop registers a callback invoked with the error returned by Stop (nil
// on success). It is an observability hook only; it does not affect Run's
// return value or whether Up exits.
func OnStop(f func(error)) BootOpt {
	return func(b *bootOpts) {
		b.onStop = f
	}
}
