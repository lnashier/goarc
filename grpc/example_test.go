package grpc_test

import (
	"context"
	"fmt"
	"time"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/lnashier/goarc/v2"
	goarcgrpc "github.com/lnashier/goarc/v2/grpc"
)

// Example shows a grpc.Service composed with goarc.Run: NewService binds
// no port until Start is called, RegisterService wires up services
// beforehand (here the standard gRPC health-check service, which needs no
// protoc-generated code of our own), and Run drives the full Start/Stop
// lifecycle — here stopping itself after a short timeout instead of
// waiting for an OS signal.
func Example() {
	svc := goarcgrpc.NewService(
		goarcgrpc.ServicePort(0), // 0: let the OS choose a free port
		goarcgrpc.App(func(s *goarcgrpc.Service) error {
			healthpb.RegisterHealthServer(s, healthOK{})
			return nil
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if err := goarc.Run(svc, goarc.Context(ctx)); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("stopped cleanly")
	// Output:
	// stopped cleanly
}

type healthOK struct {
	healthpb.UnimplementedHealthServer
}

func (healthOK) Check(context.Context, *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}
