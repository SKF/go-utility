package ddpgx

import (
	"context"
	"time"

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

func Connect(ctx context.Context, serviceName, url string, tracerOpts ...TracerOpt) (Connection, error) {
	trace := newTracer(serviceName, driverPgx, tracerOpts...)

	startTime := time.Now()

	conn, err := pgx.Connect(ctx, url)
	trace.TryTrace(ctx, startTime, "Connect", nil, err)

	return &traceConn{
		conn:  conn,
		trace: trace,
	}, err
}

func ConnectPoolConfig(ctx context.Context, serviceName string, config *pgxpool.Config, tracerOpts ...TracerOpt) (Connection, error) {
	trace := newTracer(serviceName, driverPgxPool, tracerOpts...)

	startTime := time.Now()

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		trace.TryTrace(ctx, startTime, "ConnectPoolConfig", nil, err)
		return nil, err
	}

	trace.TryTrace(ctx, startTime, "ConnectPoolConfig", nil, nil)

	return &traceConn{
		conn:  &poolCloser{pool},
		trace: trace,
	}, nil
}
