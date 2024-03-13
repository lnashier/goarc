package http

import "time"

type ServerOpt func(*serverOpts)

var defaultServerOpts = serverOpts{
	name: "unknown",
	port: 8080,
	apps: []func(*Server) error{},
}

type serverOpts struct {
	name              string
	port              int
	shutdownGracetime time.Duration
	apps              []func(*Server) error
}

func (s *serverOpts) apply(opts []ServerOpt) {
	for _, o := range opts {
		o(s)
	}
}

func ServerName(name string) ServerOpt {
	return func(s *serverOpts) {
		s.name = name
	}
}

func ServerPort(port int) ServerOpt {
	return func(s *serverOpts) {
		s.port = port
	}
}

func ServerShutdownGracetime(t time.Duration) ServerOpt {
	return func(s *serverOpts) {
		s.shutdownGracetime = t
	}
}

func App(app ...func(*Server) error) ServerOpt {
	return func(s *serverOpts) {
		s.apps = append(s.apps, app...)
	}
}
