# Idiomatic Go Applications Framework

[![GoDoc](https://pkg.go.dev/badge/github.com/lnashier/goarc)](https://pkg.go.dev/github.com/lnashier/goarc)
[![Go Report Card](https://goreportcard.com/badge/github.com/lnashier/goarc)](https://goreportcard.com/report/github.com/lnashier/goarc)

The goarc is an idiomatic Go applications framework that focuses on application Life-Cycle.

## Installation

Simply add the following import to your code, and then `go [build|run|test]` will automatically fetch the necessary
dependencies:

```go
import "github.com/lnashier/goarc"
```

## Examples

[Examples](examples/)

## Toy HTTP Example

```go
package main

import (
	"github.com/lnashier/goarc"
	goarchttp "github.com/lnashier/goarc/http"
	xhttp "github.com/lnashier/goarc/x/http"
	"net/http"
	"time"
)

func main() {
	goarc.Up(goarchttp.NewService(
		goarchttp.ServiceName("toy"),
		goarchttp.ServicePort(8080),
		goarchttp.ServiceShutdownGracetime(2*time.Second),
		goarchttp.App(func(srv *goarchttp.Service) error {

			// BYO http.Handler
			srv.Register("/toys/byo", http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Hello World!"))
			}))

			// Use pre-assembled http.Handler to work with JSON response type
			srv.Register("/toys/json", http.MethodGet, xhttp.JSONHandler(func(r *http.Request) (any, error) {
				return []string{"Hello World!"}, nil
			}))

			// Use pre-assembled http.Handler to work with TEXT response type
			srv.Register("/toys/text", http.MethodGet, xhttp.TextHandler(func(r *http.Request) (string, error) {
				return "Hello World!", nil
			}))

			return nil
		}),
	))
}
```


## Toy CLI Example

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/lnashier/goarc"
	goarccli "github.com/lnashier/goarc/cli"
	xtime "github.com/lnashier/goarc/x/time"
	"time"
)

func main() {
	goarc.Up(goarccli.NewService(
		goarccli.ServiceName("mockcli"),
		goarccli.App(
			func(svc *goarccli.Service) error {
				svc.Register("echo", func(ctx context.Context, args []string) error {
					xtime.SleepWithContext(ctx, time.Duration(10)*time.Second)

					if len(args) > 0 {
						fmt.Println(args[0])
						return nil
					}

					return errors.New("nothing to echo")
				})

				return nil
			},
		),
	))
}
```
