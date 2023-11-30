package v2

import (
	"context"
)

func (c *backend) Board(ctx context.Context) (board Board, err error) {
	err = c.Get(ctx, boardPath, &board)
	return
}
