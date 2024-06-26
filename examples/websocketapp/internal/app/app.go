package app

import (
	"github.com/gorilla/websocket"
	goarchttp "github.com/lnashier/goarc/http"
	xhttp "github.com/lnashier/goarc/x/http"
	"net/http"
	"time"
	"websocketapp/internal/app/echo"
)

func App(srv *goarchttp.Service) error {
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
				xhttp.NewError(http.StatusNotAcceptable, "", nil).WriteJSON(w)
				return
			}

			conn, err := upgrader.Upgrade(w, req, nil)
			if err != nil {
				xhttp.NewError(http.StatusBadRequest, err.Error(), err).WriteJSON(w)
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
				xhttp.NewError(http.StatusInternalServerError, err.Error(), err).WriteJSON(w)
				return
			}
		},
	))

	return nil
}
