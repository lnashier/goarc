package goarc

import (
	"os"
	"os/signal"
	"syscall"
)

// Up manages the lifecycle of a service. It blocks until the service shuts down or the service.Start() method returns.
// It listens to specific signals and gracefully shut down the service when any of these signals are received:
//
//	syscall.SIGINT
//	syscall.SIGTERM
//	syscall.SIGQUIT
//	syscall.SIGABRT
//
// It exits with a non-zero status code if an error occurs during either the startup or shutdown process.
func Up(s Service, opt ...BootOpt) {
	opts := defaultBootOpts
	opts.apply(opt...)

	var ch chan error

	go func(s Service) {
		sig := make(chan os.Signal, 1)
		// e.g. kill -SIGQUIT <pid>
		// Notify the channel for the specified signals
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
		// Block until a signal is received
		<-sig
		ch = make(chan error)
		// Attempt to stop the service
		ch <- s.Stop()
	}(s)

	// Start the service
	err := s.Start()
	opts.onStart(err)
	if ch != nil {
		opts.onStop(<-ch)
	}
}

type BootOpt func(*bootOpts)

type bootOpts struct {
	onStart func(error)
	onStop  func(error)
}

func (b *bootOpts) apply(opt ...BootOpt) {
	for _, o := range opt {
		o(b)
	}
}

var defaultBootOpts = bootOpts{
	// If an error occurs during startup, exit with a non-zero status code
	onStart: func(err error) {
		if err != nil {
			os.Exit(1)
		}
	},
	// If an error occurs during service stop, exit with a non-zero status code
	onStop: func(err error) {
		if err != nil {
			os.Exit(1)
		}
	},
}

func OnStart(f func(error)) BootOpt {
	return func(b *bootOpts) {
		b.onStart = f
	}
}

func OnStop(f func(error)) BootOpt {
	return func(b *bootOpts) {
		b.onStop = f
	}
}
