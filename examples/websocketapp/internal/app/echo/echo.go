package echo

import (
	"github.com/gorilla/websocket"
	"github.com/lnashier/go-app/log"
	"math"
	"strconv"
	"sync"
	"time"
)

type Echoer struct {
	Conn             *websocket.Conn
	Msgs             chan *Message
	ConnClosed       chan struct{}
	ServiceGoingAway chan struct{}
}

func (e *Echoer) Stop() error {
	close(e.ServiceGoingAway)
	return nil
}

type Message struct {
	Type int
	Data []byte
}

func (e *Echoer) Run() error {
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go e.receive(wg)

	wg.Add(1)
	go e.send(wg)

	wg.Add(1)
	go e.pingPong(wg)

	wg.Wait()
	return nil
}

func (e *Echoer) receive(wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(e.ConnClosed)

	for {
		msgType, data, err := e.Conn.ReadMessage()
		switch {
		case err != nil:
			if !websocket.IsCloseError(err, 1000) {
				log.Error("received error %s", err.Error())
			}
			return
		default:
			e.Msgs <- &Message{
				Type: msgType,
				Data: data,
			}
		}
	}
}

func (e *Echoer) send(wg *sync.WaitGroup) {
	defer wg.Done()

	writeWait := time.Duration(10) * time.Second

	for {
		select {
		case <-e.ServiceGoingAway:
			e.Conn.WriteControl(websocket.CloseMessage, nil, time.Now().Add(writeWait))
			return
		case <-e.ConnClosed:
			return
		case msg, ok := <-e.Msgs:
			if !ok {
				return
			}
			if err := e.Conn.WriteMessage(msg.Type, msg.Data); err != nil {
				return
			}
		}
	}
}

func (e *Echoer) pingPong(wg *sync.WaitGroup) {
	defer wg.Done()

	readWait := time.Duration(10) * time.Second
	writeWait := time.Duration(10) * time.Second
	connMaxLife := time.Duration(60) * time.Second

	pingWait := time.Duration((readWait.Seconds()*9)/10) * time.Second
	pingTicker := time.NewTicker(pingWait)
	maxPings := math.MaxInt64
	if connMaxLife > 0 {
		maxPings = int(connMaxLife / pingWait)
	}
	pingCounter := 0

	e.Conn.SetPongHandler(func(pong string) error {
		_ = e.Conn.SetReadDeadline(time.Now().Add(readWait))
		return nil
	})

	e.Conn.SetPingHandler(func(ping string) error {
		_ = e.Conn.SetWriteDeadline(time.Now().Add(writeWait))
		return nil
	})

	for {
		select {
		case <-e.ServiceGoingAway:
			return
		case <-e.ConnClosed:
			return
		case <-pingTicker.C:
			pingCounter++
			_ = e.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := e.Conn.WriteControl(websocket.PingMessage, []byte(strconv.Itoa(pingCounter)), time.Now().Add(writeWait)); err != nil {
				return
			}
			if pingCounter > maxPings {
				e.Conn.WriteControl(websocket.CloseMessage, nil, time.Now().Add(writeWait))
				return
			}
		}
	}
}
