# Example Multiple Apps on Single Service

# Build & Run

```shell
go build -trimpath -a -o ./bin/multiapps \
  -ldflags "-X github.com/lnashier/goarc/x/buildinfo.Version=$(cat VERSION) \
            -X github.com/lnashier/goarc/x/buildinfo.Hash=$(git rev-parse HEAD)" .
./bin/multiapps
```

# Run

```shell
go run .
```
