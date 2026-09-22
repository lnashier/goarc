// Package config is a thin, stable wrapper around Viper: NewWithPath,
// NewWithFile and their watched variants build a Config, Loaded reads it,
// and the Get*/Set/Unmarshal methods mirror Viper's own. Get loads the
// process's own config from ./configs/<env>.<ext>, based on x/env.Get.
package config
