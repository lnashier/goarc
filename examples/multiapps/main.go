package main

import (
	"github.com/lnashier/goarc"
	shttp "github.com/lnashier/goarc/http"
	"github.com/lnashier/goarc/x/buildinfo"
	"github.com/lnashier/goarc/x/health"
	"multiapps/apps/echo"
	"multiapps/apps/hello"
	"time"
)

func main() {
	goarc.Up(
		shttp.NewService(
			shttp.ServiceName("multiapps"),
			shttp.ServicePort(8080),
			shttp.ServiceShutdownGracetime(time.Duration(1)*time.Second),
			shttp.App(hello.App, echo.App, health.App, buildinfo.App),
		),
	)
}
