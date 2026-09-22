package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/lnashier/goarc/v2/x/env"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestNewWithFile_LoadsValues(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "app.yaml")
	writeFile(t, file, "name: goarc\nport: 8080\nenabled: true\n")

	cfg, err := Loaded(NewWithFile(file))
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.GetString("name"); got != "goarc" {
		t.Fatalf("GetString(name) = %q, want %q", got, "goarc")
	}
	if got := cfg.GetInt("port"); got != 8080 {
		t.Fatalf("GetInt(port) = %d, want 8080", got)
	}
	if got := cfg.GetBool("enabled"); !got {
		t.Fatal("GetBool(enabled) = false, want true")
	}
}

func TestNewWithPath_ViaGetAndGetters(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), `
name: goarc
tags:
  - a
  - b
labels:
  x: "1"
  y: "2"
meta:
  a: 1
  b: two
`)

	cfg, err := Loaded(NewWithPath(dir, "app"))
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Get("name"); got != "goarc" {
		t.Fatalf("Get(name) = %v, want %q", got, "goarc")
	}
	if got := cfg.GetStringSlice("tags"); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("GetStringSlice(tags) = %v, want [a b]", got)
	}
	if got := cfg.GetStringMapString("labels"); got["x"] != "1" || got["y"] != "2" {
		t.Fatalf("GetStringMapString(labels) = %v, want {x:1 y:2}", got)
	}
	if got := cfg.GetStringMap("meta"); got["a"] != 1 || got["b"] != "two" {
		t.Fatalf("GetStringMap(meta) = %v, want {a:1 b:two}", got)
	}
}

func TestConfig_ReadConfig_FromReader(t *testing.T) {
	cfg := New()
	cfg.SetConfigType("yaml")
	if err := cfg.ReadConfig(strings.NewReader("name: from-reader\n")); err != nil {
		t.Fatal(err)
	}
	if got := cfg.GetString("name"); got != "from-reader" {
		t.Fatalf("GetString(name) = %q, want %q", got, "from-reader")
	}
}

func TestConfig_UnmarshalKey(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), "db:\n  host: localhost\n  port: 5432\n")

	cfg, err := Loaded(NewWithPath(dir, "app"))
	if err != nil {
		t.Fatal(err)
	}
	var db struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	}
	if err := cfg.UnmarshalKey("db", &db); err != nil {
		t.Fatal(err)
	}
	if db.Host != "localhost" || db.Port != 5432 {
		t.Fatalf("UnmarshalKey(db) = %+v, want {localhost 5432}", db)
	}
}

func TestConfig_SetEnvPrefix(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), "name: default\n")

	cfg, err := Loaded(NewWithPath(dir, "app"))
	if err != nil {
		t.Fatal(err)
	}
	cfg.SetEnvPrefix("goarc")
	t.Setenv("GOARC_NAME", "prefixed")
	if err := cfg.BindEnv("name"); err != nil {
		t.Fatal(err)
	}
	if got := cfg.GetString("name"); got != "prefixed" {
		t.Fatalf("GetString(name) = %q, want %q", got, "prefixed")
	}
}

func TestNewDefaultWatchedPath_NoOpListenerDoesNotPanic(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), "name: goarc\n")

	cfg := NewDefaultWatchedPath(dir, "app")
	if _, err := Loaded(cfg); err != nil {
		t.Fatal(err)
	}
}

func TestNewCustomWatchedPath_ListenerIsRegistered(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "app.yaml")
	writeFile(t, file, "name: goarc\n")

	called := make(chan fsnotify.Event, 1)
	cfg := NewCustomWatchedPath(dir, "app", func(e fsnotify.Event) { called <- e })
	if _, err := Loaded(cfg); err != nil {
		t.Fatal(err)
	}

	writeFile(t, file, "name: updated\n")

	select {
	case <-called:
	case <-time.After(3 * time.Second):
		t.Fatal("OnConfigChange listener was never called after the file changed")
	}
}

func TestNewWithPath_LoadsValues(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), "name: goarc\n")

	cfg, err := Loaded(NewWithPath(dir, "app"))
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.GetString("name"); got != "goarc" {
		t.Fatalf("GetString(name) = %q, want %q", got, "goarc")
	}
}

func TestLoaded_ReturnsErrorWhenFileMissing(t *testing.T) {
	dir := t.TempDir()
	if _, err := Loaded(NewWithPath(dir, "missing")); err == nil {
		t.Fatal("Loaded() = nil error, want one when the config file does not exist")
	}
}

func TestConfig_Set_OverridesLoadedValue(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), "name: goarc\n")

	cfg, err := Loaded(NewWithPath(dir, "app"))
	if err != nil {
		t.Fatal(err)
	}
	cfg.Set("name", "overridden")
	if got := cfg.GetString("name"); got != "overridden" {
		t.Fatalf("GetString(name) = %q, want %q", got, "overridden")
	}
}

func TestConfig_GetStringDecoded(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), "secret: aGVsbG8=\n") // base64("hello")

	cfg, err := Loaded(NewWithPath(dir, "app"))
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.GetStringDecoded("secret"); got != "hello" {
		t.Fatalf("GetStringDecoded(secret) = %q, want %q", got, "hello")
	}
}

func TestConfig_Sub(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), "db:\n  host: localhost\n  port: 5432\n")

	cfg, err := Loaded(NewWithPath(dir, "app"))
	if err != nil {
		t.Fatal(err)
	}
	sub := cfg.Sub("db")
	if sub == nil {
		t.Fatal("Sub(db) = nil, want a sub-config")
	}
	if got := sub.GetString("host"); got != "localhost" {
		t.Fatalf("sub.GetString(host) = %q, want %q", got, "localhost")
	}

	if cfg.Sub("missing") != nil {
		t.Fatal("Sub(missing) != nil, want nil for a key that doesn't exist")
	}
}

func TestConfig_Unmarshal(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), "name: goarc\nport: 8080\n")

	cfg, err := Loaded(NewWithPath(dir, "app"))
	if err != nil {
		t.Fatal(err)
	}

	var target struct {
		Name string `mapstructure:"name"`
		Port int    `mapstructure:"port"`
	}
	if err := cfg.Unmarshal(&target); err != nil {
		t.Fatal(err)
	}
	if target.Name != "goarc" || target.Port != 8080 {
		t.Fatalf("Unmarshal() = %+v, want {goarc 8080}", target)
	}
}

func TestConfig_BindEnv(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "app.yaml"), "name: default\n")

	cfg, err := Loaded(NewWithPath(dir, "app"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("APP_NAME", "from-env")
	if err := cfg.BindEnv("name", "APP_NAME"); err != nil {
		t.Fatal(err)
	}
	if got := cfg.GetString("name"); got != "from-env" {
		t.Fatalf("GetString(name) = %q, want %q (env should take precedence over the file)", got, "from-env")
	}
}

func TestNewDefaultWatchedFile_NoOpListenerDoesNotPanic(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "app.yaml")
	writeFile(t, file, "name: goarc\n")

	cfg := NewDefaultWatchedFile(file)
	if _, err := Loaded(cfg); err != nil {
		t.Fatal(err)
	}
}

func TestNewCustomWatchedFile_ListenerIsRegistered(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "app.yaml")
	writeFile(t, file, "name: goarc\n")

	called := make(chan fsnotify.Event, 1)
	cfg := NewCustomWatchedFile(file, func(e fsnotify.Event) { called <- e })
	if _, err := Loaded(cfg); err != nil {
		t.Fatal(err)
	}

	writeFile(t, file, "name: updated\n")

	select {
	case <-called:
	case <-time.After(3 * time.Second):
		t.Fatal("OnConfigChange listener was never called after the file changed")
	}
}

func TestGet_LoadsFromConfigsDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "configs", env.Get().String()+".yaml"), "name: goarc\n")
	t.Chdir(dir)

	cfg := Get()
	if got := cfg.GetString("name"); got != "goarc" {
		t.Fatalf("GetString(name) = %q, want %q", got, "goarc")
	}
}

func TestGet_PanicsWhenConfigMissing(t *testing.T) {
	t.Chdir(t.TempDir())

	defer func() {
		if recover() == nil {
			t.Fatal("Get() did not panic when the config file is missing")
		}
	}()
	Get()
}
