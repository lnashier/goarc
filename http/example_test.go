package http_test

import (
	"context"
	"fmt"
	nethttp "net/http"
	"time"

	"github.com/lnashier/goarc/v2"
	goarchttp "github.com/lnashier/goarc/v2/http"
)

// Example shows an http.Service composed with goarc.Run: NewService binds
// no port until Start is called, Register wires up routes beforehand, and
// Run drives the full Start/Stop lifecycle — here stopping itself after a
// short timeout instead of waiting for an OS signal.
func Example() {
	svc := goarchttp.NewService(
		goarchttp.ServicePort(0), // 0: let the OS choose a free port
		goarchttp.App(func(s *goarchttp.Service) error {
			s.Register("/hello", nethttp.MethodGet, nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
				w.Write([]byte("world"))
			}))
			return nil
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if err := goarc.Run(svc, goarc.Context(ctx)); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("stopped cleanly")
	// Output:
	// stopped cleanly
}
