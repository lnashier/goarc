package http

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// ServerUp is to start the server and wait for shutdown.
// Function blocks until server shuts down.
func ServerUp(s *Server) {
	go func(s *Server) {
		sig := make(chan os.Signal, 1)
		// e.g. kill -SIGQUIT <pid>
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
		<-sig
		err := s.Stop()
		if err != nil {
			panic(fmt.Sprintf("couldn't stop server: %v", err))
		}
	}(s)

	err := s.Start()
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Sprintf("couldn't start server: %v", err))
		}
	}
}
