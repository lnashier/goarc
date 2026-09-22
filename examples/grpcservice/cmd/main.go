package main

import (
	"fmt"
	"github.com/lnashier/goarc/v2"
	goarcgrpc "github.com/lnashier/goarc/v2/grpc"
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
