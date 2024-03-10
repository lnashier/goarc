package cli

import (
	"github.com/lnashier/goarc/config"
)

type Opt func(*opts)

type opts struct {
	cfg *config.Config
	//logger *log.Logger TODO
	apps []func(*config.Config, *Service) error
}

func (s *opts) applyOptions(opts []Opt) {
	for _, o := range opts {
		o(s)
	}
}

func defaultOpts() *opts {
	return &opts{
		apps: []func(*config.Config, *Service) error{},
	}
}

func WithConfig(cfg *config.Config) Opt {
	return func(s *opts) {
		s.cfg = cfg
	}
}

/*TODO
// WithLogger
func WithLogger(logger *log.Logger) Opt {
	return func(s *opts) {
		s.logger = logger
	}
}
*/

func WithApp(app ...func(*config.Config, *Service) error) Opt {
	return func(s *opts) {
		s.apps = append(s.apps, app...)
	}
}
