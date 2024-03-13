package main

import (
	"github.com/lnashier/goarc"
	"github.com/lnashier/goarc/http"
	"httpapp/internal/app"
	"time"
)

func main() {
	goarc.Up(
		http.NewService(
			http.ServiceName("httpapp"),
			http.ServicePort(8080),
			http.ServiceShutdownGracetime(time.Duration(1)*time.Second),
			http.App(app.App),
		),
	)
}
