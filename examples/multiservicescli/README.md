# Multiple Services CLI App

A `cli.Service` with two commands, `service1` and `service2`, each of which launches its own
`http.Service` for just the duration of that command by calling `goarc.Up` again from inside the
command's runner, passing `goarc.Context(ctx)` — so the nested HTTP service's lifetime tracks
that one command's `ctx`, not the whole process. This is a different composition pattern from
`Group`: it's for "run this one transport for as long as this command runs," not "run several
transports together for the life of the process."

# Build & Run

```shell
go build -o bin/multiservicescli .
```

```shell
./bin/multiservicescli service1
```

```shell
./bin/multiservicescli service2
```

# Run

```shell
go run . service1
```

```shell
go run . service2
```
