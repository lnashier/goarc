# Examples

Each example is its own Go module — `cd` into one and `go run .` (or `go run ./cmd`, where the
entry point lives under `cmd/`). See each example's own README for exact build/run commands.

| Example | What it shows |
|---|---|
| [toyhttp](toyhttp/) | Minimal `http.Service`: a BYO `http.Handler` plus the pre-assembled `JSONHandler`/`TextHandler` adapters |
| [mockhttp](mockhttp/) | `http.Service` wired to a file-based, hot-reloading `x/config.Config` |
| [httpservice](httpservice/) | A more complete, layered `http.Service` — `internal/app` structure, a `Controller`, request validation, a manually-registered `health.Controller` `Component` |
| [multiapps](multiapps/) | One `http.Service` composed from several independent `App` funcs across packages (`x/health`, `x/buildinfo`, two app packages), plus `-ldflags -X` build-info injection |
| [websocketapp](websocketapp/) | `http.Service` with a WebSocket endpoint (`gorilla/websocket`) registered as a `Component`, so open connections get a clean close signal on shutdown |
| [grpcservice](grpcservice/) | `grpc.Service` with a protoc-generated `Echo` service — unary, server-stream, client-stream, and bidi-stream RPCs |
| [mockcli](mockcli/) | Minimal `cli.Service`: one command that honors `ctx` cancellation |
| [encoder](encoder/) | `cli.Service` with two independent subcommands (`base64encode`/`base64decode`) |
| [multiservicescli](multiservicescli/) | `cli.Service` where each subcommand launches its own `http.Service` for just the duration of that command (nested `goarc.Up`, not `Group`) |
| [byoservice](byoservice/) | Bring-your-own `Service`: implements `goarc.Service` directly instead of using a transport package |
| [byofunction](byofunction/) | Adapts a pair of plain functions into a `Service` via `goarc.Func`, no custom type needed |
