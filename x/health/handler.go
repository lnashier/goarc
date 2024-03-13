package health

import (
	"net/http"
)

type Controller struct {
	done chan struct{}
}

func New() *Controller {
	return &Controller{done: make(chan struct{})}
}

// Stop sets ready-handler status to NotFound.
// live-handler status remains OK until service completely goes away.
func (hc *Controller) Stop() {
	close(hc.done)
}

// LiveHandler handles the HTTP request for the live endpoint,
// indicating that the service is up and running.
// This should not be used for load-balancer purposes.
func (hc *Controller) LiveHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

// ReadyHandler handles the HTTP request for the ready endpoint,
// indicating whether the service is ready to accept new connections.
// This should be used for load-balancer purposes.
func (hc *Controller) ReadyHandler(w http.ResponseWriter, _ *http.Request) {
	select {
	case <-hc.done:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
	default:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	}
}
