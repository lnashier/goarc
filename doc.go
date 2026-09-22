// Package goarc provides a small application lifecycle: a Service interface
// with an explicit Start/Stop contract, a Group for composing several
// Services into one, and Run/Up to drive a Service from a standalone
// process with OS-signal-triggered graceful shutdown.
//
// Transport-specific Service implementations — HTTP, gRPC, CLI — live in
// their own subpackages (github.com/lnashier/goarc/v2/http, .../grpc, .../cli)
// and depend on this package, not the other way around.
package goarc
