package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/lnashier/goarc"
	goarccli "github.com/lnashier/goarc/cli"
	xtime "github.com/lnashier/goarc/x/time"
	"time"
)

func main() {
	goarc.Up(goarccli.NewService(
		goarccli.ServiceName("mockcli"),
		goarccli.App(
			func(svc *goarccli.Service) error {
				svc.Register("echo", func(ctx context.Context, args []string) error {
					fmt.Println("going to echo after 10 seconds")
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
