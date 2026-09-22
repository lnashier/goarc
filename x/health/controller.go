package health

import (
	"context"
	"sync"

	xhttp "github.com/lnashier/goarc/v2/x/http"
)

// Controller implements goarc.Component.
type Controller struct {
	done     chan struct{}
	stopOnce sync.Once
}

func New() *Controller {
	return &Controller{done: make(chan struct{})}
}

// Stop sets ready-handler status to NotFound.
// live-handler status remains OK until service completely goes away.
//
// Stop is safe to call more than once.
func (hc *Controller) Stop(context.Context) error {
	hc.stopOnce.Do(func() {
		close(hc.done)
	})
	return nil
}

// Live handles the HTTP request for the live endpoint,
// indicating that the service is up and running.
// This should not be used for load-balancer purposes.
func (hc *Controller) Live() string {
	return "up"
}

// Ready handles the HTTP request for the ready endpoint,
// indicating whether the service is ready to accept new connections.
// This should be used for load-balancer purposes.
func (hc *Controller) Ready() (string, error) {
	select {
	case <-hc.done:
		return "", xhttp.NotFoundf(nil, "service not found")
	default:
		return "up", nil
	}
}
