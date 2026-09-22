// Package grpc provides a gRPC goarc.Service: a grpc-go server with
// bounded graceful shutdown (falling back to a forceful stop if
// GracefulStop doesn't finish within ServiceShutdownGracetime), driven by
// NewService and RegisterService, and composable into a larger process via
// goarc.Up/Run.
package grpc
