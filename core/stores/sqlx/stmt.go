package sqlx

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/core/timex"
)

const defaultSlowThreshold = time.Millisecond * 500

var (
	slowThreshold = syncx.ForAtomicDuration(defaultSlowThreshold)
	logStmtSql    = syncx.ForAtomicBool(true)
	logSlowSql    = syncx.ForAtomicBool(true)
)

// DisableLog disables logging of sql statements, includes info and slow logs.
func DisableLog() {
	logStmtSql.Set(false)
	logSlowSql.Set(false)
}

// DisableStmtLog disables info logging of sql statements, but keeps slow logs.
func DisableStmtLog() {
	logStmtSql.Set(false)
}

// SetSlowThreshold sets the slow threshold.
func SetSlowThreshold(threshold time.Duration) {
	slowThreshold.Set(threshold)
}

func exec(ctx context.Context, conn sessionConn, q string, args ...any) (sql.Result, error) {
	guard := newGuard("exec")
	if err := guard.start(q, args...); err != nil {
		return nil, err
	}

	result, err := conn.ExecContext(ctx, q, args...)
	guard.finish(ctx, err)

	return result, err
}

func execStmt(ctx context.Context, conn stmtConn, q string, args ...any) (sql.Result, error) {
	guard := newGuard("execStmt")
	if err := guard.start(q, args...); err != nil {
		return nil, err
	}

	result, err := conn.ExecContext(ctx, args...)
	guard.finish(ctx, err)

	return result, err
}

func query(ctx context.Context, conn sessionConn, scanner func(*sql.Rows) error, q string, args ...any) error {
	guard := newGuard("query")
	if err := guard.start(q, args...); err != nil {
		return err
	}

	rows, err := conn.QueryContext(ctx, q, args...)
	guard.finish(ctx, err)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanner(rows)
}

func queryStmt(ctx context.Context, conn stmtConn, scanner func(*sql.Rows) error, q string, args ...any) error {
	guard := newGuard("queryStmt")
	if err := guard.start(q, args...); err != nil {
		return err
	}

	rows, err := conn.QueryContext(ctx, args...)
	guard.finish(ctx, err)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanner(rows)
}

type (
	sqlGuard interface {
		start(q string, args ...any) error
		finish(ctx context.Context, err error)
	}

	nilGuard struct {
		command   string
		startTime time.Duration
	}

	realSqlGuard struct {
		command   string
		stmt      string
		startTime time.Duration
	}
)

func newGuard(command string) sqlGuard {
	if logStmtSql.True() || logSlowSql.True() {
		return &realSqlGuard{
			command: command,
		}
	}

	return &nilGuard{
		command: command,
	}
}

func (n *nilGuard) start(_ string, _ ...any) error {
	n.startTime = timex.Now()
	return nil
}

func (n *nilGuard) finish(_ context.Context, _ error) {
	duration := timex.Since(n.startTime)
	metricReqDur.Observe(duration.Milliseconds(), n.command)
}

func (e *realSqlGuard) start(q string, args ...any) error {
	stmt, err := format(q, args...)
	if err != nil {
		return err
	}

	e.stmt = stmt
	e.startTime = timex.Now()

	return nil
}

func (e *realSqlGuard) finish(ctx context.Context, err error) {
	duration := timex.Since(e.startTime)
	logger := logx.WithContext(ctx).WithDuration(duration)
	if e.slowLog(ctx, duration) {
		logger.Slowf("[SQL] %s: slowcall - %s", e.command, e.stmt)
	} else if e.statementLog(ctx) {
		logger.Infof("sql %s: %s", e.command, e.stmt)
	}

	if err != nil {
		logSqlError(ctx, e.stmt, err)
	}

	metricReqDur.Observe(duration.Milliseconds(), e.command)
}

func (e *realSqlGuard) slowLog(ctx context.Context, duration time.Duration) bool {
	sqlLogOpt, ok := sqlLogOptionFromContext(ctx)
	if ok {
		return duration > slowThreshold.Load() && sqlLogOpt.EnableSlow
	}

	return duration > slowThreshold.Load() && logSlowSql.True()
}

func (e *realSqlGuard) statementLog(ctx context.Context) bool {
	sqlLogOpt, ok := sqlLogOptionFromContext(ctx)
	if ok {
		return sqlLogOpt.EnableStatement
	}

	return logStmtSql.True()
}

var emptySqlLogOption = logOption{}

type (
	logOptionKey struct{}
)

func newLogOptionContext(ctx context.Context, logOpt logOption) context.Context {
	return context.WithValue(ctx, logOptionKey{}, logOpt)
}

func sqlLogOptionFromContext(ctx context.Context) (logOption, bool) {
	value := ctx.Value(logOptionKey{})
	if value == nil {
		return emptySqlLogOption, false
	}

	return value.(logOption), true
}
