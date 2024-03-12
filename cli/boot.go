package cli

import (
	"os"
	"os/signal"
	"syscall"
)

// Up is to boot up the cli Service
// Program listens for the following SIG:
//
//	syscall.SIGINT
//	syscall.SIGTERM
//	syscall.SIGQUIT
//	syscall.SIGABRT
//
// On success
func Up(s *Service) {
	go func(s *Service) {
		sic := make(chan os.Signal, 1)
		// e.g. kill -SIGQUIT <pid>
		signal.Notify(sic, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
		<-sic
		if err := s.Stop(); err != nil {
			os.Exit(1)
		}
	}(s)

	if err := s.Start(); err != nil {
		os.Exit(1)
	}
}
