package ddpgx

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	dd_ext "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

type traceTx struct {
	parent pgx.Tx
	trace  *internalTracer
}

func (t *traceTx) Begin(ctx context.Context) (pgx.Tx, error) {
	t.trace.Start()
	tx, err := t.parent.Begin(ctx)
	t.trace.TryTrace(ctx, "Begin", nil, err)

	return &traceTx{
		parent: tx,
		trace:  t.trace,
	}, err
}

func (t *traceTx) Commit(ctx context.Context) error {
	t.trace.Start()
	err := t.parent.Commit(ctx)
	t.trace.TryTrace(ctx, "Commit", nil, err)

	return err
}

func (t *traceTx) Rollback(ctx context.Context) error {
	t.trace.Start()
	err := t.parent.Rollback(ctx)
	t.trace.TryTrace(ctx, "Rollback", nil, err)

	return err
}

func (t *traceTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	t.trace.Start()
	rowsAffected, err := t.parent.CopyFrom(ctx, tableName, columnNames, rowSrc)
	t.trace.TryTrace(ctx, "CopyFrom", nil, err)

	return rowsAffected, err
}

func (t *traceTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	t.trace.Start()
	results := t.parent.SendBatch(ctx, b)
	t.trace.TryTrace(ctx, "SendBatch", nil, nil)

	return results
}

func (t *traceTx) LargeObjects() pgx.LargeObjects {
	return t.parent.LargeObjects()
}

func (t *traceTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	t.trace.Start()
	stmt, err := t.parent.Prepare(ctx, name, sql)
	t.trace.TryTrace(ctx, "Prepare", nil, err)

	return stmt, err
}

func (t *traceTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error) {
	t.trace.Start()
	tag, err := t.parent.Exec(ctx, sql, arguments...)
	t.trace.TryTrace(ctx, "Exec", nil, err)

	return tag, err
}

func (t *traceTx) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	t.trace.Start()
	rows, err := t.parent.Query(ctx, query, args...)

	metadata := argsToAttributes(args...)
	metadata[dd_ext.SQLQuery] = query
	t.trace.TryTrace(ctx, "Query", metadata, err)

	return rows, err
}

func (t *traceTx) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	t.trace.Start()
	row := t.parent.QueryRow(ctx, query, args...)

	metadata := argsToAttributes(args...)
	metadata[dd_ext.SQLQuery] = query
	t.trace.TryTrace(ctx, "QueryRow", metadata, nil)

	return row
}

func (t *traceTx) Conn() *pgx.Conn {
	return t.parent.Conn()
}
