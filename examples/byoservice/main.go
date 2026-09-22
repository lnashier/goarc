package main

import (
	"context"
	"fmt"
	"time"

	"github.com/lnashier/goarc/v2"
	xtime "github.com/lnashier/goarc/v2/x/time"
)

type Service struct{}

// Start must self-terminate when ctx is done — here that's automatic,
// since the only blocking call is SleepWithContext, which already returns
// as soon as ctx is done.
func (s *Service) Start(ctx context.Context) error {
	fmt.Println("Starting service")
	defer fmt.Println("Service done!")

	fmt.Println("Doing some random work")
	xtime.SleepWithContext(ctx, time.Duration(10)*time.Second)
	fmt.Println("Done with random work")

	return nil
}

func (s *Service) Stop(context.Context) error {
	fmt.Println("Stopping service")
	return nil
}

func main() {
	goarc.Up(&Service{})
}
