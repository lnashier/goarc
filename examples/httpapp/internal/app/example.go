package app

import (
	"crypto/md5"
	"errors"
	"github.com/gorilla/mux"
	xhttp "github.com/lnashier/goarc/x/http"
	"net/http"
)

func (c Controller) SaveExample(req *http.Request) (any, error) {
	exampleReq := &ExampleRequest{}
	err := xhttp.RequestParse(req, exampleReq)
	if err != nil {
		return xhttp.BadRequestf(err, err.Error())
	}

	msgID := md5.Sum([]byte(exampleReq.Data))

	return ExampleResponse{
		MsgID: string(msgID[:]),
	}, nil

}

func (c Controller) GetExample(req *http.Request) (any, error) {
	msgID := mux.Vars(req)["id"]

	data, ok := c.store[msgID]
	if ok {
		return xhttp.NotFound(nil)
	}
	return data, nil
}

type ExampleRequest struct {
	Data string `json:"data"`
}

func (er ExampleRequest) Validate(*http.Request) error {
	if len(er.Data) < 1 {
		return errors.New("missing data")
	}
	return nil
}

type ExampleResponse struct {
	MsgID string `json:"msgId"`
}
