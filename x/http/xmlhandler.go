package http

import (
	"encoding/xml"
	"net/http"
)

// XMLHandler wraps original Route function to write response in XML Format
type XMLHandler struct {
	Route Route
}

func (x *XMLHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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
