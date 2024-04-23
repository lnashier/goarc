package goarc

// Service represents a generic service that goarc framework can Start and Stop.
type Service interface {
	// Start should initiate the startup of the service.
	// It returns an error if the startup process encounters any issues.
	Start() error

	// Stop should initiate the shutdown of the service.
	// It returns an error if the shutdown process encounters any issues.
	Stop() error
}

// ServiceFunc serves as an adapter, enabling the utilization of regular functions as goarc services.
// If f is a function with the correct signature, ServiceFunc(f) creates a Service that invokes f.
// The function f receives a boolean parameter indicating whether it's invoked as Start (true) or Stop (false).
type ServiceFunc func(bool) error

// Start calls f(true).
func (s ServiceFunc) Start() error {
	return s(true)
}

// Stop calls f(false).
func (s ServiceFunc) Stop() error {
	return s(false)
}
