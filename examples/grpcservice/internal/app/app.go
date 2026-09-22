package app

import (
	goarcgrpc "github.com/lnashier/goarc/v2/grpc"
	"grpcservice/internal/app/echo"
)

func App(srv *goarcgrpc.Service) error {
	echo.Register(srv)
	return nil
}
