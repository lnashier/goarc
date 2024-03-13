package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	host    string
	hc      *http.Client
	encoder func(any) ([]byte, error)
	decoder func([]byte, any) error
}

func NewClient(opt ...ClientOpt) *Client {
	opts := defaultClientOpts
	opts.apply(opt)

	hc := &http.Client{
		Timeout: time.Duration(opts.timeout) * time.Second,
	}
	if opts.tr != nil {
		hc.Transport = opts.tr
	}

	return &Client{
		hc:      hc,
		host:    opts.host,
		encoder: opts.encoder,
		decoder: opts.decoder,
	}
}

type Request struct {
	*http.Request
	body io.ReadSeeker
}

func (c *Client) NewRequest(ctx context.Context, method, path string, header http.Header, body io.ReadSeeker) (*Request, error) {
	var rc io.ReadCloser
	var bodyLen int64
	if body != nil {
		// some servers are strict on Content-Length header
		bodyLen, _ = io.Copy(io.Discard, body)
		rc = io.NopCloser(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.host, path), rc)
	if err != nil {
		return nil, err
	}
	req.Header = header
	req.ContentLength = bodyLen
	return &Request{req, body}, nil
}

func (c *Client) NewRequestEncoded(ctx context.Context, method, path string, header http.Header, body any) (*Request, error) {
	encodedBody, err := c.encoder(body)
	if err != nil {
		return nil, err
	}
	return c.NewRequest(ctx, method, path, header, bytes.NewReader(encodedBody))
}

func (c *Client) Do(req *Request, retry *Retry) (*http.Response, error) {
	if retry == nil {
		retry = noRetry()
	}
	remain := retry.Max + 1
	var retryErr error
	for {
		// Always rewind request body when non-nil.
		if req.body != nil {
			if _, err := req.body.Seek(0, 0); err != nil {
				return nil, fmt.Errorf("failed to seek req body: %v", err)
			}
		}

		// Attempt the request
		resp, err := c.hc.Do(req.Request)

		// Check if request should be retried
		var retryOK bool
		retryOK, retryErr = retry.Policy(resp, err)

		// Decide if retrying or not
		if !retryOK {
			if retryErr != nil {
				err = retryErr
			}
			return resp, err
		}

		// Before retrying consume any response to reuse the connection
		if err == nil {
			drainBody(resp.Body)
		}

		remain--
		if remain < 1 {
			break
		}

		wait := retry.Backoff(retry.WaitMin, retry.WaitMax, retry.Max-remain, resp)
		retry.OnTry(wait, retry.Max-remain, retryErr)
		time.Sleep(wait)
	}
	retry.OnTry(0, retry.Max-remain, retryErr)
	return nil, fmt.Errorf("%s %s giving up after %d retries", req.Method, req.URL, retry.Max)
}

func (c *Client) DoDecoded(req *Request, result any, retry *Retry) (*http.Response, error) {
	resp, err := c.Do(req, retry)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, fmt.Errorf("could not read response body: %v", err)
	}

	resp.Body = io.NopCloser(bytes.NewReader(respBody))
	if len(respBody) == 0 {
		return resp, nil
	}

	// Let's first work on successful response
	if resp.StatusCode == http.StatusOK {
		// decode to given result only if successful
		err = c.decoder(respBody, result)
		if err != nil {
			return resp, fmt.Errorf("failed to decode response-body: %v", err)
		}
		return resp, nil
	}

	return resp, NewError(resp.StatusCode, string(respBody), nil)
}

func (c *Client) Get(ctx context.Context, path string, header http.Header, retry *Retry) (*http.Response, error) {
	req, err := c.NewRequest(ctx, http.MethodGet, path, header, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req, retry)
}

// GetDecoded decodes response-body
func (c *Client) GetDecoded(ctx context.Context, path string, header http.Header, result any, retry *Retry) (*http.Response, error) {
	req, err := c.NewRequest(ctx, http.MethodGet, path, header, nil)
	if err != nil {
		return nil, err
	}
	return c.DoDecoded(req, result, retry)
}

func (c *Client) Post(ctx context.Context, path string, header http.Header, body io.ReadSeeker, retry *Retry) (*http.Response, error) {
	req, err := c.NewRequest(ctx, http.MethodPost, path, header, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req, retry)
}

// PostEncodedDecoded encodes request-body and decodes response-body
func (c *Client) PostEncodedDecoded(ctx context.Context, path string, header http.Header, body, result any, retry *Retry) (*http.Response, error) {
	req, err := c.NewRequestEncoded(ctx, http.MethodPost, path, header, body)
	if err != nil {
		return nil, err
	}
	return c.DoDecoded(req, result, retry)
}

func (c *Client) Patch(ctx context.Context, path string, header http.Header, body io.ReadSeeker, retry *Retry) (*http.Response, error) {
	req, err := c.NewRequest(ctx, http.MethodPatch, path, header, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req, retry)
}

// PatchEncodedDecoded encodes request-body and decodes response-body
func (c *Client) PatchEncodedDecoded(ctx context.Context, path string, header http.Header, body, result any, retry *Retry) (*http.Response, error) {
	req, err := c.NewRequestEncoded(ctx, http.MethodPatch, path, header, body)
	if err != nil {
		return nil, err
	}
	return c.DoDecoded(req, result, retry)
}

func drainBody(b io.ReadCloser) {
	defer b.Close()
	io.Copy(io.Discard, b)
}
