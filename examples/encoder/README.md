# Example CLI App

A `cli.Service` with two independent, unrelated subcommands — `base64encode` and
`base64decode` — showing that a single `cli.Service` isn't limited to one command; `Register`
can be called as many times as needed to build out a real command-line tool.

# Build & Run

```shell
go build -o bin/ .
./bin/encoder base64encode str1 | xargs -L1 ./bin/encoder base64decode
```

# Run

```shell
go run . base64encode str1 
```
