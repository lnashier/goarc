package main

import (
	"github.com/lnashier/goarc"
	goarchttp "github.com/lnashier/goarc/http"
	"github.com/lnashier/goarc/x/buildinfo"
	"github.com/lnashier/goarc/x/health"
	"multiapps/apps/echo"
	"multiapps/apps/hello"
	"time"
)

func main() {
	goarc.Up(
		goarchttp.NewService(
			goarchttp.ServiceName("multiapps"),
			goarchttp.ServicePort(8080),
			goarchttp.ServiceShutdownGracetime(time.Duration(1)*time.Second),
			goarchttp.App(health.App, buildinfo.App, hello.App, echo.App),
		),
	)
}
