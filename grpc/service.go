package grpc

import (
	"fmt"
	"google.golang.org/grpc"
	"net"
	"time"
)

type Service struct {
	name              string
	grpcServer        *grpc.Server
	exitCh            chan struct{}
	shutdownGracetime time.Duration
	opts              serviceOpts
}

func NewService(opt ...ServiceOpt) *Service {
	opts := defaultServiceOpts
	opts.apply(opt)

	s := &Service{
		name:              opts.name,
		opts:              opts,
		grpcServer:        grpc.NewServer(),
		exitCh:            make(chan struct{}),
		shutdownGracetime: opts.shutdownGracetime,
	}

	// Configure app(s); if provided
	for _, app := range opts.apps {
		if err := app(s); err != nil {
			panic(fmt.Sprintf("failed to configure app: %v", err))
		}
	}

	return s
}

// Start starts the service.
// Services should be registered using RegisterService before calling Start.
func (s *Service) Start() error {
	lis, err := net.Listen(s.opts.network, fmt.Sprintf(":%d", s.opts.port))
	if err != nil {
		return err
	}
	return s.grpcServer.Serve(lis)
}

// Stop gracefully shuts down the service without interrupting any
// active connections. Stop works by first calling all long-running registered
// components, and then calling underlying grpc-server GracefulStop.
func (s *Service) Stop() error {
	s.grpcServer.GracefulStop()
	return nil
}

// RegisterService registers a service and its implementation to the gRPC server.
// This must be called before invoking Start.
func (s *Service) RegisterService(sd *grpc.ServiceDesc, ss any) {
	s.grpcServer.RegisterService(sd, ss)
}
