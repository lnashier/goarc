package main

import (
	"github.com/lnashier/goarc"
	goarchttp "github.com/lnashier/goarc/http"
	"github.com/lnashier/goarc/x/buildinfo"
	"github.com/lnashier/goarc/x/config"
	"github.com/lnashier/goarc/x/health"
	"time"
	"websocketapp/internal/app"
)

func main() {
	cfg := config.Get()
	goarc.Up(
		goarchttp.NewService(
			goarchttp.ServiceName(cfg.GetString("name")),
			goarchttp.ServicePort(cfg.GetInt("server.port")),
			goarchttp.ServiceShutdownGracetime(time.Duration(cfg.GetInt("server.shutdown.gracetime"))*time.Second),
			goarchttp.App(health.App, buildinfo.App, app.App),
		),
	)
}
