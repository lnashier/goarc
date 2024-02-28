package service

import (
	"net/http"
)

var (
	healthStatusUp   = []byte(`{"status": "up"}`)
	healthStatusDown = []byte(`{"status": "down"}`)
)

type healthController struct {
	done chan struct{}
}

func newHealthController() *healthController {
	return &healthController{done: make(chan struct{})}
}

func (hc *healthController) setStatusDown() {
	close(hc.done)
}

func (hc *healthController) LiveHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(healthStatusUp)
}

func (hc *healthController) ReadyHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	select {
	case <-hc.done:
		w.WriteHeader(http.StatusNotFound)
		w.Write(healthStatusDown)
	default:
		w.WriteHeader(http.StatusOK)
		w.Write(healthStatusUp)
	}
}
