package config

import (
	"os"
	"reflect"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	os.Setenv("OPENAI_RATELIMIT", "50")
	defer os.Unsetenv("OPENAI_API_KEY")
	defer os.Unsetenv("OPENAI_RATELIMIT")

	cfg := Load()

	if cfg.OPENAI.API_KEY != "test-key" {
		t.Errorf("expected API_KEY=test-key, got %s", cfg.OPENAI.API_KEY)
	}
	if cfg.OPENAI.RATELIMIT != 50 {
		t.Errorf("expected RATELIMIT=50, got %d", cfg.OPENAI.RATELIMIT)
	}
}

func TestLoadDefaultRateLimit(t *testing.T) {
	cfg := Load()
	if cfg.OPENAI.RATELIMIT == 0 {
		t.Error("RATELIMIT should default to 100")
	}
}

func TestFlattenKeys(t *testing.T) {
	type inner struct{ B string }
	type outer struct {
		A string
		C inner
	}
	keys := flattenKeys("", reflect.ValueOf(outer{}))
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d: %v", len(keys), keys)
	}
}
