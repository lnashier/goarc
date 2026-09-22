# Idiomatic Go Applications Framework

[![GoDoc](https://pkg.go.dev/badge/github.com/lnashier/goarc/v2)](https://pkg.go.dev/github.com/lnashier/goarc/v2)
[![CI](https://github.com/lnashier/goarc/actions/workflows/ci.yml/badge.svg)](https://github.com/lnashier/goarc/actions/workflows/ci.yml)

The goarc is an idiomatic Go applications framework that focuses on application Life-Cycle.

Its core is small: a `Service` interface (`Start(ctx) error` / `Stop(ctx) error`), a `Group` that
composes several Services into one, and `Run`/`Up` to drive a Service from a process with
OS-signal-triggered graceful shutdown. `http`, `grpc`, and `cli` are `Service` implementations
for those transports; everything under `x/` is optional support (logging, an HTTP client, config,
health checks, ...).

## Use Cases

goarc is aimed at long-running Go processes that need to start up, serve, and shut down cleanly
— the shape almost every backend service, worker, and CLI tool needs, but that's easy to get
subtly wrong (a shutdown signal arriving mid-startup, a slow RPC blocking termination forever, a
readiness probe that doesn't flip when it should). Concretely:

- **HTTP APIs and microservices.** `http.Service` gives you routing (`gorilla/mux`), middleware
  chains (`negroni`), and graceful shutdown that actually drains in-flight requests within a
  deadline instead of either cutting them off or blocking forever. Pair it with `x/health` for
  Kubernetes-style `/alive`/`/ready` endpoints (readiness flips automatically when the service
  starts stopping) and `x/buildinfo` for a `/buildinfo` endpoint reporting version, commit hash,
  host, and uptime — the first thing you want when debugging "which build is actually deployed."
- **gRPC services.** `grpc.Service` gives the same lifecycle guarantees for gRPC: shutdown tries
  `GracefulStop` first and falls back to a forced `Stop` if a slow or stuck RPC would otherwise
  hang termination past its deadline.
- **CLI tools and operational scripts.** `cli.Service` (built on `cobra`) gives every subcommand
  a `context.Context` that's canceled on `Ctrl-C`/`SIGTERM`, so a long-running command (a
  migration, a batch job, a dev-loop watcher) can stop cleanly mid-operation instead of being
  killed mid-write. `ServiceArgs` makes commands testable without touching the real process
  argv.
- **Processes that run more than one transport.** `Group` composes any number of `Service`s —
  HTTP and gRPC side by side, or an API server alongside a background queue consumer or poller —
  under one call to `Run`/`Up`. One member failing brings the whole group down cleanly instead of
  leaving orphaned siblings running. A `cli.Service` command can also launch its own `Service` for
  just the duration of that command, by calling `goarc.Up`/`Run` again from inside the command's
  runner with the command's own `ctx` (see `examples/multiservicescli`) — a different composition
  from `Group`, useful when the nested service's lifetime should track one specific command
  rather than the whole process.
- **Resilient outbound HTTP calls.** `x/http.Client` adds configurable retry with exponential
  backoff, JSON encode/decode helpers, and a typed `*Error` carrying the HTTP status — for
  talking to upstreams that are occasionally flaky without hand-rolling retry logic per call
  site.
- **Structured, aggregation-friendly logging.** `x/log.Logger` emits JSON lines with a stable
  shape (hostname, service, commit hash, level, timestamp, message) that's easy to ship to
  Splunk/ELK/Datadog-style backends, and also implements `slog.Handler` — so it can back a
  standard `log/slog.Logger` and interoperate with any library or code already using the
  stdlib's structured logging, while still producing goarc's JSON shape.
- **Config that can change without a restart.** `x/config` wraps Viper with file- or path-based
  loading, environment-variable binding, and optional `fsnotify`-based hot-reload via a callback
  — useful for feature flags or tunables you don't want to redeploy for.
- **Testing services, not just unit-testing handlers.** `Run` (unlike `Up`) never calls
  `os.Exit`, so a `Service`'s full lifecycle is directly testable; `ServicePort(0)` plus `Addr()`
  get you a real, OS-assigned port with no test-to-test collisions.

## Installation

Simply add the following import to your code, and then `go [build|run|test]` will automatically fetch the necessary
dependencies:

```
import "github.com/lnashier/goarc/v2"
```

## Examples

Each example is its own Go module (see [examples/](examples/)); `cd` into one and `go run .` (or
`go run ./cmd`, where the entry point lives under `cmd/`).

| Example | What it shows |
|---|---|
| [toyhttp](examples/toyhttp/) | Minimal `http.Service`: a BYO `http.Handler` plus the pre-assembled `JSONHandler`/`TextHandler` adapters |
| [mockhttp](examples/mockhttp/) | `http.Service` wired to a file-based, hot-reloading `x/config.Config` |
| [httpservice](examples/httpservice/) | A more complete, layered `http.Service` — `internal/app` structure, a `Controller`, request validation, a manually-registered `health.Controller` `Component` |
| [multiapps](examples/multiapps/) | One `http.Service` composed from several independent `App` funcs across packages (`x/health`, `x/buildinfo`, two app packages), plus `-ldflags -X` build-info injection |
| [websocketapp](examples/websocketapp/) | `http.Service` with a WebSocket endpoint (`gorilla/websocket`) registered as a `Component`, so open connections get a clean close signal on shutdown |
| [grpcservice](examples/grpcservice/) | `grpc.Service` with a protoc-generated `Echo` service — unary, server-stream, client-stream, and bidi-stream RPCs |
| [mockcli](examples/mockcli/) | Minimal `cli.Service`: one command that honors `ctx` cancellation |
| [encoder](examples/encoder/) | `cli.Service` with two independent subcommands (`base64encode`/`base64decode`) |
| [multiservicescli](examples/multiservicescli/) | `cli.Service` where each subcommand launches its own `http.Service` for just the duration of that command (nested `goarc.Up`, not `Group` — see [Composing Multiple Services](#composing-multiple-services)) |
| [byoservice](examples/byoservice/) | Bring-your-own `Service`: implements `goarc.Service` directly instead of using a transport package |
| [byofunction](examples/byofunction/) | Adapts a pair of plain functions into a `Service` via `goarc.Func`, no custom type needed |

## Upgrading from v1

See [`docs/MIGRATING.md`](docs/MIGRATING.md) for a guided walkthrough of what needs to change in
your code, or [`docs/DIFFERENCES.md`](docs/DIFFERENCES.md) for the complete, exhaustive list of
every difference between v1 and v2.

## Toy HTTP Example

```go
package main

import (
	"github.com/lnashier/goarc/v2"
	goarchttp "github.com/lnashier/goarc/v2/http"
	xhttp "github.com/lnashier/goarc/v2/x/http"
	"net/http"
	"time"
)

func main() {
	goarc.Up(goarchttp.NewService(
		goarchttp.ServiceName("toy"),
		goarchttp.ServicePort(8080),
		goarchttp.ServiceShutdownGracetime(2*time.Second),
		goarchttp.App(func(srv *goarchttp.Service) error {

			// BYO http.Handler
			srv.Register("/toys/byo", http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello World!"))
			}))

			// Use pre-assembled http.Handler to work with JSON response type
			srv.Register("/toys/json", http.MethodGet, xhttp.JSONHandler(func(r *http.Request) (any, error) {
				return []string{"Hello World!"}, nil
			}))

			// Use pre-assembled http.Handler to work with TEXT response type
			srv.Register("/toys/text", http.MethodGet, xhttp.TextHandler(func(r *http.Request) (string, error) {
				return "Hello World!", nil
			}))

			return nil
		}),
	))
}
```

## Toy gRPC Example

```go
package main

import (
	"context"
	"github.com/lnashier/goarc/v2"
	goarcgrpc "github.com/lnashier/goarc/v2/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	goarc.Up(goarcgrpc.NewService(
		goarcgrpc.ServiceName("toy"),
		goarcgrpc.ServicePort(5001),
		goarcgrpc.App(func(srv *goarcgrpc.Service) error {
			// RegisterService works with any grpc.ServiceRegistrar-shaped
			// call, including protoc-generated Register*Server functions —
			// here the standard gRPC health-check service, which needs no
			// codegen of our own.
			healthpb.RegisterHealthServer(srv, toyHealth{})
			return nil
		}),
	))
}

type toyHealth struct {
	healthpb.UnimplementedHealthServer
}

func (toyHealth) Check(context.Context, *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}
```

## Toy CLI Example

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/lnashier/goarc/v2"
	goarccli "github.com/lnashier/goarc/v2/cli"
	xtime "github.com/lnashier/goarc/v2/x/time"
	"time"
)

func main() {
	goarc.Up(goarccli.NewService(
		goarccli.ServiceName("mockcli"),
		goarccli.App(
			func(svc *goarccli.Service) error {
				svc.Register("echo", func(ctx context.Context, args []string) error {
					xtime.SleepWithContext(ctx, time.Duration(10)*time.Second)

					// A registered command must honor ctx being done as a
					// request to return promptly — Run/Up (or Stop) signal
					// shutdown by canceling it.
					if ctx.Err() != nil {
						return nil
					}

					if len(args) > 0 {
						fmt.Println(args[0])
						return nil
					}

					return errors.New("nothing to echo")
				})

				return nil
			},
		),
	))
}
```

## Composing Multiple Services

A process that needs to run more than one transport — say HTTP and gRPC together — composes them
with `Group` instead of picking one:

```go
goarc.Up(
	goarc.NewGroup().
		Add("http", httpService).
		Add("grpc", grpcService),
)
```

`Group` starts every member concurrently and, the moment any one of them stops (cleanly or not),
brings the rest down too — so a crashed member never leaves its siblings running orphaned. See
[Group](https://pkg.go.dev/github.com/lnashier/goarc/v2#Group) for details.
