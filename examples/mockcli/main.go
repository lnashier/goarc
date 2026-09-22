package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/lnashier/goarc/v2"
	goarccli "github.com/lnashier/goarc/v2/cli"
	xtime "github.com/lnashier/goarc/v2/x/time"
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

					// SleepWithContext returns early on cancellation too, so
					// check ctx before treating this as "the 10s elapsed".
					// A canceled command is a clean stop, not a failure —
					// same convention goarc's own Service implementations
					// follow — so it returns nil, not ctx.Err().
					if ctx.Err() != nil {
						fmt.Println("echo canceled before its 10s were up")
						return nil
					}

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
