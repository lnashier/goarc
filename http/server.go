package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	name              string
	httpServer        *http.Server
	router            *mux.Router
	preempt           *negroni.Negroni
	healthController  *healthController
	components        []Component
	exitCh            chan struct{}
	shutdownGracetime time.Duration
}

func NewServer(opt ...ServerOpt) *Server {
	opts := defaultServerOpts
	opts.apply(opt)

	preempt := negroni.New()

	s := &Server{
		name: opts.name,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", opts.port),
			Handler: preempt,
		},
		preempt:           preempt,
		router:            mux.NewRouter(),
		healthController:  newHealthController(),
		exitCh:            make(chan struct{}),
		components:        make([]Component, 0),
		shutdownGracetime: opts.shutdownGracetime,
	}

	setupHealthEndpoints(s)

	// Configure app(s); if provided
	for _, app := range opts.apps {
		if err := app(s); err != nil {
			panic(fmt.Sprintf("failed to configure app: %v", err))
		}
	}

	return s
}

// Start starts the server instance
// Routes should be registered before calling start
func (s *Server) Start() error {
	s.preempt.UseHandler(s.router)
	err := s.httpServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop gracefully shuts down the server without interrupting any
// active connections. Stop works by first calling all long-running registered
// components, and then calling underlying http-server's shutdown.
func (s *Server) Stop() error {
	close(s.exitCh)
	s.healthController.setStatusDown()
	// Stop any long-running services within the app
	for _, comp := range s.components {
		comp.Stop()
	}

	time.Sleep(s.shutdownGracetime)
	return s.httpServer.Shutdown(context.Background())
}

// Register registers a RouteHandler for a given path and http.Method
// preHandlers can be optionally supplied, they will run before routeHandler in the order supplied
func (s *Server) Register(path, method string, routeHandler http.Handler, preHandlers ...negroni.Handler) {
	chain := negroni.New(preHandlers...)
	chain.UseHandler(&reject{
		h:    routeHandler,
		done: s.exitCh,
	})
	path = "/" + strings.TrimLeft(path, "/")
	if err := s.router.Handle(path, chain).Methods(method).GetError(); err != nil {
		panic(fmt.Sprintf("couldn't register %s error %v", path, err.Error()))
	}
}

// Component registers a Component that will run within the app that requires stopping when server shuts down
func (s *Server) Component(comp Component) {
	s.components = append(s.components, comp)
}

func setupHealthEndpoints(s *Server) {
	s.router.Methods(http.MethodGet).Path("/alive").HandlerFunc(s.healthController.LiveHandler)
	s.router.Methods(http.MethodGet).Path("/ready").HandlerFunc(s.healthController.ReadyHandler)
}
