package main

import (
	"github.com/lnashier/goarc/config"
	"github.com/lnashier/goarc/http/service"
	"websocketapp/internal/app"
)

func main() {
	service.Up(service.Build(
		service.WithConfig(config.Get()),
		service.WithApp(app.App),
	))
}
