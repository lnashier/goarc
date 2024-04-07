package main

import (
	"github.com/lnashier/goarc"
	goarchttp "github.com/lnashier/goarc/http"
	xhttp "github.com/lnashier/goarc/x/http"
	"net/http"
	"time"
)

func main() {
	goarc.Up(goarchttp.NewService(
		goarchttp.ServiceName("toy"),
		goarchttp.ServicePort(8080),
		goarchttp.ServiceShutdownGracetime(2*time.Second),
		goarchttp.App(func(srv *goarchttp.Service) error {

			// BYO http.Handler
			srv.Register("/toys/1", http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello World!"))
			}))

			// Use pre-assembled http.Handler to work with JSON response type
			srv.Register("/toys/2", http.MethodGet, xhttp.JSONHandler(func(r *http.Request) (any, error) {
				return []string{"Hello World!"}, nil
			}))

			return nil
		}),
	))
}
