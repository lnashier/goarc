package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/lnashier/goarc/v2"
	goarchttp "github.com/lnashier/goarc/v2/http"
	"github.com/lnashier/goarc/v2/x/buildinfo"
	"github.com/lnashier/goarc/v2/x/config"
	"github.com/lnashier/goarc/v2/x/env"
	"github.com/lnashier/goarc/v2/x/health"
	xhttp "github.com/lnashier/goarc/v2/x/http"
	xjson "github.com/lnashier/goarc/v2/x/json"
	"net/http"
	"time"
)

func main() {
	cfg := GetConfig()

	service := goarchttp.NewService(
		goarchttp.ServiceName(cfg.GetString("name")),
		goarchttp.ServicePort(cfg.GetInt("server.port")),
		goarchttp.ServiceShutdownGracetime(time.Duration(cfg.GetInt("server.shutdown.gracetime"))*time.Second),
		goarchttp.App(
			health.App,
			buildinfo.App,
			func(srv *goarchttp.Service) error {
				srv.Register(
					"/examples",
					http.MethodPost,
					&CustomHandler{
						ContentType: "application/json; charset=UTF-8",
						Route: func(req *http.Request) (any, error) {
							return xjson.Marshal(map[string]string{
								"msgId": "mock-msg-id",
							}), nil
						},
					},
				)

				srv.Register(
					"/examples",
					http.MethodGet,
					&CustomHandler{
						ContentType: "text/plain; charset=UTF-8",
						Route: func(req *http.Request) (any, error) {
							return []byte("mock-data"), nil
						},
					},
				)

				return nil
			},
		),
	)

	goarc.Up(service)
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
	Route       func(*http.Request) (any, error)
}

func (h *CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := h.Route(r)
	if err != nil {
		_, _ = xhttp.ConvertError(err).WriteJSON(w)
		return
	}
	w.Header().Set("Content-Type", h.ContentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result.([]byte))
}
