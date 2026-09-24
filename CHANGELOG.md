# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- `http.ServiceHost` — bind the service to a specific host/IP (e.g. `127.0.0.1` to keep a
  metrics or admin listener on loopback). Default is unchanged (all interfaces).
- `http.ServiceReadHeaderTimeout`, `ServiceReadTimeout`, `ServiceWriteTimeout`,
  `ServiceIdleTimeout`, `ServiceMaxHeaderBytes` — set the corresponding `http.Server` timeouts (slow-client protection
  for public listeners). All default to zero, i.e. unchanged behavior. Driven by feedback from
  a downstream project that had to wrap its own `http.Server` to get these.
- CI workflow (`go build`, `go vet`, `golangci-lint`, `go test -race -cover`, plus a build+vet
  pass over every `examples/*` module) gating every commit on the `v2` branch.
- `.golangci.yml` lint configuration.
- `Group`, composing named `Service`s into one, with concurrent `Start`/`Stop` and fail-fast
  cascade shutdown when any member's `Start` returns.
- `Run(s Service, opt ...BootOpt) error`, a non-exiting counterpart to `Up` for tests and
  embedding.
- `ShutdownTimeout` boot option bounding how long `Stop` is given after `Start` returns.
- `x/log.Logger` now implements `slog.Handler`, so it can back a `slog.Logger`
  (`slog.New(logger)`) for standard, attribute-based structured logging while still producing
  goarc's JSON `Entry` shape. Attributes bound via `slog.Logger.With`/`WithGroup` are nested
  under `Entry.Attrs`, following `log/slog`'s own group-nesting and empty-group-elision rules —
  verified against the standard library's own `testing/slogtest.TestHandler` conformance suite.
- `x/log.Entry` gained an `Attrs` field (`omitempty`) carrying slog attributes/groups.
- `http.Service` now implements the new `goarc.Service` — see Changed below.
- `http.Service.Addr()` returns the actual bound address once `Start` has listened, so
  `ServicePort(0)` (an OS-assigned free port) is usable in tests and dynamic-port deployments.
- `internal/servicetest`, a shared conformance suite (`Stop` safe before/after/concurrent with
  `Start`, `Start` self-terminating on `ctx` cancellation) that any `goarc.Service`
  implementation in this module can run against itself; used by `http.Service` and now
  `grpc.Service`, `cli` next.
- `goarc.Component`, hoisted out of `http` into the core package so it's one shared concept any
  `Service` implementation can use, not a per-package duplicate. `http.Component` and
  `grpc.Component` are now aliases for it.
- `grpc.Service` now implements the new `goarc.Service` — see Changed below.
- `grpc.Service.Component`, matching `http.Service`'s — v1's `grpc.Service` had no way to
  register auxiliary long-running work at all.
- `grpc.Service.Addr()`, matching `http.Service.Addr()`, for the same `ServicePort(0)`
  testability.
- `cli.Service` now implements the new `goarc.Service` — see Changed below.
- `cli.ServiceArgs`, overriding the arguments the root command parses instead of the process's
  real `os.Args[1:]` — needed to make `cli.Service` deterministically testable at all, and
  useful for embedding a `cli.Service` where argument parsing shouldn't come from the process's
  real command line.
- `x/health.Controller` now implements the new `Component`.
- `doc.go` and tests for every `x/*` package that lacked them (`config`, `env`, `buildinfo`,
  `health`, `time`, and `json`, folded in for completeness though it wasn't originally scoped
  for this milestone).

### Changed

- All direct dependencies updated to their latest versions: `fsnotify` v1.8.0→v1.10.1, `cobra`
  v1.9.1→v1.10.2, `viper` v1.19.0→v1.21.0, `grpc-go` v1.71.0→v1.84.0 (pulling `protobuf`,
  `genproto`, and `golang.org/x/{net,sys,text}` updates transitively). `gorilla/mux`,
  `pkg/errors`, and `urfave/negroni` were already at latest. Applied the same way to every
  `examples/*` module. No code changes were needed anywhere — `go build`, `go vet`, and
  `go test ./... -race` pass unchanged across the whole module and every example, and every
  example was also re-verified with a live smoke test (not just a compile check).
- **Breaking:** minimum Go version raised from 1.24.1 to 1.26.0 — not a direct choice, but a
  transitive consequence of `grpc-go` v1.84.0's own minimum requirement (itself raised further to
  1.26.0 by `golang.org/x/net`/`x/sys`/`x/text`'s requirements). Go's module system propagated
  this automatically to every `examples/*` `go.mod` too, since each depends on this module via a
  `replace` directive into the same build list.
- **Breaking:** module path is now `github.com/lnashier/goarc/v2`, per Go's major-version
  module convention — required because of the breaking `Service`/`Component` interface changes
  in this release. Every import path in this repo (including every `examples/*` module) and in
  the README's code samples is updated accordingly.
- **Breaking:** `Service.Start` and `Service.Stop` now take a `context.Context`. `Start` must
  self-terminate when `ctx` is done rather than depending on a concurrent call to `Stop` to
  unblock it; `Stop` must be safe to call at any time, including before `Start`, concurrently,
  or more than once. This removes the race class that caused v1's `cli.Service` shutdown bug at
  its root rather than at that one call site.
- **Breaking:** `ServiceFunc` (a single func keyed by a `bool` Start/Stop flag) is replaced by
  `Func{StartFunc, StopFunc}`, with `StopFunc` optional.
- **Breaking:** `Up`'s `OnStart`/`OnStop` hooks are now pure observability callbacks — they no
  longer control whether the process exits. `Up` now always calls `Stop` after `Start` returns
  (even when `Start` failed for a reason unrelated to shutdown), and decides exit-vs-not once,
  from the joined result of both phases, fixing a v1 gap where an early, non-signal `Start`
  failure skipped `Stop`/cleanup and `onStop` entirely.
- `boot.go` now builds its shutdown context with `signal.NotifyContext` instead of a manual
  `signal.Notify` + channel, removing the goroutine/channel race pattern that produced the
  `cli.Service` bug.
- `x/http`'s `route.go` (the pre-assembled `JSONHandler`/`TextHandler`/`XMLHandler`
  `http.Handler` adapters) is renamed `handler.go` — no API change, just a file that was never
  about routing.
- `x/http.Client.DoDecoded` now treats any 2xx status as success (previously only exactly `200`
  decoded — a `201 Created` or `202 Accepted` response was wrongly treated as an error) and
  correctly returns an `*Error` for a non-2xx status even with an empty body (previously an
  empty body short-circuited to a nil error regardless of status, so e.g. an empty `404`
  response was silently treated as success).
- `x/http.Is4xx` now uses `errors.As` instead of a plain type assertion, so it recognizes a
  wrapped `*Error` (e.g. via `fmt.Errorf("...: %w", err)`), matching `ConvertError`'s existing
  behavior.
- **Breaking:** `http.Component.Stop` now takes a `context.Context` and returns an `error`
  (was `Stop()`, no return value). Errors from every registered `Component`'s `Stop`, plus the
  underlying `http.Server.Shutdown`, are joined via `errors.Join` and returned from
  `Service.Stop` — v1 silently discarded any component shutdown error entirely.
- `http.ServiceShutdownGracetime`'s semantics changed: v1 used it as a blind `time.Sleep` before
  calling `http.Server.Shutdown` (a workaround from when `Stop` had no `context.Context` to
  express a deadline with). It's now the deadline `context.WithTimeout` gives to component
  shutdown and `http.Server.Shutdown` when `Start` self-terminates on `ctx` cancellation — real
  in-flight-request draining bounded by an actual deadline, not an unconditional sleep. An
  explicit `Stop(ctx)` call uses `ctx`'s own deadline as-is and ignores this option entirely.
- `grpc.serviceopts.apply` is now variadic (`apply(opt ...ServiceOpt)`), matching `http` and
  `cli`, instead of taking a `[]ServiceOpt` slice — the option-parsing inconsistency flagged
  during the initial review.
- `grpc.Service`'s redundant `name`/`shutdownGracetime` fields (duplicating `opts.name` /
  `opts.shutdownGracetime` on the same struct) are gone; `opts` is the single source of truth,
  matching `http.Service`.
- **Breaking:** `cli.Service.Start`/`Stop` now take/use `context.Context`, replacing the
  `exitCh` field that was the actual root cause of the v1 shutdown race (see Fixed below). Note
  that unlike `http`/`grpc.Service`, `cli.Service`'s `Start` can only self-terminate on `ctx`
  cancellation to the extent the *registered command itself* honors its own `ctx` — `Start`
  blocks on user code it doesn't own, not on a framework-controlled listener loop, so this is a
  documented, inherent limitation of the CLI transport rather than something more code could
  close.
- **Breaking:** `x/health.Controller.Stop` now takes a `context.Context` and returns an
  `error`, matching the new `Component`.
- `x/buildinfo`'s `Key` constants (`KeyStartTime`, `KeyUptime`, `KeyVersion`, `KeyHash`) are now
  explicitly typed `Key`, matching `KeyHost` — v1 left them as untyped string constants that
  happened to work via implicit conversion, but were inconsistent with the declared `Key` type.

### Fixed

- `x/http` `Client.Do`: the wait between retries was a plain `time.Sleep` that ignored the
  request's context, so a cancelled or expired context could still block for the full backoff
  (up to `WaitMax` per retry). The wait now returns immediately with an error wrapping the
  context's error. Reported by a downstream project.

- **`x/health.Controller.Stop` is now idempotent.** V1's `Stop()` called `close(hc.done)`
  directly with no guard — calling it twice (now an explicit requirement of the `Component`
  contract) would panic on the second call. Guarded with `sync.Once`.
- **`x/buildinfo`'s reported `startTime` is no longer mislabeled.** V1 formatted
  `time.Now()` (in the process's local timezone) with the literal layout string
  `"2006-01-02T15:04:05Z"` — the trailing `Z` isn't a recognized timezone-offset token in that
  position, so it was appended as a literal character regardless of the actual offset, making a
  non-UTC `startTime` look like UTC. Now formatted as `startTime.UTC().Format(time.RFC3339)`.
- **`x/env`'s cached `Environment` was read/written without synchronization.** `Get()` checked
  and set a package-level var with no lock — a data race under `-race` if called concurrently
  before the first resolution. Now backed by `sync.OnceValue`.

- **The original bug that started this rebuild, actually fixed at its root.**
  `cli.Service`'s `exitCh` was created inside `Start()`, not `NewService()`; because `Up`'s
  signal-handling goroutine ran concurrently with `Start`, a shutdown signal arriving before
  `Start` had executed `exitCh = make(chan struct{})` made `Stop()` call `close(nil channel)` →
  panic. `cli.Service` no longer has an `exitCh` at all — shutdown is now signaled via a channel
  created once in `NewService` (same pattern as `http`/`grpc.Service`) and Stop is unconditionally
  safe to call before Start, per the `Service` contract from M1 and the conformance suite that
  now runs against all three transports.

- **`grpc.Service.Stop` now actually honors `ServiceShutdownGracetime`.** V1 stored the option
  but never read it in `Stop`, which just called `GracefulStop()` and would wait indefinitely
  for in-flight RPCs to finish — no way to bound it at all. `Stop` now races `GracefulStop`
  against `ctx`'s deadline (derived from `ServiceShutdownGracetime` when self-triggered, or the
  caller's own `ctx` when called explicitly) and forcibly closes the server via `Stop` (the
  grpc-go one) if the deadline passes first.

- `x/log.Logger`'s zero value is now genuinely usable, as its doc comment already claimed but
  its implementation did not deliver: every field (`Verifier`, `Publisher`, `Logger`,
  `ServiceName`, `Hostname`, `Hash`) now has a safe, lazily-applied default, and
  `DefaultLogger = &Logger{}` is literally the zero value. Previously, a hand-built `&Logger{}`
  (as opposed to the package's own `DefaultLogger` var) would nil-panic on first use.
- `Logger.Panic` now panics with the formatted message as the panic value, instead of an empty
  string — a `recover()` handler can now actually report why it panicked.
- `x/http`'s `JSONHandler`/`XMLHandler` no longer write a `200 OK` status before attempting to
  marshal the response body. A marshal failure is now caught and reported as an error response
  instead of silently producing a `200` with an empty body.
- `x/http.Client.NewRequest` no longer silently discards an error from measuring the request
  body's length (via `io.Copy`); the error is now returned instead of producing a request with
  a wrong `Content-Length`.
- `x/http`'s reflection-based required-field validator (`requiredFieldsError`, née
  `getRequiredValidationError`) no longer panics on a non-struct, a nil pointer, a slice of
  non-pointer structs, or an unexported field — each of those previously crashed the request
  handler; unsupported shapes now return a descriptive error instead.
- Every `examples/*` module updated to the new APIs and verified with a real, running
  smoke test (an HTTP request, an OS `SIGINT`, or both) rather than just a successful build:
  `byoservice`'s custom `Service` and `byofunction`'s `goarc.Func` for the new `Start(ctx)`/
  `Stop(ctx)` shapes, `websocketapp`'s `Echoer.Component` for the new `Stop(ctx) error`.
- Two `go vet` printf findings in `httpservice`/`multiapps`: `xhttp.BadRequestf(err,
  err.Error())` passed the error's message as the *format string*, not a value — a `%` in the
  message could have produced garbled output. Now `BadRequestf(err, "%s", err.Error())`.
- `httpservice`'s `/examples` demo endpoint, found broken while smoke-testing it live: raw MD5
  hash bytes were stuffed directly into a JSON string field, producing invalid UTF-8 (rendered
  as `�` replacement characters) instead of a usable ID; the ID was also never actually
  saved to the store, and the "found" check was inverted (`if ok` where it meant `if !ok`), so
  `GET /examples/{id}` could never succeed even by accident. Now hex-encodes the hash and
  actually persists it.
- `mockcli`'s `echo` command, also found while live-testing it: it printed its argument after
  `SleepWithContext` regardless of whether the sleep ran its full 10s or returned early because
  `ctx` was canceled — the exact distinction this whole rebuild is about. Now checks `ctx.Err()`
  first and returns cleanly (nil, matching how every other `Service` in this SDK treats a clean
  ctx-triggered stop) instead of printing.
- `x/buildinfo`'s reported `startTime` was mislabeled — see Fixed above, confirmed live via
  `multiapps`'s `/buildinfo` endpoint, which now correctly ends in `Z`.

This marks the start of the v2 rebuild: reworking goarc into a tested, CI-gated SDK while
keeping its core shape (`Service`, `NewService`, `App(...)`, `Register(...)`) recognizable.
Every package in the module — and now every `examples/*` module — builds, vets, and (where
applicable) passes `go test ./... -race` cleanly, with CI covering both.
