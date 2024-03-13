package main

import (
	"github.com/lnashier/goarc"
	shttp "github.com/lnashier/goarc/http"
	"httpapp/internal/app"
	"time"
)

func main() {
	goarc.Up(
		shttp.NewService(
			shttp.ServiceName("httpapp"),
			shttp.ServicePort(8080),
			shttp.ServiceShutdownGracetime(time.Duration(1)*time.Second),
			shttp.App(app.App),
		),
	)
}
