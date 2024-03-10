package config

import (
	"encoding/base64"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/lnashier/goarc/env"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"io"
)

// Config is the configuration
type Config struct {
	v *viper.Viper
}

// New is to create a new blank config object
func New() *Config {
	s := &Config{
		v: viper.New(),
	}
	return s
}

// NewWithPath is to create a new config object for given path+name
func NewWithPath(path string, name string) *Config {
	s := New()
	s.AddConfigPath(path)
	s.SetConfigName(name)
	return s
}

// NewCustomWatchedPath is to create a new config object; that has
// - custom listener attached
func NewCustomWatchedPath(path string, name string, f func(fsnotify.Event)) *Config {
	s := NewWithPath(path, name)
	s.WatchConfig()
	s.OnConfigChange(f)
	return s
}

// NewDefaultWatchedPath is to create a new config object; that has
// - no-op listener attached
func NewDefaultWatchedPath(path string, name string) *Config {
	return NewCustomWatchedPath(path, name, func(in fsnotify.Event) {})
}

// NewWithFile is to create a new config object for given file
func NewWithFile(file string) *Config {
	s := New()
	s.SetConfigFile(file)
	return s
}

// NewCustomWatchedFile is to create a new config object; that has
// - custom listener attached
func NewCustomWatchedFile(file string, f func(fsnotify.Event)) *Config {
	s := NewWithFile(file)
	s.WatchConfig()
	s.OnConfigChange(f)
	return s
}

// NewDefaultWatchedFile is to create a new config object; that has
// - no-op listener attached
func NewDefaultWatchedFile(file string) *Config {
	return NewCustomWatchedFile(file, func(in fsnotify.Event) {})
}

// Loaded is to load the configuration to the config object
func Loaded(s *Config) (*Config, error) {
	if err := s.ReadInConfig(); err != nil {
		return nil, errors.Wrapf(err, "Failed to load config")
	}
	return s, nil
}

// Get can retrieve any value given the key to use.
// Get is case-insensitive for a key.
// Get has the behavior of returning the value associated with the first
// place from where it is set. Config will check in the following order:
// override, flag, env, config file, key/value store, default
//
// Get returns an interface. For a specific value use one of the Get____ methods.
func (s *Config) Get(key string) any {
	return s.v.Get(key)
}

// GetBool returns the value associated with the key as a boolean.
func (s *Config) GetBool(key string) bool {
	return s.v.GetBool(key)
}

// GetInt returns the value associated with the key as an integer.
func (s *Config) GetInt(key string) int {
	return s.v.GetInt(key)
}

// GetString returns the value associated with the key as a string.
func (s *Config) GetString(key string) string {
	return s.v.GetString(key)
}

// GetStringDecoded returns the value associated with the key as a base64 decoded string.
func (s *Config) GetStringDecoded(key string) string {
	value := s.GetString(key)
	decodedValueBytes, _ := base64.StdEncoding.DecodeString(value)
	return string(decodedValueBytes)
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (s *Config) GetStringSlice(key string) []string {
	return s.v.GetStringSlice(key)
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (s *Config) GetStringMapString(key string) map[string]string {
	return s.v.GetStringMapString(key)
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func (s *Config) GetStringMap(key string) map[string]any {
	return s.v.GetStringMap(key)
}

// Set sets the value for the key in the override register.
// Set is case-insensitive for a key.
// Will be used instead of values obtained via
// flags, config file, ENV, default, or key/value store.
func (s *Config) Set(key string, value any) {
	s.v.Set(key, value)
}

// SetConfigType sets the type of the configuration returned by the
// remote source, e.g. "json".
func (s *Config) SetConfigType(in string) {
	s.v.SetConfigType(in)
}

// SetConfigName sets name for the config file.
// Does not include extension.
func (s *Config) SetConfigName(in string) {
	s.v.SetConfigName(in)
}

// SetConfigFile explicitly defines the path, name and extension of the config file.
// Config will use this and not check any of the config paths.
func (s *Config) SetConfigFile(in string) {
	s.v.SetConfigFile(in)
}

// AddConfigPath adds a path for Config to search for the config file in.
// Can be called multiple times to define multiple search paths.
func (s *Config) AddConfigPath(in string) {
	s.v.AddConfigPath(in)
}

// BindEnv binds a Config key to a ENV variable.
// ENV variables are case-sensitive.
// If only a key is provided, it will use the env key matching the key, uppercase.
// EnvPrefix will be used when set when env name is not provided.
func (s *Config) BindEnv(input ...string) error {
	return s.v.BindEnv(input...)
}

// WatchConfig watches for any change in underlying config source
func (s *Config) WatchConfig() {
	s.v.WatchConfig()
}

// OnConfigChange registers callback for config change
func (s *Config) OnConfigChange(run func(in fsnotify.Event)) {
	s.v.OnConfigChange(run)
}

// ReadConfig will read a configuration file, setting existing keys to nil if the
// key does not exist in the file.
func (s *Config) ReadConfig(in io.Reader) error {
	return s.v.ReadConfig(in)
}

// ReadInConfig will discover and load the configuration file from disk
// and key/value stores, searching in one of the defined paths.
func (s *Config) ReadInConfig() error {
	return s.v.ReadInConfig()
}

// Unmarshal unmarshals the config into a Config. Make sure that the tags
// on the fields of the structure are properly set.
func (s *Config) Unmarshal(rawVal any) error {
	return s.v.Unmarshal(rawVal)
}

// UnmarshalKey takes a single key and unmarshal it into given value.
func (s *Config) UnmarshalKey(key string, rawVal any) error {
	return s.v.UnmarshalKey(key, rawVal)
}

// SetEnvPrefix defines a prefix that ENVIRONMENT variables will use.
// E.g. if your prefix is "old", the env registry will look for env
// variables that start with "OLD_".
func (s *Config) SetEnvPrefix(in string) {
	s.v.SetEnvPrefix(in)
}

// Sub returns new Config instance representing a subtree of this instance.
// Sub is case-insensitive for a key.
func (s *Config) Sub(key string) *Config {
	nv := s.v.Sub(key)
	if nv != nil {
		return &Config{v: nv}
	}
	return nil
}

// Get loads app config, panic if app config load fails
func Get() *Config {
	cfg, err := Loaded(NewWithPath("configs", env.Get().String()))
	if err != nil {
		panic(fmt.Sprintf("failed to load app config: %v", err.Error()))
	}
	return cfg
}
