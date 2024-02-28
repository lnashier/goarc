package service

import (
	"bufio"
	"github.com/lnashier/go-app/pkg/log"
	"net"
	"net/http"
	"time"
)

type LogHandler struct {
	logger  *log.Logger
	handler http.Handler
}

func (h *LogHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t := time.Now()
	rTracker := makeResponseTracker(w)
	url := *req.URL
	h.handler.ServeHTTP(rTracker, req)
	h.logger.Net(&NetLogMessage{
		Site:          url.Hostname(),
		Src:           req.Header.Get("Origin"),
		Method:        req.Method,
		Header:        req.Header,
		URL:           url.EscapedPath(),
		Query:         url.RawQuery,
		Status:        rTracker.Status(),
		ResponseBytes: rTracker.Size(),
		Msec:          time.Since(t).Nanoseconds() / 1e6,
	})
}

func makeResponseTracker(w http.ResponseWriter) trackingResponseWriter {
	var tracker trackingResponseWriter

	if _, ok := w.(http.Hijacker); ok {
		tracker = &hijackTracker{
			responseTracker{
				w:      w,
				status: http.StatusOK,
			}}
	} else {
		tracker = &responseTracker{
			w:      w,
			status: http.StatusOK,
		}
	}

	h, ok1 := tracker.(http.Hijacker)
	c, ok2 := w.(http.CloseNotifier)

	if ok1 && ok2 {
		return hijackCloseNotifier{tracker, h, c}
	}

	if ok2 {
		return &closeNotifyWriter{tracker, c}
	}

	return tracker
}

type trackingResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	Status() int
	Size() int
}

// responseTracker is wrapper of http.ResponseWriter that keeps track of its HTTP
// status code and body size
type responseTracker struct {
	w      http.ResponseWriter
	status int
	size   int
}

func (t *responseTracker) Header() http.Header {
	return t.w.Header()
}

func (t *responseTracker) Write(b []byte) (int, error) {
	size, err := t.w.Write(b)
	t.size += size
	return size, err
}

func (t *responseTracker) WriteHeader(s int) {
	t.w.WriteHeader(s)
	t.status = s
}

func (t *responseTracker) Status() int {
	return t.status
}

func (t *responseTracker) Size() int {
	return t.size
}

func (t *responseTracker) Flush() {
	f, ok := t.w.(http.Flusher)
	if ok {
		f.Flush()
	}
}

type hijackTracker struct {
	responseTracker
}

func (t *hijackTracker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h := t.responseTracker.w.(http.Hijacker)
	conn, rw, err := h.Hijack()
	if err == nil && t.responseTracker.status == 0 {
		// The status will be StatusSwitchingProtocols if there was no error and
		// WriteHeader has not been called yet
		t.responseTracker.status = http.StatusSwitchingProtocols
	}
	return conn, rw, err
}

type closeNotifyWriter struct {
	trackingResponseWriter
	http.CloseNotifier
}

type hijackCloseNotifier struct {
	trackingResponseWriter
	http.Hijacker
	http.CloseNotifier
}

type NetLogMessage struct {
	Site          string      `json:"site"`
	Src           string      `json:"src"`
	Method        string      `json:"method"`
	URL           string      `json:"url"`
	Query         string      `json:"query"`
	Header        http.Header `json:"header"`
	Status        int         `json:"status"`
	ResponseBytes int         `json:"responseBytes"`
	Msec          int64       `json:"msec"`
}
