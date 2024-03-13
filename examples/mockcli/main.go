package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/lnashier/goarc"
	"github.com/lnashier/goarc/cli"
	xtime "github.com/lnashier/goarc/x/time"
	"time"
)

func main() {
	goarc.Up(cli.NewService(
		cli.ServiceName("mockcli"),
		cli.App(
			func(svc *cli.Service) error {
				svc.Register("echo", func(ctx context.Context, args []string) error {
					xtime.SleepWithContext(ctx, time.Duration(10)*time.Second)

					if len(args) > 0 {
						fmt.Println(args[0])
						return nil
					}
					return errors.New("nothing to echo")
				})

				return nil
			},
		),
	))
}
