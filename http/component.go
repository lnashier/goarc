package http

// Component defines log-running services within the http service
// that require graceful shutdown when service shuts down
type Component interface {
	Stop()
}
