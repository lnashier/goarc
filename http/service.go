package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

// Service is an HTTP goarc.Service, backed by gorilla/mux for routing and
// negroni for middleware chains.
type Service struct {
	opts       serviceOpts
	httpServer *http.Server
	router     *mux.Router
	preempt    *negroni.Negroni
	components []Component

	stopped  chan struct{}
	stopOnce sync.Once
	stopErr  error

	addrMu sync.Mutex
	addr   string
}

func NewService(opt ...ServiceOpt) *Service {
	opts := defaultServiceOpts
	opts.apply(opt...)

	preempt := negroni.New()

	s := &Service{
		opts: opts,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", opts.port),
			Handler: preempt,
		},
		preempt: preempt,
		router:  mux.NewRouter(),
		stopped: make(chan struct{}),
	}

	// Configure app(s); if provided
	for _, app := range opts.apps {
		if err := app(s); err != nil {
			panic(fmt.Sprintf("failed to configure app: %v", err))
		}
	}

	return s
}

// Start implements goarc.Service. Routes should be registered before
// calling Start.
//
// Start blocks until the server stops — either because ctx is done (Start
// then shuts down on its own, within ServiceShutdownGracetime) or because
// Stop was called directly.
func (s *Service) Start(ctx context.Context) error {
	s.preempt.UseHandler(s.router)

	lis, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}
	s.setAddr(lis.Addr().String())

	startDone := make(chan struct{})
	defer close(startDone)

	go func() {
		select {
		case <-ctx.Done():
			stopCtx, cancel := context.WithTimeout(context.Background(), s.opts.shutdownGracetime)
			defer cancel()
			_ = s.stop(stopCtx)
		case <-s.stopped:
		case <-startDone:
		}
	}()

	err = s.httpServer.Serve(lis)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Addr returns the address the service is listening on, once Start has
// bound its listener; empty otherwise. It is primarily useful with
// ServicePort(0), which asks the OS to choose a free port — for example in
// tests, to avoid hardcoding a port that might already be in use.
func (s *Service) Addr() string {
	s.addrMu.Lock()
	defer s.addrMu.Unlock()
	return s.addr
}

func (s *Service) setAddr(addr string) {
	s.addrMu.Lock()
	s.addr = addr
	s.addrMu.Unlock()
}

// Stop implements goarc.Service: it gracefully shuts down the service
// without interrupting active connections, first stopping every registered
// Component, then the underlying http.Server, both bounded by ctx.
//
// Stop is safe to call before Start, concurrently with a running Start, or
// more than once — only the first call does any work; later calls return
// its result.
func (s *Service) Stop(ctx context.Context) error {
	return s.stop(ctx)
}

func (s *Service) stop(ctx context.Context) error {
	s.stopOnce.Do(func() {
		close(s.stopped)

		var errs []error
		for _, comp := range s.components {
			if err := comp.Stop(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		if err := s.httpServer.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		s.stopErr = errors.Join(errs...)
	})
	return s.stopErr
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
