package app

import (
	chttp "github.com/lnashier/goarc/http"
	"net/http"
)

func App(srv *chttp.Server) error {
	ctr, err := NewController()
	if err != nil {
		return err
	}

	srv.Register("/examples", http.MethodPost, &chttp.JSONHandler{Route: ctr.SaveExample})
	srv.Register("/example/{id}", http.MethodGet, &chttp.TextHandler{Route: ctr.GetExample})

	return nil
}
