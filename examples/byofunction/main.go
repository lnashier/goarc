package main

import (
	"context"
	"fmt"
	"github.com/lnashier/goarc"
	xtime "github.com/lnashier/goarc/x/time"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	goarc.Up(goarc.ServiceFunc(
		// The same function is invoked for both starting and stopping the service.
		// It's important to note that Start and Stop represent two separate executions
		// of the same provided function.
		// Any local variables won't persist across these executions as expected.
		func(start bool) error {
			if !start {
				fmt.Println("Stopping service")
				cancel()
				return nil
			}

			fmt.Println("Starting service")
			defer fmt.Println("Service done!")

			fmt.Println("Doing some random work")
			xtime.SleepWithContext(ctx, time.Duration(10)*time.Second)
			fmt.Println("Done with random work")

			return nil
		},
	))
}
