package log

import (
	"bytes"
	"encoding/json"
	"log"
	"log/slog"
	"strings"
	"testing"
	"testing/slogtest"
	"time"

	"github.com/lnashier/goarc/v2/x/buildinfo"
)

func TestZeroValue_IsUsable(t *testing.T) {
	var buf bytes.Buffer
	l := &Logger{}
	l.Logger = log.New(&buf, "", 0) // only override output, so the test doesn't hit real stdout

	l.Info("hello %s", "world")

	var e Entry
	if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
		t.Fatalf("output is not valid JSON: %v (%q)", err, buf.String())
	}
	if e.Message != "hello world" {
		t.Fatalf("Message = %q, want %q", e.Message, "hello world")
	}
	if e.Service != "unknown" {
		t.Fatalf("Service = %q, want %q", e.Service, "unknown")
	}
}

func TestVerifier_GatesDebugInfoError(t *testing.T) {
	var buf bytes.Buffer
	var seen []Level
	l := &Logger{
		Logger: log.New(&buf, "", 0),
		Verifier: func(lv Level) bool {
			seen = append(seen, lv)
			return lv == ErrorLevel
		},
	}

	l.Debug("d")
	l.Info("i")
	l.Error("e")

	lines := nonEmptyLines(buf.String())
	if len(lines) != 1 {
		t.Fatalf("got %d log lines, want 1 (only ErrorLevel should pass the Verifier): %v", len(lines), lines)
	}
	var e Entry
	if err := json.Unmarshal([]byte(lines[0]), &e); err != nil {
		t.Fatal(err)
	}
	if e.Level != string(ErrorLevel) {
		t.Fatalf("logged entry level = %q, want %q", e.Level, ErrorLevel)
	}
	if got := diffStr(seen, []Level{DebugLevel, InfoLevel, ErrorLevel}); got != "" {
		t.Fatal(got)
	}
}

func TestVerifier_DoesNotGatePanicOrNet(t *testing.T) {
	var buf bytes.Buffer
	l := &Logger{
		Logger:   log.New(&buf, "", 0),
		Verifier: func(Level) bool { return false },
	}

	l.Net("net event")

	func() {
		defer func() { _ = recover() }()
		l.Panic("boom %d", 1)
	}()

	lines := nonEmptyLines(buf.String())
	if len(lines) != 2 {
		t.Fatalf("got %d log lines, want 2 (Net and Panic bypass Verifier): %q", len(lines), buf.String())
	}
}

func TestPanic_CarriesMessageInPanicValue(t *testing.T) {
	l := &Logger{Logger: log.New(&bytes.Buffer{}, "", 0)}

	defer func() {
		r := recover()
		if r != "boom 42" {
			t.Fatalf("recovered value = %v, want %q", r, "boom 42")
		}
	}()
	l.Panic("boom %d", 42)
}

func TestPublisher_CalledWithEntry(t *testing.T) {
	var got *Entry
	l := &Logger{
		Logger:    log.New(&bytes.Buffer{}, "", 0),
		Publisher: func(e *Entry) { got = e },
	}

	l.Info("hi")

	if got == nil {
		t.Fatal("Publisher was not called")
	}
	if got.Message != "hi" {
		t.Fatalf("Publisher got Message = %q, want %q", got.Message, "hi")
	}
}

func TestSlogHandler_Conformance(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: log.New(&buf, "", 0)}

	results := func() []map[string]any {
		var out []map[string]any
		for _, line := range nonEmptyLines(buf.String()) {
			var e Entry
			if err := json.Unmarshal([]byte(line), &e); err != nil {
				t.Fatalf("output line is not valid JSON: %v (%q)", err, line)
			}
			m := map[string]any{slog.MessageKey: e.Message}
			if e.Timestamp != "" {
				ts, err := time.Parse(time.RFC3339, e.Timestamp)
				if err != nil {
					t.Fatalf("bad timestamp %q: %v", e.Timestamp, err)
				}
				m[slog.TimeKey] = ts
			}
			if e.Level != "" {
				m[slog.LevelKey] = e.Level
			}
			for k, v := range e.Attrs {
				m[k] = v
			}
			out = append(out, m)
		}
		return out
	}

	if err := slogtest.TestHandler(logger, results); err != nil {
		t.Error(err)
	}
}

func TestSlogHandler_GroupsNestAndElideWhenEmpty(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: log.New(&buf, "", 0)}
	sl := slog.New(logger)

	sl.WithGroup("empty").Info("no attrs under this group")
	sl.WithGroup("g").With("a", 1).Info("has attrs", "b", 2)

	lines := nonEmptyLines(buf.String())
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}

	var e1 Entry
	if err := json.Unmarshal([]byte(lines[0]), &e1); err != nil {
		t.Fatal(err)
	}
	if e1.Attrs != nil {
		t.Fatalf("Attrs = %v, want nil (empty group must be elided)", e1.Attrs)
	}

	var e2 Entry
	if err := json.Unmarshal([]byte(lines[1]), &e2); err != nil {
		t.Fatal(err)
	}
	g, ok := e2.Attrs["g"].(map[string]any)
	if !ok {
		t.Fatalf("Attrs[\"g\"] = %v (%T), want a nested map", e2.Attrs["g"], e2.Attrs["g"])
	}
	if g["a"] != float64(1) || g["b"] != float64(2) {
		t.Fatalf(`group "g" = %v, want {"a":1,"b":2}`, g)
	}
}

func TestPackageLevelFuncs_UseDefaultLogger(t *testing.T) {
	var buf bytes.Buffer
	orig := DefaultLogger.Logger
	DefaultLogger.Logger = log.New(&buf, "", 0)
	defer func() { DefaultLogger.Logger = orig }()

	Net("n")
	Debug("d")
	Info("i")
	Error("e")
	func() {
		defer func() { _ = recover() }()
		Panic("p")
	}()

	lines := nonEmptyLines(buf.String())
	if len(lines) != 5 {
		t.Fatalf("got %d lines, want 5: %q", len(lines), buf.String())
	}
}

func TestDefaults_ApplyWhenFieldsUnset(t *testing.T) {
	l := &Logger{}

	if l.serviceName() != "unknown" {
		t.Fatalf("serviceName() = %q, want %q", l.serviceName(), "unknown")
	}
	if l.hostname() == "" {
		t.Fatal("hostname() = \"\", want a non-empty default")
	}
	if l.hash() != shortHash(buildinfo.Hash) {
		t.Fatalf("hash() = %q, want %q", l.hash(), shortHash(buildinfo.Hash))
	}
	if l.output() != defaultOutput {
		t.Fatal("output() did not return the package default")
	}

	l2 := &Logger{ServiceName: "svc", Hostname: "host", Hash: "abcdef12"}
	if l2.serviceName() != "svc" {
		t.Fatalf("serviceName() = %q, want %q", l2.serviceName(), "svc")
	}
	if l2.hostname() != "host" {
		t.Fatalf("hostname() = %q, want %q", l2.hostname(), "host")
	}
	if l2.hash() != shortHash("abcdef12") {
		t.Fatalf("hash() = %q, want %q", l2.hash(), shortHash("abcdef12"))
	}
}

func TestLevelFor(t *testing.T) {
	cases := []struct {
		in   slog.Level
		want Level
	}{
		{slog.LevelDebug, DebugLevel},
		{slog.LevelDebug - 1, DebugLevel},
		{slog.LevelInfo, InfoLevel},
		{slog.LevelWarn, InfoLevel},
		{slog.LevelError, ErrorLevel},
		{slog.LevelError + 4, ErrorLevel},
	}
	for _, c := range cases {
		if got := levelFor(c.in); got != c.want {
			t.Errorf("levelFor(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func diffStr(got, want []Level) string {
	if len(got) != len(want) {
		return "length mismatch"
	}
	for i := range got {
		if got[i] != want[i] {
			return "mismatch at index"
		}
	}
	return ""
}
