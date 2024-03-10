package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/lnashier/goarc/config"
	"github.com/lnashier/goarc/env"
	chandler "github.com/lnashier/goarc/http/handler"
	"github.com/lnashier/goarc/http/service"
	"github.com/lnashier/goarc/log"
	"github.com/lnashier/goarc/zson"
	"net/http"
)

func main() {
	service.Up(service.Build(
		service.WithConfig(GetConfig()),
		service.WithApp(func(cfg *config.Config, srv *service.Server) error {
			srv.Register(
				"/examples",
				http.MethodPost,
				&CustomHandler{
					"application/json; charset=UTF-8",
					func(req *http.Request) (any, error) {
						return zson.Marshal(map[string]string{
							"msgId": "mock-msg-id",
						}), nil
					},
				},
			)

			srv.Register(
				"/examples",
				http.MethodGet,
				&CustomHandler{
					"text/plain; charset=UTF-8",
					func(req *http.Request) (any, error) {
						return []byte("mock-data"), nil
					},
				},
			)

			return nil
		}),
	))
}

func GetConfig() *config.Config {
	cfg, err := config.Loaded(config.NewCustomWatchedPath("./", env.Get().String(), func(e fsnotify.Event) {
		log.Info("config file updated: %s", e.String())
	}))
	if err != nil {
		panic(fmt.Sprintf("failed to load app config: %v", err.Error()))
	}
	return cfg
}

type CustomHandler struct {
	ContentType string
	Route       chandler.Route
}

func (h *CustomHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	result, err := h.Route(req)
	if err != nil {
		chandler.HandleError(w, err)
		return
	}
	w.Header().Set("Content-Type", h.ContentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result.([]byte))
}
