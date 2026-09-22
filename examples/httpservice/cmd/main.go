package main

import (
	"fmt"
	"github.com/lnashier/goarc/v2"
	goarchttp "github.com/lnashier/goarc/v2/http"
	"github.com/lnashier/goarc/v2/x/health"
	xhttp "github.com/lnashier/goarc/v2/x/http"
	"httpservice/internal/app"
	"net/http"
	"time"
)

func main() {
	goarc.Up(
		goarchttp.NewService(
			goarchttp.ServiceName("httpservice"),
			goarchttp.ServicePort(8080),
			goarchttp.ServiceShutdownGracetime(time.Duration(1)*time.Second),
			goarchttp.App(
				app.App,
				func(srv *goarchttp.Service) error {
					ctr := health.New()

					// Register health controller for shutdown signal
					srv.Component(ctr)

					// Custom health endpoints
					srv.Register("/alive", http.MethodGet, xhttp.TextHandler(func(r *http.Request) (string, error) {
						return ctr.Live(), nil
					}))

					srv.Register("/ready", http.MethodGet, xhttp.JSONHandler(func(r *http.Request) (any, error) {
						status, err := ctr.Ready()
						return map[string]string{
							"status": status,
						}, err
					}))

					return nil
				},
			),
		),
		goarc.OnStart(func(err error) {
			fmt.Println("Start Err: ", err)
		}),
		goarc.OnStop(func(err error) {
			fmt.Println("Stop Err: ", err)
		}),
	)
}
