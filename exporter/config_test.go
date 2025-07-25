package exporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// LoadConfig loads the configuration from a file
func TestConfig(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	config, err := LoadConfig("../config.example.toml")

	require.NoError(err)
	require.NotNil(config)
	require.Len(config.Controllers, 1)

	controller := config.Controllers[0]
	assert.Equal("my-controller", controller.Alias)
	assert.Equal("192.168.10.1", controller.Host)
	assert.Equal("admin", controller.Password)
	assert.EqualValues(8443, controller.Port)
}
