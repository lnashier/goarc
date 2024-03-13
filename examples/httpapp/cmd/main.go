package main

import (
	"github.com/lnashier/goarc"
	"github.com/lnashier/goarc/http"
	"httpapp/internal/app"
	"time"
)

func main() {
	goarc.Up(
		http.NewServer(
			http.ServerName("httpapp"),
			http.ServerPort(8080),
			http.ServerShutdownGracetime(time.Duration(1)*time.Second),
			http.App(app.App),
		),
	)
}
