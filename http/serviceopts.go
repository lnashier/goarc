package http

import "time"

type ServiceOpt func(*serviceOpts)

var defaultServiceOpts = serviceOpts{
	name:              "unknown",
	port:              8080,
	shutdownGracetime: time.Duration(1) * time.Second,
	apps:              []func(*Service) error{},
}

type serviceOpts struct {
	name              string
	port              int
	shutdownGracetime time.Duration
	apps              []func(*Service) error
}

func (s *serviceOpts) apply(opt ...ServiceOpt) {
	for _, o := range opt {
		o(s)
	}
}

func ServiceName(name string) ServiceOpt {
	return func(s *serviceOpts) {
		s.name = name
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
