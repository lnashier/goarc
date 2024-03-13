package main

import (
	"github.com/lnashier/goarc"
	sgrpc "github.com/lnashier/goarc/grpc"
	"grpcservice/internal/app"
)

func main() {
	goarc.Up(sgrpc.NewService(
		sgrpc.ServicePort(5001),
		sgrpc.App(app.App),
	))
}
