package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/lnashier/go-app/pkg/cli"
	"github.com/lnashier/go-app/pkg/config"
	"github.com/lnashier/go-app/pkg/log"
)

func main() {
	cli.Up(cli.Build(
		cli.WithConfig(func() *config.Config {
			cfg := config.New()
			cfg.Set("name", "encoder")
			cfg.Set("log.app.debug", false)
			cfg.Set("log.app.info", false)
			cfg.Set("log.app.error", false)
			return cfg
		}()),
		cli.WithApp(func(cfg *config.Config, svc *cli.Service) error {
			svc.Register("base64encode", func(ctx context.Context, args []string) error {
				log.Info("base64encode enter")
				defer log.Info("base64encode exit")

				if len(args) > 0 {
					fmt.Println(base64.StdEncoding.EncodeToString([]byte(args[0])))
					return nil
				}

				return errors.New("nothing to encode")
			})

			svc.Register("base64decode", func(ctx context.Context, args []string) error {
				log.Info("base64decode enter")
				defer log.Info("base64decode exit")

				if len(args) > 0 {
					str, err := base64.StdEncoding.DecodeString(args[0])
					if err != nil {
						return err
					}
					fmt.Println(string(str))
					return nil
				}

				return errors.New("nothing to decode")
			})

			return nil
		}),
	))
}
