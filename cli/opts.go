package cli

type Opt func(*opts)

var defaultOpts = opts{
	name: "unknown",
	apps: []func(*Service) error{},
}

type opts struct {
	name string
	apps []func(*Service) error
}

func (s *opts) apply(opts []Opt) {
	for _, o := range opts {
		o(s)
	}
}

func ServiceName(name string) Opt {
	return func(s *opts) {
		s.name = name
	}
}

func App(app ...func(*Service) error) Opt {
	return func(s *opts) {
		s.apps = append(s.apps, app...)
	}
}
