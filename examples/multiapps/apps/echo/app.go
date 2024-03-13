package echo

import (
	"errors"
	shttp "github.com/lnashier/goarc/http"
	xhttp "github.com/lnashier/goarc/x/http"
	"net/http"
)

func App(srv *shttp.Service) error {
	srv.Register("/echo", http.MethodPost, &xhttp.JSONHandler{Route: func(req *http.Request) (any, error) {
		echoReq := &EchoRequest{}
		err := xhttp.RequestParse(req, echoReq)
		if err != nil {
			return xhttp.BadRequestf(err, err.Error())
		}

		return EchoResponse{
			Message: echoReq.Message,
		}, nil

	}})

	return nil
}

type EchoRequest struct {
	Message string `json:"message"`
}

func (er EchoRequest) Validate(*http.Request) error {
	if len(er.Message) < 1 {
		return errors.New("missing message")
	}
	return nil
}

type EchoResponse struct {
	Message string `json:"message"`
}
