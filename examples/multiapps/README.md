# Example Multiple Apps on Single Service

One `http.Service` assembled from four independent `App` funcs — `x/health`, `x/buildinfo`, and
two of this example's own packages (`apps/hello`, `apps/echo`) — showing that `App` is meant to
be composed from separately-maintainable pieces rather than one large setup function. The build
command also shows injecting `x/buildinfo.Version`/`.Hash` via `-ldflags -X`, so `/buildinfo`
reports real values instead of `"undefined"`.

# Build & Run

```shell
go build -trimpath -a -o ./bin/multiapps \
  -ldflags "-X github.com/lnashier/goarc/v2/x/buildinfo.Version=$(cat VERSION) \
            -X github.com/lnashier/goarc/v2/x/buildinfo.Hash=$(git rev-parse HEAD)" .
./bin/multiapps
```

# Run

```shell
go run .
```
