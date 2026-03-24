package config

import "gopkg.in/yaml.v3"

func unmarshal(data []byte, cfg *Config) error {
	return yaml.Unmarshal(data, cfg)
}
