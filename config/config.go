package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Bind string // HTTP endpoint to listen on

	// list of Triax EoC controllers
	Controllers []struct {
		Endpoint string // URL incl. credentials to reach the controller
		Insecure bool   // set to true if the TLS certificate validation should be skipped
		MAC      string // used as identifier in export
	} `toml:"eoc-controller"`
}

func LoadFile(file string) (*Config, error) {
	cfg := Config{}
	if _, err := toml.DecodeFile(file, &cfg); err != nil {
		return nil, fmt.Errorf("loading config file %q failed: %w", file, err)
	}
	return &cfg, nil
}
