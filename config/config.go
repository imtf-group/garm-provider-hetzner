package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

type Config struct {
	Location string `toml:"location"`
	Token    string `toml:"token"`
}

func NewConfig(cfgFile string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(cfgFile, &config); err != nil {
		return nil, fmt.Errorf("error decoding config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}
	return &config, nil
}

func (c *Config) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("missing token")
	}

	if c.Location == "" {
		return fmt.Errorf("missing location")
	}
	return nil
}
