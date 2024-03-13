package app

import (
	shttp "github.com/lnashier/goarc/http"
	xhttp "github.com/lnashier/goarc/x/http"
	"net/http"
)

func App(srv *shttp.Service) error {
	ctr, err := NewController()
	if err != nil {
		return err
	}

	srv.Register("/examples", http.MethodPost, &xhttp.JSONHandler{Route: ctr.SaveExample})
	srv.Register("/example/{id}", http.MethodGet, &xhttp.TextHandler{Route: ctr.GetExample})

	return nil
}
