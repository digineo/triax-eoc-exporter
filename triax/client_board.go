package triax

import (
	"context"
)

func (c *Client) Board(ctx context.Context) (board Board, err error) {
	err = c.Get(ctx, boardPath, &board)
	return
}
