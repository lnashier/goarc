package main

import (
	"github.com/lnashier/goarc"
	goarchttp "github.com/lnashier/goarc/http"
	"github.com/lnashier/goarc/x/health"
	"httpservice/internal/app"
	"net/http"
	"time"
)

func main() {
	goarc.Up(
		goarchttp.NewService(
			goarchttp.ServiceName("httpservice"),
			goarchttp.ServicePort(8080),
			goarchttp.ServiceShutdownGracetime(time.Duration(10)*time.Second),
			goarchttp.App(
				app.App,
				// Register health endpoints
				func(srv *goarchttp.Service) error {
					ctr := health.New()
					// Register health controller for shutdown signal
					srv.Component(ctr)
					// Custom endpoints
					srv.Register("/alive", http.MethodGet, http.HandlerFunc(ctr.LiveHandler))
					srv.Register("/ready", http.MethodGet, http.HandlerFunc(ctr.ReadyHandler))
					return nil
				},
			),
		),
	)
}
