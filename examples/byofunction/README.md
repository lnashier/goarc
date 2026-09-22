# Example BYO Function

Shows `goarc.Func`, the lighter-weight alternative to implementing `goarc.Service` as a custom
type: a `StartFunc`/`StopFunc` pair of plain closures, with no type declaration needed at all.
Also demonstrates the `OnStart`/`OnStop` boot hooks, which are pure observability in v2 — they
never affect whether the process exits.

# Build & Run

```shell
go build -o bin/ .
./bin/byofunction
```

# Run

```shell
go run . 
```
