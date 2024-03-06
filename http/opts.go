package http

import (
	"encoding/json"
	"net/http"
)

type Opt func(*opts)

type opts struct {
	host    string
	timeout int64
	tr      *http.Transport
	encoder func(any) ([]byte, error)
	decoder func([]byte, any) error
}

func (s *opts) applyOptions(opts []Opt) {
	for _, o := range opts {
		o(s)
	}
}

func defaultOpts() *opts {
	return &opts{
		encoder: json.Marshal,
		decoder: json.Unmarshal,
	}
}

func WithHost(host string) Opt {
	return func(s *opts) {
		s.host = host
	}
}

func WithTimeout(t int64) Opt {
	return func(s *opts) {
		s.timeout = t
	}
}

func WithTransport(tr *http.Transport) Opt {
	return func(s *opts) {
		s.tr = tr
	}
}

func WithEncoder(e func(any) ([]byte, error)) Opt {
	return func(s *opts) {
		s.encoder = e
	}
}

func WithDecoder(d func([]byte, any) error) Opt {
	return func(s *opts) {
		s.decoder = d
	}
}
