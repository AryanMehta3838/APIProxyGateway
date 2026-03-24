package config

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Redis  RedisConfig  `yaml:"redis"`
	Routes []Route      `yaml:"routes"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type RedisConfig struct {
	Enabled bool   `yaml:"enabled"`
	Addr    string `yaml:"addr"`
}

type Route struct {
	Name        string          `yaml:"name"`
	PathPrefix  string          `yaml:"path_prefix"`
	StripPrefix string          `yaml:"strip_prefix"`
	TimeoutMS   int             `yaml:"timeout_ms"`
	Methods     []string        `yaml:"methods"`
	Upstream    string          `yaml:"upstream"`
	RateLimit   RateLimitConfig `yaml:"rate_limit"`
}

type RateLimitConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Requests      int    `yaml:"requests"`
	WindowSeconds int    `yaml:"window_seconds"`
	KeyStrategy   string `yaml:"key_strategy"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg Config
	if err := unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate config %q: %w", path, err)
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return errors.New("server.port must be between 1 and 65535")
	}
	if c.Redis.Enabled && strings.TrimSpace(c.Redis.Addr) == "" {
		return errors.New("redis.addr must not be empty when redis.enabled is true")
	}

	for i, route := range c.Routes {
		if route.Name == "" {
			return fmt.Errorf("routes[%d].name must not be empty", i)
		}
		if route.PathPrefix != "" && route.PathPrefix[0] != '/' {
			return fmt.Errorf("routes[%d].path_prefix must start with /", i)
		}
		if route.StripPrefix != "" && route.StripPrefix[0] != '/' {
			return fmt.Errorf("routes[%d].strip_prefix must start with /", i)
		}
		if route.TimeoutMS < 0 {
			return fmt.Errorf("routes[%d].timeout_ms must be zero or greater", i)
		}
		if route.RateLimit.Enabled {
			if route.RateLimit.Requests < 1 {
				return fmt.Errorf("routes[%d].rate_limit.requests must be greater than zero", i)
			}
			if route.RateLimit.WindowSeconds < 1 {
				return fmt.Errorf("routes[%d].rate_limit.window_seconds must be greater than zero", i)
			}
			if route.RateLimit.KeyStrategy != "" && !strings.EqualFold(route.RateLimit.KeyStrategy, "ip") {
				return fmt.Errorf("routes[%d].rate_limit.key_strategy must be ip", i)
			}
		}
		if len(route.Methods) == 0 {
			return fmt.Errorf("routes[%d].methods must not be empty", i)
		}
		for j, method := range route.Methods {
			if method == "" {
				return fmt.Errorf("routes[%d].methods[%d] must not be empty", i, j)
			}
			if !validMethod(method) {
				return fmt.Errorf("routes[%d].methods[%d] must be a valid HTTP method", i, j)
			}
		}
		if route.Upstream != "" {
			parsed, err := url.Parse(route.Upstream)
			if err != nil {
				return fmt.Errorf("routes[%d].upstream must be a valid URL: %w", i, err)
			}
			if !parsed.IsAbs() || (parsed.Scheme != "http" && parsed.Scheme != "https") {
				return fmt.Errorf("routes[%d].upstream must be an absolute http or https URL", i)
			}
		}
	}

	return nil
}

func validMethod(method string) bool {
	switch method {
	case http.MethodConnect, http.MethodDelete, http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPatch, http.MethodPost, http.MethodPut, http.MethodTrace:
		return true
	default:
		return false
	}
}
