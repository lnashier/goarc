package http

import (
	"encoding/json"
	"net/http"
)

type ClientOpt func(*clientOpts)

var defaultClientOpts = clientOpts{
	encoder: json.Marshal,
	decoder: json.Unmarshal,
}

type clientOpts struct {
	host    string
	timeout int64
	tr      *http.Transport
	encoder func(any) ([]byte, error)
	decoder func([]byte, any) error
}

func (s *clientOpts) apply(opts []ClientOpt) {
	for _, o := range opts {
		o(s)
	}
}

func WithHost(host string) ClientOpt {
	return func(s *clientOpts) {
		s.host = host
	}
}

func WithTimeout(t int64) ClientOpt {
	return func(s *clientOpts) {
		s.timeout = t
	}
}

func WithTransport(tr *http.Transport) ClientOpt {
	return func(s *clientOpts) {
		s.tr = tr
	}
}

func WithEncoder(e func(any) ([]byte, error)) ClientOpt {
	return func(s *clientOpts) {
		s.encoder = e
	}
}

func WithDecoder(d func([]byte, any) error) ClientOpt {
	return func(s *clientOpts) {
		s.decoder = d
	}
}
