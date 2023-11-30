package v2

import (
	"context"
	"net/http"

	"github.com/digineo/triax-eoc-exporter/client"
	"github.com/digineo/triax-eoc-exporter/types"
)

func init() {
	client.Register(New)
}

func New(ctx context.Context, c *client.Client) (types.Backend, error) {
	b := backend{c}

	req := loginRequest{Username: c.Username, Password: c.Password}
	res := loginResponse{}
	_, err := c.ApiRequestRaw(ctx, http.MethodPost, loginPath, &req, &res)

	if err != nil {
		return nil, err
	}

	c.SetCookie(res.Cookie)

	return &b, nil
}

type backend struct {
	*client.Client
}
