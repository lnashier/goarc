package app

import (
	"github.com/gorilla/websocket"
	"github.com/lnashier/go-app/config"
	chttp "github.com/lnashier/go-app/http"
	chandler "github.com/lnashier/go-app/http/handler"
	"github.com/lnashier/go-app/http/service"
	"net/http"
	"time"
	"websocketapp/internal/app/echo"
)

func App(cfg *config.Config, srv *service.Server) error {
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
				chandler.HandleError(w, chttp.NewError(http.StatusNotAcceptable, "", nil))
				return
			}

			conn, err := upgrader.Upgrade(w, req, nil)
			if err != nil {
				chandler.HandleError(w, chttp.NewError(http.StatusBadRequest, err.Error(), err))
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
				chandler.HandleError(w, chttp.NewError(http.StatusInternalServerError, err.Error(), err))
				return
			}
		},
	))

	return nil
}
