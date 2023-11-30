package v2

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodes(t *testing.T) {
	files := []string{
		"2.7/nodes.json",
		"2.8/nodes.json",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			file, err := os.Open("testdata/" + file)
			require.NoError(t, err)

			res := nodeStatusResponse{}
			assert.NoError(t, json.NewDecoder(file).Decode(&res))
		})
	}
}

func TestInfo(t *testing.T) {
	file, err := os.Open("testdata/2.8/info.json")
	require.NoError(t, err)

	res := sysinfoResponse{}
	require.NoError(t, json.NewDecoder(file).Decode(&res))

	assert.EqualValues(t, 290, res.Uptime)
	assert.EqualValues(t, 0.24462475, res.Load)
}
