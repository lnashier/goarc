package http

import (
	"net/http"
)

// JSONHandler wraps original Route function to write response in JSON Format
type JSONHandler struct {
	Route Route
}

func (h *JSONHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	result, err := h.Route(req)
	if err != nil {
		HandleError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(marshal(result))
}
