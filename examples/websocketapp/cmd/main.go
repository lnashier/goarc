package main

import (
	"github.com/lnashier/goarc"
	"github.com/lnashier/goarc/http"
	"github.com/lnashier/goarc/x/config"
	"time"
	"websocketapp/internal/app"
)

func main() {
	cfg := config.Get()
	goarc.Up(
		http.NewService(
			http.ServiceName(cfg.GetString("name")),
			http.ServicePort(cfg.GetInt("server.port")),
			http.ServiceShutdownGracetime(time.Duration(cfg.GetInt("server.shutdown.gracetime"))*time.Second),
			http.App(app.App),
		),
	)
}
