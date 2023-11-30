package client

import (
	"context"
	"log/slog"

	"github.com/digineo/triax-eoc-exporter/types"
)

type builderFunction func(context.Context, *Client) (types.Backend, error)

var backends []builderFunction

func Register(f builderFunction) {
	backends = append(backends, f)
}

func Try(ctx context.Context, client *Client) types.Backend {
	for i := range backends {
		backend, err := backends[i](ctx, client)
		if backend != nil {
			return backend
		}
		slog.Info("backend not working", "error", err)
	}

	return nil
}
