package health

import (
	shttp "github.com/lnashier/goarc/http"
	xhttp "github.com/lnashier/goarc/x/http"
	"net/http"
)

// App configures health-related endpoints /alive and /ready for the given service.
// To customize the endpoints, get a New Controller and register the endpoints with custom names.
//
// See Controller.Stop, Controller.Live and Controller.Ready
func App(srv *shttp.Service) error {
	ctr := New()
	srv.Component(ctr)
	srv.Register("/alive", http.MethodGet, xhttp.TextHandler(func(*http.Request) (string, error) {
		return ctr.Live(), nil
	}))
	srv.Register("/ready", http.MethodGet, xhttp.TextHandler(func(*http.Request) (string, error) {
		return ctr.Ready()
	}))
	return nil
}
