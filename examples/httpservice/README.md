# Example HTTP Service

A more complete `http.Service` than `toyhttp`: a layered `internal/app` package with its own
`Controller` (`SaveExample`/`GetExample`, backed by an in-memory store), request validation via
`xhttp.RequestParse`, and a `health.Controller` registered manually as a `Component` (rather than
via `x/health.App`) alongside hand-rolled `/alive` and `/ready` routes — showing the pieces
`x/health.App` wires up for you, built by hand instead.

# Build & Run

```shell
go build -o ./bin/httpservice ./cmd
./bin/httpservice
```

# Run

```shell
go run ./cmd
```