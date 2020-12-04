package ddpgx

import (
	"context"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

type poolCloser struct {
	*pgxpool.Pool
}

func (c *poolCloser) ConnInfo() *pgtype.ConnInfo {
	return nil
}

func (c *poolCloser) Close(ctx context.Context) error {
	startTime := time.Now()
	c.Pool.Close()
	tryTrace(ctx, startTime, "pgxpool:Close", nil, nil)
	return nil
}
