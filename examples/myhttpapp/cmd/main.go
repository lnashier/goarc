package main

import (
	"github.com/lnashier/go-app/config"
	"github.com/lnashier/go-app/examples/myhttpapp/app"
	"github.com/lnashier/go-app/http/service"
)

func main() {
	service.Up(service.Build(
		service.WithConfig(config.Get()),
		service.WithApp(app.App),
	))
}
