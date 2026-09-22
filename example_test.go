package goarc_test

import (
	"context"
	"fmt"
	"time"

	"github.com/lnashier/goarc/v2"
)

func Example() {
	svc := goarc.Func{
		StartFunc: func(ctx context.Context) error {
			fmt.Println("started")
			<-ctx.Done()
			fmt.Println("stopping")
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := goarc.Run(svc, goarc.Context(ctx)); err != nil {
		fmt.Println("error:", err)
	}

	// Output:
	// started
	// stopping
}

func ExampleGroup() {
	g := goarc.NewGroup()
	g.Add("a", goarc.Func{StartFunc: func(ctx context.Context) error {
		<-ctx.Done()
		fmt.Println("a stopped")
		return nil
	}})
	g.Add("b", goarc.Func{StartFunc: func(ctx context.Context) error {
		<-ctx.Done()
		fmt.Println("b stopped")
		return nil
	}})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := g.Start(ctx); err != nil {
		fmt.Println("error:", err)
	}

	// Unordered output:
	// a stopped
	// b stopped
}
