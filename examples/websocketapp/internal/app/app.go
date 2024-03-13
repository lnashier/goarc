package app

import (
	"github.com/gorilla/websocket"
	shttp "github.com/lnashier/goarc/http"
	xhttp "github.com/lnashier/goarc/x/http"
	"net/http"
	"time"
	"websocketapp/internal/app/echo"
)

func App(srv *shttp.Service) error {
	upgrader := &websocket.Upgrader{
		HandshakeTimeout:  time.Duration(1000) * time.Millisecond,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		Subprotocols:      []string{},
		CheckOrigin:       func(r *http.Request) bool { return true },
		EnableCompression: false,
	}

	srv.Register("/echo", http.MethodGet, http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			if ok := websocket.IsWebSocketUpgrade(req); !ok {
				xhttp.HandleError(w, xhttp.NewError(http.StatusNotAcceptable, "", nil))
				return
			}

			conn, err := upgrader.Upgrade(w, req, nil)
			if err != nil {
				xhttp.HandleError(w, xhttp.NewError(http.StatusBadRequest, err.Error(), err))
				return
			}
			defer conn.Close()

			echoer := &echo.Echoer{
				Conn:             conn,
				Msgs:             make(chan *echo.Message),
				ConnClosed:       make(chan struct{}),
				ServiceGoingAway: make(chan struct{}),
			}
			srv.Component(echoer)

			if err = echoer.Run(); err != nil {
				xhttp.HandleError(w, xhttp.NewError(http.StatusInternalServerError, err.Error(), err))
				return
			}
		},
	))

	return nil
}
