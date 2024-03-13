package main

import (
	"github.com/lnashier/goarc"
	shttp "github.com/lnashier/goarc/http"
	"github.com/lnashier/goarc/x/health"
	"httpservice/internal/app"
	"net/http"
	"time"
)

func main() {
	goarc.Up(
		shttp.NewService(
			shttp.ServiceName("httpservice"),
			shttp.ServicePort(8080),
			shttp.ServiceShutdownGracetime(time.Duration(10)*time.Second),
			shttp.App(
				app.App,
				// Register health endpoints
				func(srv *shttp.Service) error {
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
