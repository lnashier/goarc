package http

import "encoding/json"

// marshal wraps original json.Marshal to hide error
func marshal(x any) []byte {
	m, _ := json.Marshal(x)
	return m
}
