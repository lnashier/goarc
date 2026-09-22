# goarc v1 vs v2 — Complete Differences

This is the exhaustive reference: every API change, behavior change, addition, and bug fix
between v1 and v2, organized by package. For a task-oriented "what do I need to change in my
code" guide, see [`docs/MIGRATING.md`](MIGRATING.md) instead — it covers the subset of this list
that actually requires action, with before/after code.

Legend: 🔴 breaking · 🟡 behavior change (no compile error) · 🟢 addition · 🔧 internal fix (no
API surface change) · ⚪ non-functional (rename, doc, test-only)

## At a glance

| Area | Breaking | Behavior-only | Additions | Fixes |
|---|---|---|---|---|
| Module path | 🔴 | | | |
| `goarc` (core) | 🔴🔴🔴 | | 🟢🟢🟢 | |
| `http` | 🔴 | 🟡 | 🟢🟢 | |
| `grpc` | | | 🟢🟢 | 🔧 (gracetime) |
| `cli` | | | 🟢 | 🔧 (the shutdown race) |
| `x/log` | | 🟡 | 🟢🟢 | 🔧🔧 |
| `x/http` | | 🟡🟡 | | 🔧🔧🔧 |
| `x/health` | 🔴 | | | 🔧 |
| `x/buildinfo` | | | | 🔧🔧 |
| `x/env` | | | | 🔧 |
| `x/config`, `x/time`, `x/json` | | | | ⚪ (tests/docs only) |

## Module & versioning

| | v1 | v2 |
|---|---|---|
| Module path | `github.com/lnashier/goarc` | `github.com/lnashier/goarc/v2` 🔴 |
| Tests | none in the core module | every package has `_test.go`, most 90–100% coverage |
| CI | none | GitHub Actions: build/vet/lint/test (`-race`) on every push, plus a build+vet pass over every `examples/*` module |
| Lint | none | `golangci-lint` with its default linter set (`errcheck`, `govet`, `staticcheck`, `unused`, `ineffassign`) |
| Docs | `README.md` only | every package has a `doc.go` with a package-level comment and, in most packages, a runnable `Example` |

## Core (`goarc` package)

### `Service` interface 🔴

| | v1 | v2 |
|---|---|---|
| Signature | `Start() error` / `Stop() error` | `Start(ctx context.Context) error` / `Stop(ctx context.Context) error` |
| `Start` contract | blocks until done; something else calls `Stop` to unblock it | must self-terminate when `ctx` is done — cannot depend on an external `Stop` call |
| `Stop` contract | implicit — no documented concurrency/repeat-call guarantees | must be safe before `Start`, concurrently with `Start`, after `Start` returns, or called more than once |

This is the root of the `cli.Service` bug fix (see below) and the single biggest structural
change in v2: every `Service` implementation in this repo (`http`, `grpc`, `cli`) now satisfies a
shared conformance suite (`internal/servicetest`) that directly tests these guarantees under
`-race`.

### `ServiceFunc` → `Func` 🔴

| | v1 | v2 |
|---|---|---|
| Type | `type ServiceFunc func(bool) error` | `type Func struct { StartFunc, StopFunc func(ctx context.Context) error }` |
| Construction | `goarc.ServiceFunc(func(start bool) error {...})` | `goarc.Func{StartFunc: ..., StopFunc: ...}` |
| Stop optional? | no — same func handles both, `bool` distinguishes | yes — nil `StopFunc` is a no-op |

### `Group` 🟢 *(new)*

Composes named `Service`s into a single `Service`. Did not exist in v1 — v1 had no way to run,
say, an HTTP and a gRPC service in one process without hand-rolling the composition.

- `NewGroup() *Group`, `(*Group).Add(name string, svc Service) *Group` (chainable)
- `Start`: launches every member concurrently; the moment any member's `Start` returns (success
  or failure), cancels a shared derived context so the rest self-terminate too — one failure
  brings the whole group down rather than leaving siblings orphaned
- `Stop`: stops every member concurrently, same `ctx` for all, `errors.Join` of named failures
- `Group` itself implements `Service`, so it composes recursively and drops into `Up`/`Run`
  directly

### `Run` 🟢 *(new)*

```go
func Run(s Service, opt ...BootOpt) error
```

A non-exiting counterpart to `Up`: same lifecycle (signal handling, `Start`/`Stop` orchestration),
but returns the joined error instead of calling `os.Exit`. `Up` is now implemented as a thin
wrapper: call `Run`, `os.Exit(1)` if it returned an error. `Run` is what tests and embedding
scenarios should use.

### `Up` 🟡

| | v1 | v2 |
|---|---|---|
| Signal handling | manual `signal.Notify` + a channel, in a goroutine racing `Start` | `signal.NotifyContext` (stdlib, Go 1.16+) |
| `Stop` always called after `Start`? | no — if `Start` failed for a reason unrelated to a signal, `Stop`/`onStop` never ran (`ch` stayed nil) | yes, unconditionally, regardless of why `Start` returned |
| `OnStart`/`OnStop` role | controlled whether `os.Exit` was called | pure observability; `Up` decides exit-vs-not once, from `Run`'s joined result |
| Default `OnStart`/`OnStop` | `os.Exit(1)` on error, applied per-phase | no-ops; exit behavior lives directly in `Up`, applied once at the end |

### `ShutdownTimeout` 🟢 *(new `BootOpt`)*

```go
goarc.Up(svc, goarc.ShutdownTimeout(10*time.Second))
```

Bounds how long `Stop` is given after `Start` returns (default 10s) — the outer deadline for the
whole shutdown, distinct from any `Service`'s own internal gracetime option.

### `Component` 🟢 *(new, hoisted from `http`)*

```go
type Component interface {
	Stop(ctx context.Context) error
}
```

v1 had `http.Component` only (`Stop()`, no `ctx`, no error). v2 defines `Component` once in the
core package; `http.Component` and `grpc.Component` (new) are both aliases for it.

## `http` package

| | v1 | v2 |
|---|---|---|
| Implements `goarc.Service`? | no (pre-dates the `context.Context` interface) | yes |
| `Component.Stop` | `Stop()` | `Stop(ctx context.Context) error` 🔴 |
| Component stop errors | discarded entirely | joined via `errors.Join` into `Service.Stop`'s result 🟢 |
| `ServiceShutdownGracetime` | unconditional `time.Sleep` before `httpServer.Shutdown(context.Background())` | deadline for `context.WithTimeout`, applied to component-stop and `httpServer.Shutdown`, only when `Start` self-terminates on `ctx` cancellation 🟡 |
| Explicit `Stop(ctx)` deadline | n/a (no `ctx` param existed) | uses the caller's `ctx` as-is, ignoring `ServiceShutdownGracetime` |
| `Addr()` | doesn't exist | 🟢 returns the actual bound address once `Start` has listened — makes `ServicePort(0)` (OS-assigned port) usable |
| `route.go` | holds `JSONHandler`/`TextHandler`/`XMLHandler` | renamed `handler.go` ⚪ — same package, same exported API, just a file that was never about routing |
| Dead `exitCh` field | present, created and closed, never read | removed |
| `serviceopts.apply` | `apply(opt ...ServiceOpt)` | unchanged (already variadic) |

## `grpc` package

| | v1 | v2 |
|---|---|---|
| Implements `goarc.Service`? | no | yes |
| `Stop` honors `ServiceShutdownGracetime`? | **no** — option stored, never read; `GracefulStop()` could hang indefinitely | 🔧 **yes** — races `GracefulStop` against the deadline, force-closes via `grpc.Server.Stop` if exceeded |
| `Component` support | none — `grpc.Service` had no way to register auxiliary long-running work | 🟢 `Service.Component(comp Component)`, matching `http.Service` |
| `Addr()` | doesn't exist | 🟢 same as `http.Service.Addr()` |
| `serviceopts.apply` | `apply(opts []ServiceOpt)` (slice, not variadic) | `apply(opt ...ServiceOpt)` — now consistent with `http`/`cli` |
| Redundant struct fields | `Service.name`/`.shutdownGracetime` duplicated `opts.name`/`.shutdownGracetime` | removed; `opts` is the single source of truth |
| `grpcServer.Serve` return handling | returns whatever `Serve` returns | maps `grpc.ErrServerStopped` (returned when `Stop`/`GracefulStop` ran before `Serve` started) to `nil`, matching `http.ErrServerClosed` handling |

## `cli` package

| | v1 | v2 |
|---|---|---|
| Implements `goarc.Service`? | no | yes |
| **The shutdown race** | `exitCh` created inside `Start()`, not `NewService()` — a shutdown signal arriving before `Start` ran that line made `Stop()` call `close(nil channel)` → **panic** | 🔧 `exitCh` removed entirely; shutdown signaled via a channel created once in `NewService`, same pattern as `http`/`grpc.Service`; `Stop` is unconditionally safe before `Start` |
| `Start`'s self-termination guarantee | n/a | can only self-terminate on `ctx` cancellation to the extent the *running command* honors its own `ctx` — `Start` blocks on user code, not a framework-owned loop; this is documented as an inherent, permanent property of the CLI transport, not a gap |
| `ServiceArgs` | doesn't exist — `Execute()` always parsed real `os.Args[1:]`, making the service effectively untestable | 🟢 overrides parsed args; also useful for embedding |
| `Register`'s `runner` signature | `func(ctx context.Context, args []string) error` | unchanged |

## `x/log`

| | v1 | v2 |
|---|---|---|
| `slog.Handler` | not implemented | 🟢 implemented — `Logger` can back `slog.New(logger)` for standard attribute-based logging, still producing goarc's `Entry` JSON shape; verified against stdlib `testing/slogtest.TestHandler` |
| `Entry.Attrs` | doesn't exist | 🟢 new field (`omitempty`), carries slog attributes/groups, following slog's own nesting/elision rules |
| Zero-value usability | doc comment claimed `&Logger{}` was usable; it wasn't — nil `Verifier`/`Publisher`/`Logger` panicked | 🔧 every field has a safe lazy default; `DefaultLogger = &Logger{}` is *literally* the zero value now |
| `Logger.Panic`'s panic value | `panic("")` — empty, useless to `recover()` | 🔧🟡 `panic(msg)` — the formatted message |
| `Verifier`/`Publisher` defaults | required (nil = panic) | `Verifier` defaults to "always enabled", `Publisher` defaults to no-op |

## `x/http`

| | v1 | v2 |
|---|---|---|
| `Client.DoDecoded` success check | `resp.StatusCode == http.StatusOK` only | 🔧🟡 any 2xx status |
| `Client.DoDecoded` empty-body handling | **any** empty body → `nil` error, regardless of status (an empty `404`/`500` looked like success) | 🔧🟡 status checked first; empty body only short-circuits to success for a 2xx |
| `Client.NewRequest` body-length error | `io.Copy` error silently discarded, could produce a wrong `Content-Length` | 🔧 error now returned |
| `Client.Do`'s final "giving up" error | discards the last retry error | wraps it via `%w` when non-nil |
| `Is4xx` | plain type assertion (`err.(*Error)`) — missed a wrapped `*Error` | 🔧 `errors.As`, matching `ConvertError` |
| `JSONHandler`/`XMLHandler` | `WriteHeader(200)` called *before* marshaling; marshal error discarded → `200` with empty body on failure | 🔧 marshal first; a failure now produces a real error response |
| `requiredFieldsError` (né `getRequiredValidationError`) | panicked on a non-struct, nil pointer, slice of non-pointer structs, or certain field shapes | 🔧 returns a descriptive error instead of panicking in every case; rewritten using `reflect.Value.IsZero()` (also fixes a latent panic on unexported fields) |
| `route.go` handlers | in `route.go` | moved to `handler.go` ⚪ (package/API unchanged) |

## `x/health`

| | v1 | v2 |
|---|---|---|
| `Controller.Stop` | `Stop()`, no return value, bare `close(hc.done)` | `Stop(ctx context.Context) error` 🔴, `sync.Once`-guarded |
| Calling `Stop` twice | **panics** (double-close) | 🔧 safe no-op |
| Implements `goarc.Component`? | no (predates the interface) | yes |

## `x/buildinfo`

| | v1 | v2 |
|---|---|---|
| `startTime` formatting | `startTime.Format("2006-01-02T15:04:05Z")` on a local-zone `time.Time` — the trailing `Z` is a literal character in that layout position, not a recognized zone token, so a non-UTC `startTime` was mislabeled as UTC | 🔧 `startTime.UTC().Format(time.RFC3339)` |
| `Key` constants | only `KeyHost` explicitly typed `Key`; `KeyStartTime`/`KeyUptime`/`KeyVersion`/`KeyHash` were untyped string constants (worked via implicit conversion) | 🔧 all five explicitly typed `Key` |

## `x/env`

| | v1 | v2 |
|---|---|---|
| `Get()` caching | checked-then-set a package-level var with no synchronization — a real data race under concurrent first calls | 🔧 `sync.OnceValue` |
| Internal structure | inline in `Get()` | pure `parseEnv()` split out (same behavior, more testable) |

## `x/config`, `x/time`, `x/json`

No API or behavior changes. v2 adds `doc.go` and full test coverage to each; `x/json` in
particular had zero tests in v1 despite being used by `x/log` and `x/http`.

## Examples (`examples/*`)

All 11 example modules updated to build against v2. Only three needed real code changes beyond
the import-path rename (the ones using the changed interfaces directly):

- `byoservice` — custom `Service` → `Start(ctx)`/`Stop(ctx)`
- `byofunction` — `ServiceFunc` → `Func`
- `websocketapp` — `Echoer.Stop()` → `Stop(ctx) error`, now `sync.Once`-guarded

Also fixed, found via actually running the examples rather than just compiling them: two `go vet`
printf findings (an error's `.Error()` passed as a format string), `httpservice`'s `/examples`
demo endpoint being completely non-functional (MD5 hash bytes stuffed into a JSON string
producing invalid UTF-8; the value never actually saved; an inverted found-check), and
`mockcli`'s `echo` command not distinguishing a completed sleep from a `ctx`-canceled one.
