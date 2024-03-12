package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	chttp "github.com/lnashier/goarc/http"
	"github.com/lnashier/goarc/x/config"
	"github.com/lnashier/goarc/x/env"
	"github.com/lnashier/goarc/x/zson"
	"net/http"
	"time"
)

func main() {
	cfg := GetConfig()

	server := chttp.NewServer(
		chttp.ServerName(cfg.GetString("name")),
		chttp.ServerPort(cfg.GetInt("server.port")),
		chttp.ServerShutdownGracetime(time.Duration(cfg.GetInt("server.shutdown.gracetime"))*time.Second),
		chttp.App(
			func(srv *chttp.Server) error {
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
			},
		),
	)

	chttp.ServerUp(server)
}

func GetConfig() *config.Config {
	cfg, err := config.Loaded(config.NewCustomWatchedPath("./", env.Get().String(), func(e fsnotify.Event) {
		fmt.Printf("config file updated: %s\n", e.String())
	}))
	if err != nil {
		panic(fmt.Sprintf("failed to load app config: %v", err.Error()))
	}
	return cfg
}

type CustomHandler struct {
	ContentType string
	Route       chttp.Route
}

func (h *CustomHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	result, err := h.Route(req)
	if err != nil {
		chttp.HandleError(w, err)
		return
	}
	w.Header().Set("Content-Type", h.ContentType)
	w.WriteHeader(http.StatusOK)
	w.Write(result.([]byte))
}
