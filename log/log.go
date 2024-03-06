package log

import (
	"fmt"
	"github.com/lnashier/go-app/buildinfo"
	"github.com/lnashier/go-app/env"
	"github.com/lnashier/go-app/zson"
	"log"
	"os"
	"time"
)

type Type string

type Level string

const (
	// DebugLevel
	// Don't use it unless you're really debugging.
	// You don't want to see log.Debug in your code once debugged.
	DebugLevel Level = "DEBUG"
	// InfoLevel
	// Use it a lot and share information about your function status.
	// Every good thing begins with a positive intention and blossoms through thoughtful actions and perseverance.
	InfoLevel Level = "INFO"
	// ErrorLevel
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

// Logger is a logging object that generates JSON formatted log Entry to provided log.Logger.
// Each logging operation makes a single call to the log.Logger's Print method.
// Logger can be used simultaneously from multiple goroutines.
// Its zero value (DefaultLogger) is a usable object that uses os.Stdout to output log lines.
type Logger struct {
	AppName  string
	Hostname string
	Hash     string
	Logger   *log.Logger
	// Verifier verifies if provided logging Level is enabled
	// PanicLevel is always enabled
	Verifier func(Level) bool
	// Publisher allows sending log events to external logging services
	Publisher func(*Entry)
}

// DefaultLogger is the default Logger and is used by Debug, Info, Error and Panic.
var DefaultLogger = func() *Logger {
	return &Logger{
		Hostname: env.Hostname(),
		AppName:  "unknown",
		Hash:     buildinfo.Hash[:len(buildinfo.Hash)/2],
		Logger:   log.New(os.Stdout, "", 0),
		Verifier: func(a Level) bool {
			return true
		},
		Publisher: func(*Entry) {},
	}
}()

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

// Panic ...
// Panic logging level is always enabled
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

// Panic ...
// Panic logging level is always enabled
func (l *Logger) Panic(f string, v ...any) {
	l.log(PanicLevel, AppType, fmt.Sprintf(f, v...))
	panic("")
}

func (l *Logger) log(level Level, logType Type, msg any) {
	if logType == NetType || level == PanicLevel || l.Verifier(level) {
		e := &Entry{
			Hostname:  l.Hostname,
			App:       l.AppName,
			Hash:      l.Hash,
			Timestamp: time.Now().Format(time.RFC3339),
			LogType:   string(logType),
			Level:     string(level),
			Message:   msg,
		}
		l.Logger.Print(string(zson.Marshal(e)))
		l.Publisher(e)
	}
}

type Entry struct {
	Hostname  string `json:"hostname"`
	App       string `json:"app"`
	Hash      string `json:"hash"`
	Timestamp string `json:"timestamp"`
	LogType   string `json:"logType"`
	Level     string `json:"level,omitempty"`
	Message   any    `json:"message"`
}
