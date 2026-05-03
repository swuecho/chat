// Package config provides application configuration via environment variables.
package config

import (
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

// Load loads configuration from environment variables into AppConfig.
func Load(logger interface{ Fatal(...interface{}) }) AppConfig {
	cfg := AppConfig{}
	for _, key := range flattenKeys("", reflect.ValueOf(cfg)) {
		envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		if err := viper.BindEnv(key, envKey); err != nil {
			logger.Fatal("config: unable to bind env: " + err.Error())
		}
	}
	viper.AutomaticEnv()
	if err := viper.Unmarshal(&cfg); err != nil {
		logger.Fatal("config: unable to decode into struct: " + err.Error())
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
