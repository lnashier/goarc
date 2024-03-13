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

func NewService(opt ...ServiceOpt) *Service {
	opts := defaultServiceOpts
	opts.apply(opt)

	rootCmd := &cobra.Command{
		Use:   "Root",
		Short: "CLI Service",
		RunE: func(*cobra.Command, []string) error {
			return errors.New("provide APP specific command")
		},
	}

	s := &Service{
		name:    opts.name,
		rootCmd: rootCmd,
	}

	for _, app := range opts.apps {
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

// Register registers a new command with the CLI service.
func (s *Service) Register(cmd string, runner func(ctx context.Context, args []string) error) {
	s.rootCmd.AddCommand(&cobra.Command{
		Use:           cmd,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set up a context with cancellation for the command.
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				// Listen for exit signal and cancel the context when received.
				<-s.exitCh
				cancel()
			}()
			return runner(ctx, args)
		},
	})
}
