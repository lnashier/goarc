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
	host              string
	port              int
	readHeaderTimeout time.Duration
	readTimeout       time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	maxHeaderBytes    int
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

// ServiceHost sets the host or IP address the service binds to, e.g.
// "127.0.0.1" to listen on loopback only. The default is empty, which binds
// all interfaces.
func ServiceHost(host string) ServiceOpt {
	return func(s *serviceOpts) {
		s.host = host
	}
}

// ServiceReadHeaderTimeout sets http.Server.ReadHeaderTimeout, the time
// allowed to read request headers. Setting it is the main defense against
// slow-header (Slowloris) clients on a public listener. The default, zero,
// means no timeout.
func ServiceReadHeaderTimeout(d time.Duration) ServiceOpt {
	return func(s *serviceOpts) {
		s.readHeaderTimeout = d
	}
}

// ServiceReadTimeout sets http.Server.ReadTimeout, the time allowed to read
// the entire request including the body. The default, zero, means no timeout.
func ServiceReadTimeout(d time.Duration) ServiceOpt {
	return func(s *serviceOpts) {
		s.readTimeout = d
	}
}

// ServiceWriteTimeout sets http.Server.WriteTimeout, the time allowed to
// write the response. The default, zero, means no timeout. Do not set it on
// services that stream long-lived responses or upgrade connections
// (e.g. WebSockets).
func ServiceWriteTimeout(d time.Duration) ServiceOpt {
	return func(s *serviceOpts) {
		s.writeTimeout = d
	}
}

// ServiceIdleTimeout sets http.Server.IdleTimeout, how long an idle
// keep-alive connection is kept open. The default, zero, falls back to the
// read timeout.
func ServiceIdleTimeout(d time.Duration) ServiceOpt {
	return func(s *serviceOpts) {
		s.idleTimeout = d
	}
}

// ServiceMaxHeaderBytes sets http.Server.MaxHeaderBytes, the maximum size of
// request headers. The default, zero, uses Go's DefaultMaxHeaderBytes (1 MB).
func ServiceMaxHeaderBytes(n int) ServiceOpt {
	return func(s *serviceOpts) {
		s.maxHeaderBytes = n
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
