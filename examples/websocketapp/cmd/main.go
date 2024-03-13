package main

import (
	"github.com/lnashier/goarc"
	shttp "github.com/lnashier/goarc/http"
	"github.com/lnashier/goarc/x/config"
	"time"
	"websocketapp/internal/app"
)

func main() {
	cfg := config.Get()
	goarc.Up(
		shttp.NewService(
			shttp.ServiceName(cfg.GetString("name")),
			shttp.ServicePort(cfg.GetInt("server.port")),
			shttp.ServiceShutdownGracetime(time.Duration(cfg.GetInt("server.shutdown.gracetime"))*time.Second),
			shttp.App(app.App),
		),
	)
}
