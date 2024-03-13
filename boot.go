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
func Up(s Service) {
	go func(s Service) {
		sig := make(chan os.Signal, 1)
		// e.g. kill -SIGQUIT <pid>
		// Notify the channel for the specified signals
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
		// Block until a signal is received
		<-sig
		// Attempt to stop the service
		err := s.Stop()
		// If an error occurs during service stop, exit with a non-zero status code
		if err != nil {
			os.Exit(1)
		}
	}(s)

	// Start the service, and if an error occurs during startup, exit with a non-zero status code
	if err := s.Start(); err != nil {
		os.Exit(1)
	}
}
