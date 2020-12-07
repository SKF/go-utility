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
}

func (t *traceTx) Begin(ctx context.Context) (pgx.Tx, error) {
	startTime := time.Now()
	tx, err := t.parent.Begin(ctx)
	tryTrace(ctx, startTime, "pgx:tx:Begin", nil, err)

	return &traceTx{parent: tx}, err
}

func (t *traceTx) Commit(ctx context.Context) error {
	startTime := time.Now()
	err := t.parent.Commit(ctx)
	tryTrace(ctx, startTime, "pgx:tx:Commit", nil, err)

	return err
}

func (t *traceTx) Rollback(ctx context.Context) error {
	startTime := time.Now()
	err := t.parent.Rollback(ctx)
	tryTrace(ctx, startTime, "pgx:tx:Rollback", nil, err)

	return err
}

func (t *traceTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	startTime := time.Now()
	rowsAffected, err := t.parent.CopyFrom(ctx, tableName, columnNames, rowSrc)
	tryTrace(ctx, startTime, "pgx:tx:CopyFrom", nil, err)

	return rowsAffected, err
}

func (t *traceTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	startTime := time.Now()
	results := t.parent.SendBatch(ctx, b)
	tryTrace(ctx, startTime, "pgx:tx:SendBatch", nil, nil)

	return results
}

func (t *traceTx) LargeObjects() pgx.LargeObjects {
	return t.parent.LargeObjects()
}

func (t *traceTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	startTime := time.Now()
	stmt, err := t.parent.Prepare(ctx, name, sql)
	tryTrace(ctx, startTime, "pgx:tx:Prepare", nil, err)

	return stmt, err
}

func (t *traceTx) Exec(ctx context.Context, sql string, arguments ...interface{}) (commandTag pgconn.CommandTag, err error) {
	startTime := time.Now()
	tag, err := t.parent.Exec(ctx, sql, arguments...)
	tryTrace(ctx, startTime, "pgx:tx:Exec", nil, err)

	return tag, err
}

func (t *traceTx) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	startTime := time.Now()
	rows, err := t.parent.Query(ctx, query, args...)

	metadata := argsToAttributes(args...)
	metadata[dd_ext.SQLQuery] = query
	tryTrace(ctx, startTime, "pgx:tx:Query", metadata, err)

	return rows, err
}

func (t *traceTx) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	startTime := time.Now()
	row := t.parent.QueryRow(ctx, query, args...)

	metadata := argsToAttributes(args...)
	metadata[dd_ext.SQLQuery] = query
	tryTrace(ctx, startTime, "pgx:tx:QueryRow", metadata, nil)

	return row
}

func (t *traceTx) Conn() *pgx.Conn {
	return t.parent.Conn()
}
