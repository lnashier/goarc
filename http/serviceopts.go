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

// ServiceShutdownGracetime bounds the shutdown that Start triggers on its
// own when ctx is done: it becomes the deadline for stopping every
// registered Component and for http.Server.Shutdown, i.e. how long
// in-flight requests are given to finish before being forcibly closed.
//
// It has no effect on a shutdown driven by an explicit call to Stop(ctx)
// with its own ctx — that ctx's deadline, if any, is used as-is.
//
// The default is 1 second.
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
