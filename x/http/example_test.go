package http_test

import (
	"context"
	"fmt"
	nethttp "net/http"
	"net/http/httptest"

	xhttp "github.com/lnashier/goarc/v2/x/http"
)

func Example() {
	srv := httptest.NewServer(xhttp.JSONHandler(func(*nethttp.Request) (any, error) {
		return map[string]string{"hello": "world"}, nil
	}))
	defer srv.Close()

	c := xhttp.NewClient(xhttp.WithHost(srv.URL))
	var result map[string]string
	resp, err := c.GetDecoded(context.Background(), "/", nethttp.Header{}, &result, nil)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	fmt.Println(result["hello"])
	// Output:
	// world
}
