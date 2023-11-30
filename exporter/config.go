package exporter

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/digineo/triax-eoc-exporter/client"
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
	client   *client.Client
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
func (cfg *Config) getClient(target string) (*client.Client, error) {
	for i := range cfg.Controllers {
		ctrl := &cfg.Controllers[i]
		if target == ctrl.Alias || target == ctrl.Host {
			return ctrl.getClient()
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

func (ctrl *Controller) getClient() (*client.Client, error) {
	if ctrl.client == nil {

		c, err := client.NewClient(ctrl.url())
		if err != nil {
			return nil, err
		}
		ctrl.client = c
	}

	return ctrl.client, nil
}
