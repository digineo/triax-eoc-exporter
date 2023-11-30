package v2

import (
	"context"
	"errors"
	"net"
	"net/http"
)

type UpdateRequest struct {
	Name string `json:"name"`
}

const hexDigit = "0123456789abcdef"

// UpdateNode updates the name of the given node
func (c *backend) UpdateNode(ctx context.Context, mac net.HardwareAddr, req UpdateRequest) error {
	if len(mac) == 0 {
		return errors.New("invalid MAC address")
	}

	buf := make([]byte, 0, len(mac)*3-1)
	for i, b := range mac {
		if i > 0 {
			buf = append(buf, '_')
		}
		buf = append(buf, hexDigit[b>>4])
		buf = append(buf, hexDigit[b&0xF])
	}

	err := c.ApiRequest(ctx, http.MethodPut, "config/nodes/node_"+string(buf)+"/", req, nil)
	if err != nil {
		return err
	}

	return c.ApiRequest(ctx, http.MethodPost, "config/nodes/commit/", nil, nil)
}
