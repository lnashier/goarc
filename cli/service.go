package cli

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/spf13/cobra"
)

// Service is a CLI goarc.Service, backed by cobra.
//
// Unlike http.Service or grpc.Service, Start does not own the blocking
// call itself — it blocks on whatever registered command os.Args (or
// ServiceArgs) dispatches to, which is user code. Start propagates ctx
// cancellation and Stop into that command's own context immediately, but
// genuinely returning early depends on the running command actually
// honoring its ctx, the same contract Register has always documented for
// its runner parameter.
type Service struct {
	opts    serviceOpts
	rootCmd *cobra.Command

	mu     sync.Mutex
	cmdCtx context.Context

	stopped  chan struct{}
	stopOnce sync.Once
}

func NewService(opt ...ServiceOpt) *Service {
	opts := defaultServiceOpts
	opts.apply(opt...)

	rootCmd := &cobra.Command{
		Use:   "Root",
		Short: "CLI Service",
		RunE: func(*cobra.Command, []string) error {
			return errors.New("provide APP specific command")
		},
	}
	if opts.args != nil {
		rootCmd.SetArgs(opts.args)
	}

	s := &Service{
		opts:    opts,
		rootCmd: rootCmd,
		stopped: make(chan struct{}),
	}

	// Configure app(s); if provided
	for _, app := range opts.apps {
		if err := app(s); err != nil {
			panic(fmt.Sprintf("failed to configure app: %v", err))
		}
	}

	return s
}

// Start implements goarc.Service: it dispatches to whichever registered
// command os.Args (or ServiceArgs) selects, passing that command a context
// derived from ctx — so canceling ctx cancels the running command's
// context too. Start returns once that command returns.
//
// If Stop was already called before Start, Start returns immediately
// without dispatching any command.
func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	s.cmdCtx = ctx
	s.mu.Unlock()

	select {
	case <-s.stopped:
		return nil
	default:
	}

	return s.rootCmd.Execute()
}

// Stop implements goarc.Service: it cancels the context of whatever
// command is currently running (or about to run), asking it to return.
// Stop does not wait for that command to actually finish — as with any
// goarc.Service, that is the command's own responsibility, by honoring its
// ctx promptly.
//
// Stop is safe to call before Start, concurrently with a running Start, or
// more than once.
func (s *Service) Stop(context.Context) error {
	s.stopOnce.Do(func() {
		close(s.stopped)
	})
	return nil
}

// Register registers a new command with the CLI service. runner must
// honor ctx being done as a request to return promptly — Start (via ctx)
// and Stop both signal shutdown by canceling it.
func (s *Service) Register(cmd string, runner func(ctx context.Context, args []string) error) {
	s.rootCmd.AddCommand(&cobra.Command{
		Use:           cmd,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, args []string) error {
			s.mu.Lock()
			parent := s.cmdCtx
			s.mu.Unlock()
			if parent == nil {
				parent = context.Background()
			}

			ctx, cancel := context.WithCancel(parent)
			defer cancel()

			done := make(chan struct{})
			defer close(done)
			go func() {
				select {
				case <-s.stopped:
					cancel()
				case <-ctx.Done():
				case <-done:
				}
			}()

			return runner(ctx, args)
		},
	})
}
