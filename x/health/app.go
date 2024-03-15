package health

import (
	shttp "github.com/lnashier/goarc/http"
	"net/http"
)

// App configures health-related endpoints /alive and /ready for the given service.
// To customize the endpoints, get a New Controller and register the endpoints with custom names.
//
// See Controller.Stop, Controller.Live and Controller.Ready
func App(srv *shttp.Service) error {
	ctr := New()
	srv.Component(ctr)
	srv.Register("/alive", http.MethodGet, http.HandlerFunc(ctr.Live))
	srv.Register("/ready", http.MethodGet, http.HandlerFunc(ctr.Ready))
	return nil
}
