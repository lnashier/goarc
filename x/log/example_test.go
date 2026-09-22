package log_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	stdlog "log"
	"log/slog"

	"github.com/lnashier/goarc/v2/x/log"
)

func Example() {
	var buf bytes.Buffer
	l := &log.Logger{
		ServiceName: "example",
		Logger:      stdlog.New(&buf, "", 0),
	}

	l.Info("hello %s", "world")

	var e log.Entry
	_ = json.Unmarshal(buf.Bytes(), &e)
	fmt.Println(e.Service, e.Level, e.Message)
	// Output:
	// example INFO hello world
}

// ExampleLogger_slogInterop shows Logger backing a slog.Logger, so
// attribute-based structured logging through the standard log/slog API
// still produces goarc's Entry shape, with attrs nested under Entry.Attrs.
func ExampleLogger_slogInterop() {
	var buf bytes.Buffer
	l := &log.Logger{
		ServiceName: "example",
		Logger:      stdlog.New(&buf, "", 0),
	}

	slog.New(l).With("request_id", "abc123").Info("handled request")

	var e log.Entry
	_ = json.Unmarshal(buf.Bytes(), &e)
	fmt.Println(e.Message, e.Attrs["request_id"])
	// Output:
	// handled request abc123
}
