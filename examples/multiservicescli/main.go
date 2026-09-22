package main

import (
	"context"
	"net/http"
	"time"

	"github.com/lnashier/goarc/v2"
	goarccli "github.com/lnashier/goarc/v2/cli"
	goarchttp "github.com/lnashier/goarc/v2/http"
	"github.com/lnashier/goarc/v2/x/buildinfo"
	"github.com/lnashier/goarc/v2/x/health"
	xhttp "github.com/lnashier/goarc/v2/x/http"
)

func main() {
	goarc.Up(goarccli.NewService(
		goarccli.ServiceName("multiservicescli"),
		goarccli.App(
			func(svc *goarccli.Service) error {
				svc.Register("service1", func(ctx context.Context, args []string) error {
					goarc.Up(
						goarchttp.NewService(
							goarchttp.ServiceName("service1"),
							goarchttp.ServicePort(8081),
							goarchttp.ServiceShutdownGracetime(2*time.Second),
							goarchttp.App(
								func(srv *goarchttp.Service) error {
									srv.Register("/service1/toys/1", http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
										w.WriteHeader(http.StatusOK)
										w.Write([]byte("Hello World from Service1!"))
									}))

									srv.Register("/service1/toys/2", http.MethodGet, xhttp.JSONHandler(func(r *http.Request) (any, error) {
										return []string{"Hello World from Service1!"}, nil
									}))

									return nil
								},
								buildinfo.App,
								health.App,
							),
						),
						goarc.Context(ctx),
					)

					return nil
				})

				svc.Register("service2", func(ctx context.Context, args []string) error {
					goarc.Up(
						goarchttp.NewService(
							goarchttp.ServiceName("service2"),
							goarchttp.ServicePort(8082),
							goarchttp.ServiceShutdownGracetime(2*time.Second),
							goarchttp.App(
								func(srv *goarchttp.Service) error {
									srv.Register("/service2/toys/1", http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
										w.WriteHeader(http.StatusOK)
										w.Write([]byte("Hello World from Service2!"))
									}))

									srv.Register("/service2/toys/2", http.MethodGet, xhttp.JSONHandler(func(r *http.Request) (any, error) {
										return []string{"Hello World from Service2!"}, nil
									}))

									return nil
								},
								buildinfo.App,
								health.App,
							),
						),
						goarc.Context(ctx),
					)

					return nil
				})
				return nil
			},
		),
	))
}
