package triax

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSomething(t *testing.T) {
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
