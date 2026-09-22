# Toy HTTP App

The smallest possible `http.Service`. Registers three routes, each using a different handler
style: `/toys/byo` with a plain `http.Handler`, `/toys/json` with the pre-assembled
`xhttp.JSONHandler`, and `/toys/text` with `xhttp.TextHandler`.

# Build & Run

```shell
go build -o bin/ .
./bin/toyhttp
```

# Run

```shell
go run .
```
