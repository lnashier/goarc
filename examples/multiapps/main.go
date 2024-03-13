package main

import (
	"github.com/lnashier/goarc"
	"github.com/lnashier/goarc/examples/apps/echo"
	"github.com/lnashier/goarc/examples/apps/hello"
	shttp "github.com/lnashier/goarc/http"
	"time"
)

func main() {
	goarc.Up(
		shttp.NewService(
			shttp.ServiceName("multiapps"),
			shttp.ServicePort(8080),
			shttp.ServiceShutdownGracetime(time.Duration(1)*time.Second),
			shttp.App(hello.App, echo.App),
		),
	)
}
