package service

import (
	"net/http"
)

type reject struct {
	h    http.Handler
	done chan struct{}
}

func (rj *reject) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	case <-rj.done:
		// header: Retry-After
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte{})
	default:
		rj.h.ServeHTTP(w, r)
	}
}
