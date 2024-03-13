package main

import (
	"github.com/lnashier/goarc"
	"github.com/lnashier/goarc/grpc"
	"grpcservice/internal/app"
)

func main() {
	srv := grpc.NewService(grpc.ServicePort(5001))
	err := app.App(srv)
	if err != nil {
		return
	}
	goarc.Up(srv)
}
