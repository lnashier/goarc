package env

import (
	"os"
	"sync"
)

const (
	Local = "local"
	Dev   = "dev"
	Prod  = "prod"
)

type Environment string

func (e Environment) IsLocal() bool {
	return e == Local
}

func (e Environment) IsDev() bool {
	return e == Dev
}

func (e Environment) IsProd() bool {
	return e == Prod
}

func (e Environment) String() string {
	return string(e)
}

func parseEnv() Environment {
	v, ok := os.LookupEnv("ENV")
	if !ok {
		v = Local
	}
	return Environment(v)
}

var getOnce = sync.OnceValue(parseEnv)

// Get returns the process's Environment, read from the ENV environment
// variable once and cached for the lifetime of the process; it defaults to
// Local if ENV is unset.
func Get() Environment {
	return getOnce()
}

// Hostname returns "localhost" when Get().IsLocal(), otherwise the
// machine's hostname (or the lookup error's message, if it fails).
func Hostname() string {
	if Get().IsLocal() {
		return "localhost"
	}
	hostname, err := os.Hostname()
	if err != nil {
		return err.Error()
	}
	return hostname
}
