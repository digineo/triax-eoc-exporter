package exporter

import (
	"fmt"
	"net/url"

	"github.com/BurntSushi/toml"
	"github.com/digineo/triax-eoc-exporter/triax"
)

type Config struct {
	// list of Triax EoC controllers
	Controllers []Controller `toml:"eoc-controller"`
}

type Controller struct {
	Alias    string
	Host     string
	Password string
}

func LoadConfig(file string) (*Config, error) {
	cfg := Config{}

	if _, err := toml.DecodeFile(file, &cfg); err != nil {
		return nil, fmt.Errorf("loading config file %q failed: %w", file, err)
	}
	return &cfg, nil
}

func (cfg *Config) getClient(target string) (*triax.Client, error) {
	for _, ctrl := range cfg.Controllers {
		if target == ctrl.Alias || target == ctrl.Host {
			return triax.NewClient(&url.URL{
				Scheme: "https",
				User:   url.UserPassword("admin", ctrl.Password),
				Host:   ctrl.Host,
				Path:   "/",
			})
		}
	}

	return nil, nil
}
