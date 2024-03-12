package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/lnashier/goarc/cli"
)

func main() {
	cli.Up(cli.NewService(
		cli.ServiceName("encoder"),
		cli.App(
			func(svc *cli.Service) error {
				svc.Register("base64encode", func(ctx context.Context, args []string) error {
					if len(args) > 0 {
						fmt.Println(base64.StdEncoding.EncodeToString([]byte(args[0])))
						return nil
					}

					return errors.New("nothing to encode")
				})

				svc.Register("base64decode", func(ctx context.Context, args []string) error {
					if len(args) > 0 {
						str, err := base64.StdEncoding.DecodeString(args[0])
						if err != nil {
							return err
						}
						fmt.Println(string(str))
						return nil
					}

					return errors.New("nothing to decode")
				})

				return nil
			},
		),
	))
}
