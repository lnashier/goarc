package main

import (
	"context"
	"fmt"
	"github.com/lnashier/goarc"
	xtime "github.com/lnashier/goarc/x/time"
	"time"
)

type Service struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *Service) Start() error {
	fmt.Println("Starting service")
	defer fmt.Println("Service done!")

	s.ctx, s.cancel = context.WithCancel(context.Background())

	fmt.Println("Doing some random work")
	xtime.SleepWithContext(s.ctx, time.Duration(10)*time.Second)
	fmt.Println("Done with random work")

	return nil
}

func (s *Service) Stop() error {
	fmt.Println("Stopping service")
	s.cancel()
	return nil
}

func main() {
	goarc.Up(&Service{})
}
