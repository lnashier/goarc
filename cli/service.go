package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

type Service struct {
	name    string
	rootCmd *cobra.Command
	exitCh  chan struct{}
}

func NewService(opt ...Opt) *Service {
	svcOpts := defaultOpts
	svcOpts.apply(opt)

	rootCmd := &cobra.Command{
		Use:   "Root",
		Short: "CLI Service",
		RunE: func(*cobra.Command, []string) error {
			return errors.New("provide APP specific command")
		},
	}

	s := &Service{
		name:    svcOpts.name,
		rootCmd: rootCmd,
	}

	// Configure app(s)
	for _, app := range svcOpts.apps {
		if err := app(s); err != nil {
			panic(fmt.Sprintf("failed to configure app: %v", err))
		}
	}

	return s
}

func (s *Service) Start() error {
	s.exitCh = make(chan struct{})
	return s.rootCmd.Execute()
}

func (s *Service) Stop() error {
	close(s.exitCh)
	return nil
}

func (s *Service) Register(cmd string, runner func(ctx context.Context, args []string) error) {
	s.rootCmd.AddCommand(&cobra.Command{
		Use:           cmd,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				<-s.exitCh
				cancel()
			}()
			return runner(ctx, args)
		},
	})
}
