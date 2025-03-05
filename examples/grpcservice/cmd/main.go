package main

import (
	"fmt"
	"github.com/lnashier/goarc"
	goarcgrpc "github.com/lnashier/goarc/grpc"
	"grpcservice/internal/app"
)

func main() {
	goarc.Up(
		goarcgrpc.NewService(
			goarcgrpc.ServicePort(5001),
			goarcgrpc.App(app.App),
		),
		goarc.OnStart(func(err error) {
			if err != nil {
				fmt.Println(err)
			}
		}),
	)
}
