package env

import (
	"os"
)

const (
	Local = "local"
	Dev   = "dev"
	Prod  = "prod"
)

type Environment string

var environment Environment

func Get() Environment {
	if environment == "" {
		environment = getFromEnvironment()
	}
	return environment
}

func getFromEnvironment() Environment {
	env, ok := os.LookupEnv("ENV")
	if !ok {
		env = Local
	}
	return Environment(env)
}

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

func Hostname() string {
	if Get().IsLocal() {
		return "localhost"
	}
	var hostname string
	var err error
	if hostname, err = os.Hostname(); err != nil {
		hostname = err.Error()
	}
	return hostname
}
