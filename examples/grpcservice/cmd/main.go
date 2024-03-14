package main

import (
	"github.com/lnashier/goarc"
	goarcgrpc "github.com/lnashier/goarc/grpc"
	"grpcservice/internal/app"
)

func main() {
	goarc.Up(goarcgrpc.NewService(
		goarcgrpc.ServicePort(5001),
		goarcgrpc.App(app.App),
	))
}
