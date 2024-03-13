# Example GRPC App

```shell
protoc --go_out=. --go-grpc_out=. internal/proto/echo/*.proto
```

# Build & Run

```shell
go build -o ./bin/grpcservice ./cmd
./bin/grpcservice
```

# Run

```shell
go run ./cmd
```