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

func Connect(ctx context.Context, serviceName, url string) (Connection, error) {
	trace := newTracer(serviceName, driverPgx)
	return connect(ctx, url, trace)
}

func ConnectWithoutSpanValueEscape(ctx context.Context, serviceName, url string) (Connection, error) {
	trace := newTracer(serviceName, driverPgx).WithoutSpanTagValueEscape()
	return connect(ctx, url, trace)
}

func connect(ctx context.Context, url string, trace internalTracer) (Connection, error) {
	startTime := time.Now()

	conn, err := pgx.Connect(ctx, url)
	trace.TryTrace(ctx, startTime, "Connect", nil, err)

	return &traceConn{
		conn:  conn,
		trace: trace,
	}, err
}

func ConnectPoolConfig(ctx context.Context, serviceName string, config *pgxpool.Config) (Connection, error) {
	trace := newTracer(serviceName, driverPgxPool)
	return connectPoolConfig(ctx, config, trace)
}

func ConnectPoolConfigWithoutSpanValueEscape(ctx context.Context, serviceName string, config *pgxpool.Config) (Connection, error) {
	trace := newTracer(serviceName, driverPgxPool).WithoutSpanTagValueEscape()
	return connectPoolConfig(ctx, config, trace)
}

func connectPoolConfig(ctx context.Context, config *pgxpool.Config, trace internalTracer) (Connection, error) {
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
