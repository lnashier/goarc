package main

import (
	"context"
	"fmt"
	"github.com/lnashier/goarc/v2"
	xtime "github.com/lnashier/goarc/v2/x/time"
	"time"
)

func main() {
	goarc.Up(
		goarc.Func{
			StartFunc: func(ctx context.Context) error {
				fmt.Println("Starting service")
				defer fmt.Println("Service done!")

				fmt.Println("Doing some random work for 10 seconds")
				xtime.SleepWithContext(ctx, time.Duration(10)*time.Second)
				fmt.Println("Done with random work")

				return nil
			},
			StopFunc: func(context.Context) error {
				fmt.Println("Stopping service")
				return nil
			},
		},
		goarc.OnStart(func(err error) {
			fmt.Println("On Start Err: ", err)
		}),
		goarc.OnStop(func(err error) {
			fmt.Println("On Stop Err: ", err)
		}),
	)
}
