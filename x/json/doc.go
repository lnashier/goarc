// Package json wraps encoding/json's Marshal, MarshalIndent and Unmarshal
// to hide their errors, for call sites (writing a log line, an HTTP error
// body) where there's no useful way to act on a marshal/unmarshal failure
// and a best-effort result is preferable to threading an error through.
package json
