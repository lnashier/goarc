package handler

import (
	"github.com/lnashier/goarc/zson"
	"net/http"
)

// JSON wraps original Route function to write response in JSON Format
type JSON struct {
	Route Route
}

func (h *JSON) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	result, err := h.Route(req)
	if err != nil {
		HandleError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(zson.Marshal(result))
}
