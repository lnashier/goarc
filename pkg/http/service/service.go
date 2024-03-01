package service

// Component defines any component running within the app
type Component interface {
	// Stop is to shut down the running component
	Stop() error
}
