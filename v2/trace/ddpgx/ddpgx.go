package ddpgx

import (
	"context"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Connection interface {
	ConnInfo() *pgtype.ConnInfo
	Begin(context.Context) (pgx.Tx, error)
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Close(context.Context) error
}

func Connect(ctx context.Context, serviceName, url string) (Connection, error) {
	startTime := time.Now()
	conn, err := pgx.Connect(ctx, url)
	tryTrace(ctx, startTime, serviceName, "pgx", "Connect", nil, err)

	return &traceConn{
		conn:        conn,
		serviceName: serviceName,
	}, err
}

func ConnectPoolConfig(ctx context.Context, serviceName string, config *pgxpool.Config) (Connection, error) {
	startTime := time.Now()
	pool, err := pgxpool.ConnectConfig(ctx, config)

	if err != nil {
		tryTrace(ctx, startTime, serviceName, "pgxpool", "ConnectPoolConfig", nil, err)
		return nil, err
	}

	tryTrace(ctx, startTime, serviceName, "pgxpool", "ConnectPoolConfig", nil, nil)

	return &traceConn{
		conn:        &poolCloser{pool},
		serviceName: serviceName,
	}, nil
}
