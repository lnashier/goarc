package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
)

// Service is a gRPC goarc.Service.
type Service struct {
	opts       serviceOpts
	grpcServer *grpc.Server
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

	s := &Service{
		opts:       opts,
		grpcServer: grpc.NewServer(),
		stopped:    make(chan struct{}),
	}

	// Configure app(s); if provided
	for _, app := range opts.apps {
		if err := app(s); err != nil {
			panic(fmt.Sprintf("failed to configure app: %v", err))
		}
	}

	return s
}

// Start implements goarc.Service. Services should be registered using
// RegisterService before calling Start.
//
// Start blocks until the server stops — either because ctx is done (Start
// then shuts down on its own, within ServiceShutdownGracetime) or because
// Stop was called directly.
func (s *Service) Start(ctx context.Context) error {
	lis, err := net.Listen(s.opts.network, fmt.Sprintf(":%d", s.opts.port))
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

	// Serve returns nil once Stop or GracefulStop causes it to unblock —
	// except when Stop/GracefulStop had already been called before Serve
	// even started, in which case it returns ErrServerStopped instead.
	err = s.grpcServer.Serve(lis)
	if errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}
	return err
}

// Stop implements goarc.Service: it gracefully shuts down the service —
// first stopping every registered Component, then the underlying
// grpc.Server via GracefulStop, both bounded by ctx. If ctx is done before
// GracefulStop finishes (e.g. a slow or stuck in-flight RPC), Stop forcibly
// closes the server via grpc.Server.Stop rather than waiting indefinitely,
// which GracefulStop alone would do.
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

		graceful := make(chan struct{})
		go func() {
			s.grpcServer.GracefulStop()
			close(graceful)
		}()

		select {
		case <-graceful:
		case <-ctx.Done():
			s.grpcServer.Stop()
			<-graceful // GracefulStop returns promptly once Stop forces connections closed
		}

		s.stopErr = errors.Join(errs...)
	})
	return s.stopErr
}

// RegisterService registers a service and its implementation to the gRPC server.
// This must be called before invoking Start.
func (s *Service) RegisterService(sd *grpc.ServiceDesc, ss any) {
	s.grpcServer.RegisterService(sd, ss)
}

// Component registers a Component that will run within the service that requires stopping when service shuts down.
func (s *Service) Component(comp Component) {
	s.components = append(s.components, comp)
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
