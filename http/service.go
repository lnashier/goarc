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

type Service struct {
	opts       serviceOpts
	httpServer *http.Server
	router     *mux.Router
	preempt    *negroni.Negroni
	exitCh     chan struct{}
	components []Component
}

func NewService(opt ...ServiceOpt) *Service {
	opts := defaultServiceOpts
	opts.apply(opt)

	preempt := negroni.New()

	s := &Service{
		opts: opts,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", opts.port),
			Handler: preempt,
		},
		preempt:    preempt,
		router:     mux.NewRouter(),
		exitCh:     make(chan struct{}),
		components: make([]Component, 0),
	}

	// Configure app(s); if provided
	for _, app := range opts.apps {
		if err := app(s); err != nil {
			panic(fmt.Sprintf("failed to configure app: %v", err))
		}
	}

	return s
}

// Start starts the service
// Routes should be registered before calling start
func (s *Service) Start() error {
	s.preempt.UseHandler(s.router)
	err := s.httpServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop gracefully shuts down the service without interrupting any
// active connections. Stop works by first calling all long-running registered
// components, and then calling underlying http-server Shutdown.
func (s *Service) Stop() error {
	close(s.exitCh)
	// Stop any long-running components within the service
	for _, comp := range s.components {
		comp.Stop()
	}
	time.Sleep(s.opts.shutdownGracetime)
	return s.httpServer.Shutdown(context.Background())
}

// Register registers a route-handler (http.Handler) for a given path and http.Method.
// Pre-handler(s) can optionally be supplied, they will run before route-handler in the order supplied.
func (s *Service) Register(path, method string, routeHandler http.Handler, preHandler ...negroni.Handler) {
	chain := negroni.New(preHandler...)
	chain.UseHandler(routeHandler)
	path = "/" + strings.TrimLeft(path, "/")
	if err := s.router.Handle(path, chain).Methods(method).GetError(); err != nil {
		panic(fmt.Sprintf("couldn't register %s error %v", path, err.Error()))
	}
}

// Component registers a Component that will run within the service that requires stopping when service shuts down.
func (s *Service) Component(comp Component) {
	s.components = append(s.components, comp)
}
