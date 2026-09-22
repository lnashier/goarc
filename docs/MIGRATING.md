# Migrating from goarc v1 to v2

goarc v2 is a rebuild of goarc, not a patch release: the core `Service` lifecycle is now
`context.Context`-aware, which is what let it fix a real shutdown race in `cli.Service` and a
gracetime bug in `grpc.Service` at their root instead of papering over them. The overall shape —
`Service`, `NewService`, `App(...)`, `Register(...)` — is unchanged, so most application code
(route handlers, gRPC service implementations, CLI command bodies) needs no changes at all. What
needs updating is concentrated in a handful of places, covered below in the order you're likely
to hit them.

See [`docs/DIFFERENCES.md`](DIFFERENCES.md) for the complete, exhaustive list of what changed,
including non-breaking additions and internal bug fixes. This document covers only what you need
to *do* to move existing v1 code to v2.

## 0. Update the import path (required first step)

v2 is published under a new module path, per Go's convention for a breaking major version:

```diff
-import "github.com/lnashier/goarc"
+import "github.com/lnashier/goarc/v2"
```

Every subpackage moves the same way:

```diff
-goarchttp "github.com/lnashier/goarc/http"
-xhttp     "github.com/lnashier/goarc/x/http"
+goarchttp "github.com/lnashier/goarc/v2/http"
+xhttp     "github.com/lnashier/goarc/v2/x/http"
```

If you build with `-ldflags -X` to set `x/buildinfo.Version`/`.Hash`, update those target paths
too — an `-X` flag that no longer matches a real symbol path fails silently, not loudly:

```diff
--ldflags "-X github.com/lnashier/goarc/x/buildinfo.Version=$VERSION"
+-ldflags "-X github.com/lnashier/goarc/v2/x/buildinfo.Version=$VERSION"
```

The easiest way to catch every remaining reference: `grep -rn 'lnashier/goarc"' --include='*.go' .`
and `grep -rn 'lnashier/goarc/' --include='*.go' .` (excluding any that already say `/v2/`).

## 1. If you implemented a custom `Service`

This is the change everything else follows from. `Start`/`Stop` now take a `context.Context`:

```diff
-func (s *MyService) Start() error {
+func (s *MyService) Start(ctx context.Context) error {
 	...
 }

-func (s *MyService) Stop() error {
+func (s *MyService) Stop(ctx context.Context) error {
 	...
 }
```

Two behavioral requirements come with the new signature, and both are worth actually reading
rather than skimming — they're what the whole rebuild hinges on:

- **`Start` must self-terminate when `ctx` is done.** It can no longer assume something else
  will call `Stop` to unblock it — `Stop` might never be called at all in some flows (e.g. `Run`
  driven purely by ctx cancellation). If `Start`'s own blocking call doesn't take a `ctx`, wrap it
  so it does — `x/time.SleepWithContext` is exactly this for a plain sleep, and is a good
  reference pattern if your `Start` needs to watch a channel *or* `ctx.Done()`.
- **`Stop` must be safe to call at any time** — before `Start`, concurrently with a running
  `Start`, after `Start` has already returned, or more than once. If your v1 `Stop` closed a
  channel unconditionally (a common pattern — it's exactly what caused the `cli.Service` bug this
  rebuild started from), guard it with `sync.Once` now that calling it twice is a real,
  documented possibility rather than a caller error.

Before/after, adapted from this repo's own `byoservice` example:

```go
// v1
type Service struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *Service) Start() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	xtime.SleepWithContext(s.ctx, 10*time.Second)
	return nil
}

func (s *Service) Stop() error {
	s.cancel()
	return nil
}
```

```go
// v2 — ctx now comes in as a parameter, so there's no need to build and store your own
type Service struct{}

func (s *Service) Start(ctx context.Context) error {
	xtime.SleepWithContext(ctx, 10*time.Second) // already returns early on ctx.Done()
	return nil
}

func (s *Service) Stop(context.Context) error {
	return nil
}
```

Note what disappeared: the `ctx`/`cancel` fields and their manual wiring. In most cases v2's
`Service` is *less* code than v1's, not more — the framework now hands you the context you used
to have to build yourself.

## 2. `goarc.ServiceFunc` → `goarc.Func`

The single-function-keyed-by-a-`bool` adapter is replaced by two named, independently optional
funcs:

```diff
-goarc.Up(
-	goarc.ServiceFunc(func(start bool) error {
-		if !start {
-			cancel()
-			return nil
-		}
-		xtime.SleepWithContext(ctx, 10*time.Second)
-		return nil
-	}),
-)
+goarc.Up(
+	goarc.Func{
+		StartFunc: func(ctx context.Context) error {
+			xtime.SleepWithContext(ctx, 10*time.Second)
+			return nil
+		},
+		StopFunc: func(context.Context) error {
+			return nil
+		},
+	},
+)
```

`StopFunc` is optional — a nil `StopFunc` makes `Stop` a no-op, which covers the common case
where `StartFunc` already returns promptly on `ctx.Done()` and there's nothing else to release.
As with a custom `Service`, you usually no longer need your own `ctx`/`cancel` pair — `StartFunc`
receives one directly.

## 3. `Up`'s `OnStart`/`OnStop` hooks are now observability-only

In v1, whether `Up` called `os.Exit` was entangled with your `OnStart`/`OnStop` callbacks. In v2
they're pure observation — they never affect control flow — and `Up` decides exit-vs-not exactly
once, from the combined result of both `Start` and `Stop`, after both have actually run. The
signatures (`OnStart(func(error))`, `OnStop(func(error))`) are unchanged, so if your hooks just
log or record metrics, **no change is needed**.

The one behavior to double-check: **`Stop` is now always called after `Start` returns, no matter
why it returned** — including a `Start` failure unrelated to shutdown (e.g. "address already in
use"). In v1, that specific case meant `Stop`/cleanup/`OnStop` never ran at all. If your `OnStop`
hook assumed it would never fire after an early `Start` failure, that assumption no longer holds
— which is very likely what you actually want (registered components now reliably get cleaned
up), but worth a look if your hook has side effects tied to "we definitely stopped cleanly."

If you need a non-exiting entry point (for tests, or to embed a `Service` in a program that
controls its own exit behavior), use the new `goarc.Run`, which `Up` is now built on top of:

```go
err := goarc.Run(svc, goarc.Context(ctx)) // never calls os.Exit; returns the joined error
```

## 4. If you registered a `Component` (`http.Service.Component` / `grpc.Service.Component`)

`Component.Stop` now takes a `context.Context` and returns an `error`:

```diff
-func (c *MyComponent) Stop() {
-	close(c.done)
-}
+func (c *MyComponent) Stop(ctx context.Context) error {
+	c.stopOnce.Do(func() { close(c.done) })
+	return nil
+}
```

The `sync.Once` guard is worth adding even if your v1 code didn't need it — `Stop` being called
more than once is now a documented, real possibility (each registered `Component`'s `Stop` is
called once during shutdown, but nothing stops a caller from also calling it directly), and a bare
`close()` panics on the second call.

`http.Component` and `grpc.Component` are now both aliases for the same `goarc.Component` — if
you referenced the type by name (`http.Component` or `grpc.Component` in a variable declaration,
for instance), no change is needed; the alias keeps that working.

Any error your `Stop` returns is no longer silently discarded — it's joined (via `errors.Join`)
into the `Service.Stop` error that flows back to `Run`/`Up`/`Group`, so if you were quietly
swallowing a stop error before, it may now surface for the first time.

## 5. If you set `http.ServiceShutdownGracetime` or `grpc.ServiceShutdownGracetime`

The option's **name and signature are unchanged**, but what it *does* changed, and this is the
one item on this list that's a behavior change with no compile error to flag it — read this even
if `go build` passes clean.

- **v1**: an unconditional `time.Sleep(gracetime)` before shutting down, regardless of whether
  anything was actually still in flight. This was a workaround from when `Stop` had no
  `context.Context` to express a deadline with.
- **v2**: the deadline given to `context.WithTimeout` for stopping components and for the actual
  server shutdown (`http.Server.Shutdown` / racing `grpc.Server.GracefulStop`), when `Start`
  self-terminates on `ctx` cancellation. If everything drains in 50ms, shutdown now takes 50ms,
  not the full gracetime — and if a slow gRPC RPC is still running when the deadline passes, it's
  now forcibly closed instead of the deadline being read and ignored (v1's actual gRPC bug: the
  option was stored and never consulted in `Stop` at all).

If you were relying on the gracetime as "always wait N seconds before shutting down" (e.g. to
give a load balancer time to notice a flipped readiness probe), that specific behavior is gone —
model it explicitly instead, e.g. a small delay inside your readiness `Component`'s own `Stop`
before flipping status, if you still need it.

An explicit `Stop(ctx)` call — as opposed to `Start` self-terminating on `ctx` — uses your own
`ctx`'s deadline as-is and doesn't apply this option at all, in both v1 and v2.

## 6. If you called `cli.Service` commands programmatically or in tests

`cobra`'s `Execute()` parses the process's real `os.Args[1:]` by default, which made `cli.Service`
essentially untestable in v1. New in v2: `cli.ServiceArgs([]string)` overrides what gets parsed:

```go
svc := goarccli.NewService(goarccli.ServiceArgs([]string{"mycommand", "arg1"}))
```

This is additive — nothing to migrate — but worth knowing about if you're writing tests against
a `cli.Service` for the first time in v2.

Your registered commands' `runner` signature (`func(ctx context.Context, args []string) error`)
is **unchanged**. The one thing worth double-checking: `Start` can only self-terminate on `ctx`
cancellation to the extent your command actually honors its own `ctx` — `cli.Service.Start`
blocks on your code, not a framework-owned listener loop, so this is where requirement #1 above
(self-terminate on `ctx.Done()`) is entirely on you to implement, not something the framework
does for you the way it does for `http`/`grpc.Service`.

## 7. If you use `x/http.Client.DoDecoded` / `GetDecoded` / `PostEncodedDecoded` / `PatchEncodedDecoded`

Two correctness fixes changed response handling — check whether your calling code compensated
for either bug, since compensating code would now be redundant or actively wrong:

- **Any 2xx status is now treated as success**, not just exactly `200`. If your code special-cased
  `201`/`202`/`204` as errors (working around the v1 bug), remove that workaround.
- **A non-2xx response is now always an error, even with an empty body.** v1 returned `nil` error
  for *any* empty body regardless of status — an empty `404` or `500` looked like success. If you
  added your own status-code check after calling `DoDecoded` because you couldn't trust its error
  return, you can likely remove it now.

## 8. If you implemented `RequestValidation` for `x/http.RequestParse`/`RequestValidate`

No signature change, but the reflection-based `required`-tag walker was hardened to never panic
(previously it could panic on a non-struct, a nil pointer, a slice of non-pointer structs, or
certain field shapes) — it now returns a descriptive error instead. If you had defensive
`recover()` around a call to `RequestValidate`/`RequestParse` specifically to catch that, it's no
longer necessary, though it's harmless to leave in place.

## 9. If you use `x/log.Logger` directly (not just the package-level `Debug`/`Info`/`Error` funcs)

Two things, both narrow:

- `Logger.Panic`'s panic value changed from an empty string (`""`) to the formatted message. If
  you have a `recover()` handler that pattern-matches on the panicked value, update it.
- A hand-built `&log.Logger{}` is now genuinely a usable zero value (as v1's doc comment already
  incorrectly claimed) — if you were working around v1's nil-panic by always going through
  `log.DefaultLogger` or manually populating every field, that workaround is no longer necessary.

`Logger` also now implements `slog.Handler`, so it can back a `slog.Logger` via `slog.New(logger)`
for standard attribute-based structured logging — this is additive, nothing to migrate.

## 10. If you called `x/health.Controller.Stop()` directly (outside of `Component` registration)

Same shape change as item 4: `Stop()` → `Stop(ctx context.Context) error`.

## Quick-reference checklist

- [ ] Update every import path: `github.com/lnashier/goarc` → `github.com/lnashier/goarc/v2`
      (including `-ldflags -X` targets, if you set `x/buildinfo.Version`/`.Hash`)
- [ ] Custom `Service` implementations: `Start()`/`Stop()` → `Start(ctx)`/`Stop(ctx)`; make `Start`
      self-terminate on `ctx.Done()`; make `Stop` safe to call more than once
- [ ] `goarc.ServiceFunc(...)` → `goarc.Func{StartFunc: ..., StopFunc: ...}`
- [ ] Custom `Component` implementations: `Stop()` → `Stop(ctx) error`; guard with `sync.Once` if
      it does anything non-idempotent
- [ ] Re-check any `OnStop` hook logic that assumed it wouldn't fire after an early `Start`
      failure
- [ ] Re-check any code relying on `ServiceShutdownGracetime` as an unconditional sleep
- [ ] Re-check any workaround for `x/http.Client`'s old 2xx/empty-body handling bugs
- [ ] Re-check any `recover()` matching on `Logger.Panic`'s old `""` panic value
