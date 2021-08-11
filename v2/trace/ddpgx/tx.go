package ddpgx

import (
	"context"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	dd_ext "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

type traceTx struct {
	parent pgx.Tx
	trace  internalTracer
}

func (t *traceTx) Begin(ctx context.Context) (pgx.Tx, error) {
	startTime := time.Now()
	tx, err := t.parent.Begin(ctx)
	t.trace.TryTrace(ctx, startTime, "Begin", nil, err)

	return &traceTx{
		parent: tx,
		trace:  t.trace,
	}, err
}

func (t *traceTx) BeginFunc(ctx context.Context, f func(pgx.Tx) error) error {
	startTime := time.Now()
	err := t.parent.BeginFunc(ctx, f)

	t.trace.TryTrace(ctx, startTime, "BeginFunc", nil, err)

	return err
}

func (t *traceTx) Commit(ctx context.Context) error {
	startTime := time.Now()
	err := t.parent.Commit(ctx)
	t.trace.TryTrace(ctx, startTime, "Commit", nil, err)

	return err
}

func (t *traceTx) Rollback(ctx context.Context) error {
	startTime := time.Now()
	err := t.parent.Rollback(ctx)
	t.trace.TryTrace(ctx, startTime, "Rollback", nil, err)

	return err
}

func (t *traceTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	startTime := time.Now()
	rowsAffected, err := t.parent.CopyFrom(ctx, tableName, columnNames, rowSrc)
	t.trace.TryTrace(ctx, startTime, "CopyFrom", nil, err)

	return rowsAffected, err
}

func (t *traceTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	startTime := time.Now()
	results := t.parent.SendBatch(ctx, b)
	t.trace.TryTrace(ctx, startTime, "SendBatch", nil, nil)

	return results
}

func (t *traceTx) LargeObjects() pgx.LargeObjects {
	return t.parent.LargeObjects()
}

func (t *traceTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	startTime := time.Now()
	stmt, err := t.parent.Prepare(ctx, name, sql)
	t.trace.TryTrace(ctx, startTime, "Prepare", nil, err)

	return stmt, err
}

func (t *traceTx) Exec(ctx context.Context, query string, args ...interface{}) (commandTag pgconn.CommandTag, err error) {
	startTime := time.Now()
	tag, err := t.parent.Exec(ctx, query, args...)

	metadata := argsToAttributes(args...)
	metadata[dd_ext.SQLQuery] = query
	t.trace.TryTrace(ctx, startTime, "Exec", metadata, err)

	return tag, err
}

func (t *traceTx) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	startTime := time.Now()
	rows, err := t.parent.Query(ctx, query, args...)

	metadata := argsToAttributes(args...)
	metadata[dd_ext.SQLQuery] = query
	t.trace.TryTrace(ctx, startTime, "Query", metadata, err)

	return rows, err
}

func (t *traceTx) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	startTime := time.Now()
	row := t.parent.QueryRow(ctx, query, args...)

	metadata := argsToAttributes(args...)
	metadata[dd_ext.SQLQuery] = query
	t.trace.TryTrace(ctx, startTime, "QueryRow", metadata, nil)

	return row
}

func (t *traceTx) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	startTime := time.Now()
	tag, err := t.parent.QueryFunc(ctx, sql, args, scans, f)

	t.trace.TryTrace(ctx, startTime, "QueryFunc", nil, err)

	return tag, err
}

func (t *traceTx) Conn() *pgx.Conn {
	return t.parent.Conn()
}
