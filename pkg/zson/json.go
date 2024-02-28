package zson

import "encoding/json"

// Marshal wraps original json.Marshal to hide error
func Marshal(x any) []byte {
	m, _ := json.Marshal(x)
	return m
}

// MarshalIndent wraps original json.MarshalIndent to hide error
func MarshalIndent(x any, prefix, indent string) []byte {
	m, _ := json.MarshalIndent(x, prefix, indent)
	return m
}

// Unmarshal wraps original json.Unmarshal to hide error
func Unmarshal(data []byte, v any) any {
	_ = json.Unmarshal(data, v)
	return v
}
