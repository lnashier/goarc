package app

import (
	goarchttp "github.com/lnashier/goarc/http"
	xhttp "github.com/lnashier/goarc/x/http"
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
