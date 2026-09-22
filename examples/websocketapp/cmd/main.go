package main

import (
	"github.com/lnashier/goarc/v2"
	goarchttp "github.com/lnashier/goarc/v2/http"
	"github.com/lnashier/goarc/v2/x/buildinfo"
	"github.com/lnashier/goarc/v2/x/config"
	"github.com/lnashier/goarc/v2/x/health"
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
