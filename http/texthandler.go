package http

import (
	"net/http"
)

// TextHandler wraps original Route function to write response in plain/text
type TextHandler struct {
	Route Route
}

func (h *TextHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	result, err := h.Route(req)
	if err != nil {
		HandleError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result.(string)))
}
