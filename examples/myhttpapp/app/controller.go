package app

import "github.com/lnashier/go-app/pkg/config"

type Controller struct {
	cfg   *config.Config
	store map[string]string
}

func NewController(cfg *config.Config) (*Controller, error) {
	return &Controller{
		cfg:   cfg,
		store: make(map[string]string),
	}, nil
}
