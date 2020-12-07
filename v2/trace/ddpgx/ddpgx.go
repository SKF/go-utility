package ddpgx

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	driverPgx     string = "pgx"
	driverPgxPool string = "pgxpool"
	driverPgxTx   string = "pgxtx"
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
	trace := newTracer(serviceName, driverPgx)
	trace.Start()

	conn, err := pgx.Connect(ctx, url)
	trace.TryTrace(ctx, "Connect", nil, err)

	return &traceConn{
		conn:  conn,
		trace: trace,
	}, err
}

func ConnectPoolConfig(ctx context.Context, serviceName string, config *pgxpool.Config) (Connection, error) {
	trace := newTracer(serviceName, driverPgxPool)
	trace.Start()

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		trace.TryTrace(ctx, "ConnectPoolConfig", nil, err)
		return nil, err
	}

	trace.TryTrace(ctx, "ConnectPoolConfig", nil, nil)

	return &traceConn{
		conn:  &poolCloser{pool},
		trace: trace,
	}, nil
}
