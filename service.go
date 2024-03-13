package goarc

type Service interface {
	// Start should initiate the startup of the service.
	// It returns an error if the startup process encounters any issues.
	Start() error

	// Stop should initiate the shutdown of the service.
	// It returns an error if the shutdown process encounters any issues.
	Stop() error
}
