package app

import (
	"github.com/lnashier/goarc/grpc"
	"grpcservice/internal/app/echo"
)

func App(srv *grpc.Service) error {
	echo.Register(srv)
	return nil
}
