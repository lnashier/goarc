package cli_test

import (
	"context"
	"fmt"
	"time"

	"github.com/lnashier/goarc/v2"
	goarccli "github.com/lnashier/goarc/v2/cli"
)

// Example shows a cli.Service composed with goarc.Run: ServiceArgs selects
// which command runs (in a real process this would default to os.Args),
// Register wires it up, and Run drives the full Start/Stop lifecycle —
// here stopping itself after a short timeout instead of waiting for an OS
// signal. The registered command honors ctx being done, which is what
// lets Run's shutdown actually take effect.
func Example() {
	svc := goarccli.NewService(
		goarccli.ServiceArgs([]string{"wait"}),
		goarccli.App(func(s *goarccli.Service) error {
			s.Register("wait", func(ctx context.Context, args []string) error {
				<-ctx.Done()
				fmt.Println("command stopped")
				return nil
			})
			return nil
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if err := goarc.Run(svc, goarc.Context(ctx)); err != nil {
		fmt.Println("error:", err)
		return
	}
	// Output:
	// command stopped
}
