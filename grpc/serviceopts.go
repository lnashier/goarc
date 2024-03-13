package grpc

import (
	"time"
)

type ServiceOpt func(*serviceOpts)

var defaultServiceOpts = serviceOpts{
	name:              "unknown",
	network:           "tcp",
	port:              8080,
	shutdownGracetime: time.Duration(1) * time.Second,
	apps:              []func(service *Service) error{},
}

type serviceOpts struct {
	name              string
	network           string
	port              int
	shutdownGracetime time.Duration
	apps              []func(*Service) error
}

func (s *serviceOpts) apply(opts []ServiceOpt) {
	for _, o := range opts {
		o(s)
	}
}

func ServiceName(name string) ServiceOpt {
	return func(s *serviceOpts) {
		s.name = name
	}
}

func ServiceNetwork(network string) ServiceOpt {
	return func(s *serviceOpts) {
		s.network = network
	}
}

func ServicePort(port int) ServiceOpt {
	return func(s *serviceOpts) {
		s.port = port
	}
}

func ServiceShutdownGracetime(t time.Duration) ServiceOpt {
	return func(s *serviceOpts) {
		s.shutdownGracetime = t
	}
}

func App(app ...func(*Service) error) ServiceOpt {
	return func(s *serviceOpts) {
		s.apps = append(s.apps, app...)
	}
}
