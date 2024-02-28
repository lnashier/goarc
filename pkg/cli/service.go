package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lnashier/go-app/pkg/buildinfo"
	"github.com/lnashier/go-app/pkg/config"
	"github.com/lnashier/go-app/pkg/log"
	"github.com/spf13/cobra"
)

type Service struct {
	name    string
	cfg     *config.Config
	rootCmd *cobra.Command
	exitCh  chan struct{}
}

func NewService(cfg *config.Config) *Service {
	rootCmd := &cobra.Command{
		Use:   "Root",
		Short: "CLI Service",
		Run: func(*cobra.Command, []string) {
			log.Info("Provide APP specific command")
		},
	}

	return &Service{
		name:    cfg.GetString("name"),
		cfg:     cfg,
		rootCmd: rootCmd,
	}
}

func (s *Service) Start() error {
	log.Info("Service#Start enter")
	defer log.Info("Service#Start exit")

	s.exitCh = make(chan struct{})

	s.rootCmd.AddCommand(&cobra.Command{
		Use:           "buildinfo",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := json.MarshalIndent(buildinfo.New(func() buildinfo.Report {
				return buildinfo.Report{
					buildinfo.KeyAppName: s.name,
					buildinfo.KeyVersion: buildinfo.Version,
					buildinfo.KeyHash:    buildinfo.Hash,
				}
			}).Report(), "", "  ")
			fmt.Println(string(report))
			return err
		},
	})

	if err := s.rootCmd.Execute(); err != nil {
		log.Error("Service#Start failed to execute command: %v", err)
		return err
	}
	return nil
}

func (s *Service) Stop() error {
	log.Info("Service#Stop enter")
	defer log.Info("Service#Stop exit")
	s.exitCh <- struct{}{}
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
