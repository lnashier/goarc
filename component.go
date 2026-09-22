package goarc

import "context"

// Component is long-running work that a Service (http.Service,
// grpc.Service, ...) runs alongside itself and must stop when the Service
// stops.
//
// Stop must be safe to call at any time, including before the Service has
// started, and should honor ctx's deadline.
type Component interface {
	Stop(ctx context.Context) error
}
