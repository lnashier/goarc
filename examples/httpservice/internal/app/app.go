package app

import (
	goarchttp "github.com/lnashier/goarc/v2/http"
	xhttp "github.com/lnashier/goarc/v2/x/http"
	"net/http"
)

func App(srv *goarchttp.Service) error {
	ctr, err := NewController()
	if err != nil {
		return err
	}

	srv.Register("/examples", http.MethodPost, xhttp.JSONHandler(ctr.SaveExample))
	srv.Register("/examples/{id}", http.MethodGet, xhttp.JSONHandler(ctr.GetExample))

	return nil
}
