package http

import "net/http"

// Route is a function that handles an HTTP request
type Route func(*http.Request) (any, error)
