package service

import (
	"fmt"
	"github.com/lnashier/go-app/log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func Build(opts ...Opt) *Server {
	srvOpts := defaultOpts()
	srvOpts.applyOptions(opts)
	//logger := svcOpts.logger

	cfg := srvOpts.cfg

	if len(cfg.GetString("name")) < 1 {
		log.Panic("service#Build app name missing")
	}

	// Init logging
	log.DefaultLogger.AppName = cfg.GetString("name")
	log.DefaultLogger.Verifier = func(level log.Level) bool {
		return cfg.GetBool(fmt.Sprintf("log.app.%s", strings.ToLower(string(level))))
	}

	// Get server
	srv := NewServer(cfg)

	// Configure app(s)
	for _, app := range srvOpts.apps {
		if err := app(cfg, srv); err != nil {
			log.Panic("service#Build failed to configure app: %v", err)
		}
	}

	return srv
}

// Up is to boot the server and wait for shutdown. It is a blocking function.
func Up(srv *Server) {
	log.Info("service@Up enter")
	defer log.Info("service@Up exit")
	go func(srv *Server) {
		log.Info("service@Up#signal enter")
		defer log.Info("service@Up#signal exit")
		sic := make(chan os.Signal, 1)
		// e.g. kill -SIGQUIT <pid>
		signal.Notify(sic, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
		sig := <-sic
		log.Info("service@Up#signal (%v) signal received to terminate", sig.String())
		srv.Stop()
	}(srv)
	srv.Start()
}
