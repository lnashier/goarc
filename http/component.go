package http

// Component defines long-running services within the http service
// that require graceful shutdown when service shuts down.
type Component interface {
	Stop()
}
