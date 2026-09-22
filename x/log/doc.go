// Package log provides a structured, JSON-line Logger with a small
// printf-style convenience API (Debug, Info, Error, Panic, Net) and, since
// Logger also implements slog.Handler, full interop with the standard
// library's log/slog package for attribute-based structured logging.
//
// The zero value is ready to use.
package log
