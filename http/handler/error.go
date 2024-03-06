package handler

import (
	chttp "github.com/lnashier/go-app/http"
	"github.com/lnashier/go-app/zson"
	"net/http"
)

func HandleError(w http.ResponseWriter, err error) {
	var httpErr *chttp.Error
	switch specificError := err.(type) {
	case *chttp.Error:
		httpErr = specificError
	default:
		// default for unknown errors is 500 with no client-facing message
		httpErr = chttp.NewError(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(httpErr.Status)
	w.Write(zson.Marshal(httpErr))
}
