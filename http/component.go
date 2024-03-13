package http

// Component defines any component running within the http server
// that require graceful shutdown when server shuts down
type Component interface {
	Stop()
}
