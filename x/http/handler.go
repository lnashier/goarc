package http

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
)

// JSONHandler is pre-assembled to write response with content-type "application/json".
type JSONHandler func(*http.Request) (any, error)

func (h JSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := h(r)
	if err != nil {
		ConvertError(err).WriteJSON(w)
		return
	}
	data, err := json.Marshal(result)
	if err != nil {
		ConvertError(fmt.Errorf("failed to marshal response: %w", err)).WriteJSON(w)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// TextHandler is pre-assembled to write response with content-type "plain/text".
type TextHandler func(*http.Request) (string, error)

func (h TextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := h(r)
	if err != nil {
		ConvertError(err).WriteText(w)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

// XMLHandler is pre-assembled to write response with content-type "application/xml".
type XMLHandler func(*http.Request) (any, error)

func (h XMLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := h(r)
	if err != nil {
		ConvertError(err).WriteText(w)
		return
	}
	data, err := xml.Marshal(result)
	if err != nil {
		ConvertError(fmt.Errorf("failed to marshal response: %w", err)).WriteText(w)
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
