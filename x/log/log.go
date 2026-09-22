package log

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/lnashier/goarc/v2/x/buildinfo"
	"github.com/lnashier/goarc/v2/x/env"
	xjson "github.com/lnashier/goarc/v2/x/json"
)

type Type string

type Level string

const (
	// DebugLevel ...
	// Don't use it unless you're really debugging.
	// You don't want to see log.Debug in your code once debugged.
	DebugLevel Level = "DEBUG"
	// InfoLevel ...
	// Use it a lot and share information about your function status.
	InfoLevel Level = "INFO"
	// ErrorLevel ...
	// Use it whenever you do:
	// if err != nil {
	//		return err
	// }
	ErrorLevel Level = "ERROR"
	// PanicLevel is always enabled.
	// Don't use it unless you're in boot phase of the system.
	// Use ErrorLevel and let error bubble up.
	PanicLevel Level = "PANIC"

	// AppType is for application logs
	AppType Type = "app"
	// NetType is for ingress traffic logs
	NetType Type = "net"
)

// Entry is the JSON shape written for every log line.
type Entry struct {
	Hostname  string         `json:"hostname"`
	Service   string         `json:"service"`
	Hash      string         `json:"hash"`
	Timestamp string         `json:"timestamp,omitempty"`
	LogType   string         `json:"logType"`
	Level     string         `json:"level,omitempty"`
	Message   any            `json:"message"`
	Attrs     map[string]any `json:"attrs,omitempty"`
}

// Logger writes JSON-formatted Entry lines and, as a slog.Handler, plugs
// directly into the standard library's log/slog package — so it can back a
// slog.Logger (slog.New(logger)), giving structured, attribute-based
// logging through the standard slog API while still producing goarc's
// Entry shape.
//
// The zero value is ready to use: it writes to os.Stdout, enables every
// level, and publishes nowhere. Logger can be used simultaneously from
// multiple goroutines.
type Logger struct {
	ServiceName string
	Hostname    string
	Hash        string
	// Logger is the underlying output. If nil, a default writing to
	// os.Stdout is used.
	Logger *log.Logger
	// Verifier reports whether a Level is enabled. If nil, every level is
	// enabled. PanicLevel and Net entries are always logged regardless of
	// Verifier.
	Verifier func(Level) bool
	// Publisher, if set, is called with every Entry after it is written,
	// e.g. to forward it to an external logging service.
	Publisher func(*Entry)

	segments []attrSegment // accumulated via WithAttrs/WithGroup
}

var _ slog.Handler = (*Logger)(nil)

// DefaultLogger is the default Logger, used by the package-level Net,
// Debug, Info, Error and Panic functions. It is the zero value: ready to
// use as-is.
var DefaultLogger = &Logger{}

func Net(msg any) {
	DefaultLogger.Net(msg)
}

// Debug ... don't use it unless you're really debugging
func Debug(format string, v ...any) {
	DefaultLogger.Debug(format, v...)
}

func Info(format string, v ...any) {
	DefaultLogger.Info(format, v...)
}

func Error(format string, v ...any) {
	DefaultLogger.Error(format, v...)
}

// Panic is always enabled
func Panic(format string, v ...any) {
	DefaultLogger.Panic(format, v...)
}

func (l *Logger) Net(msg any) {
	l.log("", NetType, msg)
}

func (l *Logger) Debug(f string, v ...any) {
	l.log(DebugLevel, AppType, fmt.Sprintf(f, v...))
}

func (l *Logger) Info(f string, v ...any) {
	l.log(InfoLevel, AppType, fmt.Sprintf(f, v...))
}

func (l *Logger) Error(f string, v ...any) {
	l.log(ErrorLevel, AppType, fmt.Sprintf(f, v...))
}

// Panic is always enabled. Unlike Debug/Info/Error, the panic value carries
// the formatted message, so a recover() handler can report it.
func (l *Logger) Panic(f string, v ...any) {
	msg := fmt.Sprintf(f, v...)
	l.emit(PanicLevel, AppType, msg, time.Now(), nil)
	panic(msg)
}

func (l *Logger) log(level Level, logType Type, msg any) {
	if logType == NetType || level == PanicLevel || l.verifier()(level) {
		l.emit(level, logType, msg, time.Now(), nil)
	}
}

func (l *Logger) emit(level Level, logType Type, msg any, t time.Time, attrs map[string]any) {
	e := &Entry{
		Hostname: l.hostname(),
		Service:  l.serviceName(),
		Hash:     l.hash(),
		LogType:  string(logType),
		Level:    string(level),
		Message:  msg,
		Attrs:    attrs,
	}
	// A zero Record.Time (as slogtest exercises) means "no timestamp was
	// supplied"; omit it rather than formatting the zero time.
	if !t.IsZero() {
		e.Timestamp = t.Format(time.RFC3339)
	}
	l.output().Print(string(xjson.Marshal(e)))
	l.publisher()(e)
}

// Enabled implements slog.Handler.
func (l *Logger) Enabled(_ context.Context, level slog.Level) bool {
	return l.verifier()(levelFor(level))
}

// Handle implements slog.Handler. It writes r as an Entry with LogType
// AppType, and Attrs built from every attribute bound via WithAttrs and
// WithGroup, nested under any open groups, plus r's own attributes.
func (l *Logger) Handle(_ context.Context, r slog.Record) error {
	lvl := levelFor(r.Level)
	if !l.verifier()(lvl) {
		return nil
	}

	var recordAttrs []slog.Attr
	r.Attrs(func(a slog.Attr) bool {
		recordAttrs = append(recordAttrs, a)
		return true
	})

	attrs := buildAttrs(l.segments, recordAttrs)
	l.emit(lvl, AppType, r.Message, r.Time, attrs)
	return nil
}

// WithAttrs implements slog.Handler.
func (l *Logger) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return l
	}
	nl := *l
	nl.segments = append(append([]attrSegment{}, l.segments...), attrSegment{attrs: attrs})
	return &nl
}

// WithGroup implements slog.Handler.
func (l *Logger) WithGroup(name string) slog.Handler {
	if name == "" {
		return l
	}
	nl := *l
	nl.segments = append(append([]attrSegment{}, l.segments...), attrSegment{group: name})
	return &nl
}

func (l *Logger) output() *log.Logger {
	if l.Logger != nil {
		return l.Logger
	}
	return defaultOutput
}

func (l *Logger) verifier() func(Level) bool {
	if l.Verifier != nil {
		return l.Verifier
	}
	return func(Level) bool { return true }
}

func (l *Logger) publisher() func(*Entry) {
	if l.Publisher != nil {
		return l.Publisher
	}
	return func(*Entry) {}
}

func (l *Logger) serviceName() string {
	if l.ServiceName != "" {
		return l.ServiceName
	}
	return "unknown"
}

func (l *Logger) hostname() string {
	if l.Hostname != "" {
		return l.Hostname
	}
	return env.Hostname()
}

func (l *Logger) hash() string {
	if l.Hash != "" {
		return shortHash(l.Hash)
	}
	return shortHash(buildinfo.Hash)
}

var defaultOutput = log.New(os.Stdout, "", 0)

func shortHash(h string) string {
	return h[:len(h)/2]
}

// levelFor maps a slog.Level onto goarc's three Handler-facing severities.
// slog.LevelWarn falls into InfoLevel, since goarc has no dedicated "warn"
// tier in its Handler-facing model.
func levelFor(l slog.Level) Level {
	switch {
	case l >= slog.LevelError:
		return ErrorLevel
	case l >= slog.LevelInfo:
		return InfoLevel
	default:
		return DebugLevel
	}
}
