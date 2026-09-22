package cli

type ServiceOpt func(*serviceOpts)

var defaultServiceOpts = serviceOpts{
	name: "unknown",
	apps: []func(*Service) error{},
}

type serviceOpts struct {
	name string
	args []string
	apps []func(*Service) error
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

// ServiceArgs overrides the arguments the root command parses, instead of
// the process's own os.Args[1:]. Primarily useful for tests, and for
// embedding a cli.Service where argument parsing shouldn't come from the
// process's real command line.
func ServiceArgs(args []string) ServiceOpt {
	return func(s *serviceOpts) {
		s.args = args
	}
}

func App(app ...func(*Service) error) ServiceOpt {
	return func(s *serviceOpts) {
		s.apps = append(s.apps, app...)
	}
}
