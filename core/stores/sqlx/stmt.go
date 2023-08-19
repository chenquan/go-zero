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
	guard := newGuard(ctx, "exec")
	if err := guard.start(q, args...); err != nil {
		return nil, err
	}

	result, err := conn.ExecContext(ctx, q, args...)
	guard.finish(ctx, err)

	return result, err
}

func execStmt(ctx context.Context, conn stmtConn, q string, args ...any) (sql.Result, error) {
	guard := newGuard(ctx, "execStmt")
	if err := guard.start(q, args...); err != nil {
		return nil, err
	}

	result, err := conn.ExecContext(ctx, args...)
	guard.finish(ctx, err)

	return result, err
}

func query(ctx context.Context, conn sessionConn, scanner func(*sql.Rows) error, q string, args ...any) error {
	guard := newGuard(ctx, "query")
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
	guard := newGuard(ctx, "queryStmt")
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
	logOptSqlGuard struct {
		command         string
		stmt            string
		startTime       time.Duration
		enableStatement bool
		enableSlow      bool
	}
)

func newGuard(ctx context.Context, command string) sqlGuard {
	logOpt, _ := sqlLogOptionFromContext(ctx)

	logStmt := logStmtSql.True()
	if logOpt.EnableStatement != nil {
		logStmt = *logOpt.EnableStatement
	}

	logSlow := logSlowSql.True()
	if logOpt.EnableSlow != nil {
		logSlow = *logOpt.EnableSlow
	}

	if logSlow || logStmt {
		return &logOptSqlGuard{
			command:         command,
			enableSlow:      logSlow,
			enableStatement: logStmt,
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
	if e.slowLog(duration) {
		logger.Slowf("[SQL] %s: slowcall - %s", e.command, e.stmt)
	} else if e.statementLog() {
		logger.Infof("sql %s: %s", e.command, e.stmt)
	}

	if err != nil {
		logSqlError(ctx, e.stmt, err)
	}

	metricReqDur.Observe(duration.Milliseconds(), e.command)
}

func (e *realSqlGuard) slowLog(duration time.Duration) bool {
	return duration > slowThreshold.Load() && logSlowSql.True()
}

func (e *realSqlGuard) statementLog() bool {
	return logStmtSql.True()
}

func (l *logOptSqlGuard) start(q string, args ...any) error {
	stmt, err := format(q, args...)
	if err != nil {
		return err
	}

	l.stmt = stmt
	l.startTime = timex.Now()

	return nil
}

func (l *logOptSqlGuard) finish(ctx context.Context, err error) {
	duration := timex.Since(l.startTime)
	logger := logx.WithContext(ctx).WithDuration(duration)
	if l.slowLog(duration) {
		logger.Slowf("[SQL] %s: slowcall - %s", l.command, l.stmt)
	} else if l.statementLog() {
		logger.Infof("sql %s: %s", l.command, l.stmt)
	}

	if err != nil {
		logSqlError(ctx, l.stmt, err)
	}

	metricReqDur.Observe(duration.Milliseconds(), l.command)
}

func (l *logOptSqlGuard) slowLog(duration time.Duration) bool {
	return duration > slowThreshold.Load() && l.enableSlow
}

func (l *logOptSqlGuard) statementLog() bool {
	return l.enableStatement
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
