package app

import (
	"github.com/lnashier/go-app/config"
	"github.com/lnashier/go-app/http/handler"
	"github.com/lnashier/go-app/http/service"
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
