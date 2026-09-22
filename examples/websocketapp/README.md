# Example Websocket App

An `http.Service` with a `/echo` WebSocket endpoint (`gorilla/websocket`). Each connection gets
its own `Echoer`, registered as a `Component` for the duration of that connection — so when the
service shuts down, every open connection gets a clean close signal instead of being dropped.
Config-driven via `x/config.Get()`, with `x/health` and `x/buildinfo` also registered.

# Build & Run

```shell
go build -o ./bin/websocketapp ./cmd
./bin/websocketapp
```

# Run

```shell
go run ./cmd
```