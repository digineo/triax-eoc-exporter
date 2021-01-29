package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Bind string // HTTP endpoint to listen on

	// list of Triax EoC controllers
	Controllers []struct {
		Alias    string // Optional alias
		Host     string
		Username string
		Password string
	} `toml:"eoc-controller"`
}

func LoadFile(file string) (*Config, error) {
	cfg := Config{}

	if _, err := toml.DecodeFile(file, &cfg); err != nil {
		return nil, fmt.Errorf("loading config file %q failed: %w", file, err)
	}
	return &cfg, nil
}
