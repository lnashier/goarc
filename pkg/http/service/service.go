package service

// Service defines any service running within the app
type Service interface {
	// Stop is to shut down the running service
	Stop() error
}
