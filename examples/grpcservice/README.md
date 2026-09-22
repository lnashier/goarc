# Example gRPC App

A `grpc.Service` registering a protoc-generated `Echo` service that exercises all four gRPC call
shapes: `Single` (unary), `ServiceStream` (server-streaming), `ClientStream` (client-streaming),
and `BothStream` (bidirectional streaming). `Register(srv)` passes the `*grpc.Service` itself as
the `grpc.ServiceRegistrar` — the same trick works with any protoc-generated `Register*Server`
function, no goarc-specific glue needed.

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