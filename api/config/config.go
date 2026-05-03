// Package config provides application configuration via environment variables.
package config

import (
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

// AppConfig holds all application configuration.
type AppConfig struct {
	OPENAI struct {
		API_KEY   string
		RATELIMIT int
		PROXY_URL string
	}
	CLAUDE struct {
		API_KEY string
	}
	PG struct {
		HOST string
		PORT int
		USER string
		PASS string
		DB   string
	}
}

// Load reads configuration from environment variables into AppConfig.
func Load() AppConfig {
	cfg := AppConfig{}
	for _, key := range flattenKeys("", reflect.ValueOf(cfg)) {
		envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		if err := viper.BindEnv(key, envKey); err != nil {
			slog.Error("config: unable to bind env", "key", key, "error", err)
			os.Exit(1)
		}
	}
	viper.AutomaticEnv()
	if err := viper.Unmarshal(&cfg); err != nil {
		slog.Error("config: unable to decode", "error", err)
		os.Exit(1)
	}
	if cfg.OPENAI.RATELIMIT == 0 {
		cfg.OPENAI.RATELIMIT = 100
	}
	return cfg
}

func flattenKeys(prefix string, v reflect.Value) []string {
	switch v.Kind() {
	case reflect.Struct:
		var keys []string
		for i := 0; i < v.NumField(); i++ {
			name := v.Type().Field(i).Name
			keys = append(keys, flattenKeys(prefix+name+".", v.Field(i))...)
		}
		return keys
	default:
		return []string{prefix[:len(prefix)-1]}
	}
}
