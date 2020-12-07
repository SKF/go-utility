package ddpgx

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	dd_ext "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

type traceConn struct {
	conn  Connection
	trace *internalTracer
}

func (o *traceConn) ConnInfo() *pgtype.ConnInfo {
	return o.conn.ConnInfo()
}

func (o *traceConn) Begin(ctx context.Context) (pgx.Tx, error) {
	o.trace.Start()
	tx, err := o.conn.Begin(ctx)
	o.trace.TryTrace(ctx, "Begin", nil, err)

	return &traceTx{
		parent: tx,
		trace:  newTracer(o.trace.ServiceName(), driverPgxTx),
	}, err
}

func (o *traceConn) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	o.trace.Start()
	tag, err := o.conn.Exec(ctx, query, args...)

	metadata := argsToAttributes(args...)
	metadata[dd_ext.SQLQuery] = query
	o.trace.TryTrace(ctx, "Exec", metadata, err)

	return tag, err
}

func (o *traceConn) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	o.trace.Start()
	rows, err := o.conn.Query(ctx, query, args...)

	metadata := argsToAttributes(args...)
	metadata[dd_ext.SQLQuery] = query
	o.trace.TryTrace(ctx, "Query", metadata, err)

	return rows, err
}

func (o *traceConn) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	o.trace.Start()
	row := o.conn.QueryRow(ctx, query, args...)

	metadata := argsToAttributes(args...)
	metadata[dd_ext.SQLQuery] = query
	o.trace.TryTrace(ctx, "QueryRow", metadata, nil)

	return row
}

func (o *traceConn) Close(ctx context.Context) error {
	o.trace.Start()
	err := o.conn.Close(ctx)
	o.trace.TryTrace(ctx, "Close", nil, err)

	return err
}
