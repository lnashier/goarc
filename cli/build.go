package cli

import (
	"fmt"
	"github.com/lnashier/goarc/log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func Build(opts ...Opt) *Service {
	svcOpts := defaultOpts()
	svcOpts.applyOptions(opts)
	cfg := svcOpts.cfg
	//logger := svcOpts.logger

	if len(cfg.GetString("name")) < 1 {
		log.Panic("cli#Build app name missing")
	}

	// Init logging
	log.DefaultLogger.AppName = cfg.GetString("name")
	log.DefaultLogger.Verifier = func(level log.Level) bool {
		return cfg.GetBool(fmt.Sprintf("log.app.%s", strings.ToLower(string(level))))
	}

	// Get cli service
	tl := NewService(cfg)

	// Configure app(s)
	for _, app := range svcOpts.apps {
		if err := app(cfg, tl); err != nil {
			log.Panic("cli#Build failed to configure app: %v", err)
		}
	}

	return tl
}

// Up is to boot up the cli Service
// Program listens for the following SIG:
//
//	syscall.SIGINT
//	syscall.SIGTERM
//	syscall.SIGQUIT
//	syscall.SIGABRT
//
// On success
func Up(svc *Service) {
	log.Info("cli@Up enter")
	defer log.Info("cli@Up exit")
	go func(svc *Service) {
		log.Info("cli@Up#signal enter")
		defer log.Info("cli@Up#signal exit")
		sic := make(chan os.Signal, 1)
		// e.g. kill -SIGQUIT <pid>
		signal.Notify(sic, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
		sig := <-sic
		log.Info("cli@Up#signal (%v) signal received to terminate", sig.String())
		if err := svc.Stop(); err != nil {
			os.Exit(1)
		}
	}(svc)
	if err := svc.Start(); err != nil {
		os.Exit(1)
	}
}
