# Mock HTTP App

An `http.Service` driven by `x/config`: loads its config from a file in the current directory
(named after `x/env`'s detected environment, e.g. `local.yaml`) and watches it for changes via
`fsnotify`, printing a message on every reload. Registers `x/health` and `x/buildinfo`, plus a
hand-rolled `http.Handler` (`CustomHandler`) for `/examples` — not one of `x/http`'s pre-assembled
adapters — to show wiring up a fully custom handler by hand.

# Build & Run

```shell
go build -o bin/ .
./bin/mockserver
```

# Run

```shell
go run .
```
