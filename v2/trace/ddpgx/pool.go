package ddpgx

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

type poolCloser struct {
	*pgxpool.Pool
}

func (c *poolCloser) ConnInfo() *pgtype.ConnInfo {
	return nil
}

func (c *poolCloser) Close(_ context.Context) error {
	c.Pool.Close()
	return nil
}
