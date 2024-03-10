package app

import (
	"github.com/lnashier/goarc/config"
	"github.com/lnashier/goarc/http/handler"
	"github.com/lnashier/goarc/http/service"
	"net/http"
)

func App(cfg *config.Config, srv *service.Server) error {
	ctr, err := NewController(cfg)
	if err != nil {
		return err
	}

	srv.Register("/examples", http.MethodPost, &handler.JSON{Route: ctr.SaveExample})
	srv.Register("/example/{id}", http.MethodGet, &handler.Text{Route: ctr.GetExample})

	return nil
}
