package cli

type ServiceOpt func(*serviceOpts)

var defaultServiceOpts = serviceOpts{
	name: "unknown",
	apps: []func(*Service) error{},
}

type serviceOpts struct {
	name string
	apps []func(*Service) error
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

func App(app ...func(*Service) error) ServiceOpt {
	return func(s *serviceOpts) {
		s.apps = append(s.apps, app...)
	}
}
