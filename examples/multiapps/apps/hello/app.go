package hello

import (
	shttp "github.com/lnashier/goarc/http"
	xhttp "github.com/lnashier/goarc/x/http"
	"net/http"
)

func App(srv *shttp.Service) error {
	srv.Register("/hello", http.MethodGet, &xhttp.JSONHandler{Route: func(req *http.Request) (any, error) {
		return map[string]string{"message": "Hello! World"}, nil
	}})

	return nil
}
