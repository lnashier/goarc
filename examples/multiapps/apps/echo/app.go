package echo

import (
	"errors"
	goarchttp "github.com/lnashier/goarc/http"
	xhttp "github.com/lnashier/goarc/x/http"
	"net/http"
)

func App(srv *goarchttp.Service) error {
	srv.Register("/echo", http.MethodPost, xhttp.JSONHandler(func(req *http.Request) (any, error) {
		echoReq := &Request{}
		err := xhttp.RequestParse(req, echoReq)
		if err != nil {
			return nil, xhttp.BadRequestf(err, err.Error())
		}

		return &Response{
			Message: echoReq.Message,
		}, nil
	}))

	return nil
}

type Request struct {
	Message string `json:"message"`
}

func (er Request) Validate(*http.Request) error {
	if len(er.Message) < 1 {
		return errors.New("missing message")
	}
	return nil
}

type Response struct {
	Message string `json:"message"`
}
