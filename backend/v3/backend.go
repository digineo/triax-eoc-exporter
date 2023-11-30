package v3

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/digineo/triax-eoc-exporter/client"
	"github.com/digineo/triax-eoc-exporter/types"
)

func init() {
	client.Register(New)
}

type backend struct {
	*client.Client
}

func New(ctx context.Context, c *client.Client) (types.Backend, error) {
	b := backend{c}

	req := loginRequest{Username: c.Username, Password: c.Password}
	res := loginResponse{}
	httpResponse, err := c.ApiRequestRaw(ctx, http.MethodPost, loginPath, &req, &res)

	if err != nil {
		return nil, err
	}

	if !res.Status {
		return nil, fmt.Errorf("login failed: %s", res.Message)
	}

	cookie := httpResponse.Header.Get("Set-Cookie")
	if i := strings.Index(cookie, ";"); i > 0 {
		cookie = cookie[:i]
	}

	c.SetCookie(cookie)
	return &b, nil
}
