package handler

import (
	"encoding/xml"
	"net/http"
)

// XML wraps original Route function to write response in XML Format
type XML struct {
	Route Route
}

func (x *XML) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	result, err := x.Route(req)
	if err != nil {
		HandleError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	mResult, _ := xml.Marshal(result)
	w.Write(mResult)
}
