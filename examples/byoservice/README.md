# Example BYO Service

Shows implementing `goarc.Service` directly — no `http`/`grpc`/`cli` transport package — for
cases where a process's lifecycle doesn't fit any of the built-in transports. `Start(ctx)` does
10 seconds of "work" via `xtime.SleepWithContext(ctx, ...)`, which already returns early once
`ctx` is done, satisfying the `Service` contract's self-termination requirement with no extra
plumbing.

# Build & Run

```shell
go build -o bin/ .
./bin/byoservice
```

# Run

```shell
go run . 
```
