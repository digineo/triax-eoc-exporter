package exporter

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/digineo/triax-eoc-exporter/triax"
)

const defaultPort = 443

type Config struct {
	// list of Triax EoC controllers
	Controllers []Controller `toml:"eoc-controller"`
}

type Controller struct {
	Alias    string
	Host     string
	Port     uint16
	Password string
}

// LoadConfig loads the configuration from a file
func LoadConfig(file string) (*Config, error) {
	cfg := Config{}

	if _, err := toml.DecodeFile(file, &cfg); err != nil {
		return nil, fmt.Errorf("loading config file %q failed: %w", file, err)
	}

	return &cfg, nil
}

// getClient builds a client
func (cfg *Config) getClient(target string) (*triax.Client, error) {
	for _, ctrl := range cfg.Controllers {
		if target == ctrl.Alias || target == ctrl.Host {
			return triax.NewClient(ctrl.url())
		}
	}

	return nil, nil
}

// url build the URL
func (ctrl *Controller) url() *url.URL {
	host := ctrl.Host

	if ctrl.Port > 0 {
		host = net.JoinHostPort(host, strconv.Itoa(int(ctrl.Port)))
	}

	return &url.URL{
		Scheme: "https",
		User:   url.UserPassword("admin", ctrl.Password),
		Host:   host,
		Path:   "/",
	}
}
