package main

import (
	"github.com/lnashier/goarc/v2"
	goarchttp "github.com/lnashier/goarc/v2/http"
	xhttp "github.com/lnashier/goarc/v2/x/http"
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
			srv.Register("/toys/byo", http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello World!"))
			}))

			// Use pre-assembled http.Handler to work with JSON response type
			srv.Register("/toys/json", http.MethodGet, xhttp.JSONHandler(func(r *http.Request) (any, error) {
				return []string{"Hello World!"}, nil
			}))

			// Use pre-assembled http.Handler to work with TEXT response type
			srv.Register("/toys/text", http.MethodGet, xhttp.TextHandler(func(r *http.Request) (string, error) {
				return "Hello World!", nil
			}))

			return nil
		}),
	))
}
