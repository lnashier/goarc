package hello

import (
	goarchttp "github.com/lnashier/goarc/http"
	xhttp "github.com/lnashier/goarc/x/http"
	"net/http"
)

func App(srv *goarchttp.Service) error {
	srv.Register("/hello", http.MethodGet, xhttp.JSONHandler(func(req *http.Request) (any, error) {
		return map[string]string{"message": "Hello! World"}, nil
	}))

	return nil
}
