package sx

import (
	"context"
	"database/sql"
)

// Tx extends sql.Tx with some Must*** methods that panic instead of returning an error code.  Tx objects are used
// inside of transactions managed by Do.  Panics are caught by Do and returned as errors.
type Tx struct {
	*sql.Tx
}

// An sxError is used to wrap errors that we want to send back to the caller of Do.
type sxError struct {
	err error
}

// MustExec executes a query without returning any rows.  The args are for any placeholder parameters in the query.
// In case of error, the transaction is aborted and Do returns the error code.
func (tx *Tx) MustExec(query string, args ...interface{}) sql.Result {
	return tx.MustExecContext(context.Background(), query, args...)
}

// MustExecContext executes a query without returning any rows.  The args are for any placeholder parameters in the
// query.  In case of error, the transaction is aborted and Do returns the error code.
func (tx *Tx) MustExecContext(ctx context.Context, query string, args ...interface{}) sql.Result {
	res, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		panic(sxError{err})
	}
	return res
}

// MustQuery executes a query that returns rows.  The args are for any placeholder parameters in the query.
// In case of error, the transaction is aborted and Do returns the error code.
func (tx *Tx) MustQuery(query string, args ...interface{}) *Rows {
	return tx.MustQueryContext(context.Background(), query, args...)
}

// MustQueryContext executes a query that returns rows.  The args are for any placeholder parameters in the query.
// In case of error, the transaction is aborted and Do returns the error code.
func (tx *Tx) MustQueryContext(ctx context.Context, query string, args ...interface{}) *Rows {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		panic(sxError{err})
	}
	return &Rows{rows}
}

// MustQueryRow executes a query that is expected to return at most one row.  MustQueryRow always returns a non-nil
// value.  Errors are deferred until one of the Row's scan methods is called.
func (tx *Tx) MustQueryRow(query string, args ...interface{}) *Row {
	return &Row{tx.QueryRowContext(context.Background(), query, args...)}
}

// MustQueryRowContext executes a query that is expected to return at most one row.  MustQueryRow always returns a
// non-nil value.  Errors are deferred until one of the Row's scan methods is called.
func (tx *Tx) MustQueryRowContext(ctx context.Context, query string, args ...interface{}) *Row {
	return &Row{tx.QueryRowContext(ctx, query, args...)}
}

// MustPrepare creates a prepared statement for later queries or executions.  Multiple queries or executions may be
// run concurrently from the returned statement.  In case of error, the transaction is aborted and Do returns the
// error code.
//
// The caller must call the statement's Close method when the statement is no longer needed.
func (tx *Tx) MustPrepare(query string) *Stmt {
	return tx.MustPrepareContext(context.Background(), query)
}

// MustPrepareContext creates a prepared statement for later queries or executions.  Multiple queries or executions
// may be run concurrently from the returned statement.  In case of error, the transaction is aborted and Do returns
// the error code.
//
// The caller must call the statement's Close method when the statement is no longer needed.
func (tx *Tx) MustPrepareContext(ctx context.Context, query string) *Stmt {
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		panic(sxError{err})
	}
	return &Stmt{stmt}
}

// Fail aborts and rolls back the transaction, returning the given error code to the caller of Do.  Fail always
// rolls back the transaction, even if err is nil.
func (tx *Tx) Fail(err error) {
	panic(sxError{err})
}

// Stmt extends sql.Stmt with some Must*** methods that panic instead of returning an error code.  Stmt objects are
// used inside of transactions managed by Do.  Panics are caught by Do and returned as errors.
type Stmt struct {
	*sql.Stmt
}

// MustExec executes a prepared statement with the given arguments and returns an sql.Result summarizing the effect
// of the statement.  In case of error, the transaction is aborted and Do returns the error code.
func (stmt *Stmt) MustExec(args ...interface{}) sql.Result {
	return stmt.MustExecContext(context.Background(), args...)
}

// MustExecContext executes a prepared statement with the given arguments and returns an sql.Result summarizing the
// effect of the statement.  In case of error, the transaction is aborted and Do returns the error code.
func (stmt *Stmt) MustExecContext(ctx context.Context, args ...interface{}) sql.Result {
	res, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		panic(sxError{err})
	}
	return res
}

// MustQuery executes a prepared query statement with the given arguments and returns the query results as a *Rows.
// In case of error, the transaction is aborted and Do returns the error code.
func (stmt *Stmt) MustQuery(args ...interface{}) *Rows {
	return stmt.MustQueryContext(context.Background(), args...)
}

// MustQueryContext executes a prepared query statement with the given arguments and returns the query results as
// a *Rows.  In case of error, the transaction is aborted and Do returns the error code.
func (stmt *Stmt) MustQueryContext(ctx context.Context, args ...interface{}) *Rows {
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		panic(sxError{err})
	}
	return &Rows{rows}
}

// MustQueryRow executes a prepared query that is expected to return at most one row.  MustQueryRow always returns
// a non-nil value.  Errors are deferred until one of the Row's scan methods is called.
func (stmt *Stmt) MustQueryRow(args ...interface{}) *Row {
	return &Row{stmt.QueryRowContext(context.Background(), args...)}
}

// MustQueryRowContext executes a prepared query that is expected to return at most one row.  MustQueryRowContext
// always returns a non-nil value.  Errors are deferred until one of the Row's scan methods is called.
func (stmt *Stmt) MustQueryRowContext(ctx context.Context, args ...interface{}) *Row {
	return &Row{stmt.QueryRowContext(ctx, args...)}
}

// Do runs a callback function f, providing f with the prepared statement, and then closing the prepared statement
// after f returns.
func (stmt *Stmt) Do(f func(*Stmt)) {
	defer stmt.Close()
	f(stmt)
}

// Row is the result of calling MustQueryRow to select a single row.  Row extends sql.Row with some useful
// scan methods.
type Row struct {
	*sql.Row
}

// MustScan copies the columns in the current row into the values pointed at by dest.  In case of error, the
// transaction is aborted and Do returns the error code.
func (row *Row) MustScan(dest ...interface{}) {
	err := row.Scan(dest...)
	if err != nil {
		panic(sxError{err})
	}
}

// MustScans copies the columns in the current row into the struct pointed at by dest.  In case of error, the
// transaction is aborted and Do returns the error code.
func (row *Row) MustScans(dest interface{}) {
	row.MustScan(Addrs(dest)...)
}

// Rows is the result of calling MustQuery to select a set of rows.  Rows extends sql.Rows with some useful
// scan methods.
type Rows struct {
	*sql.Rows
}

// MustScan calls Scan to read in a row of the result set.  In case of error, the transaction is aborted and Do
// returns the error code.
func (rows *Rows) MustScan(dest ...interface{}) {
	err := rows.Scan(dest...)
	if err != nil {
		panic(sxError{err})
	}
}

// MustScans copies the columns in the current row into the struct pointed at by dest.  In case of error, the
// transaction is aborted and Do returns the error code.
func (rows *Rows) MustScans(dest interface{}) {
	rows.MustScan(Addrs(dest)...)
}

// Each iterates over all of the rows in a result set and runs a callback function on each row.
func (rows *Rows) Each(f func(*Rows)) {
	defer rows.Close()
	for rows.Next() {
		f(rows)
	}
	err := rows.Err()
	if err != nil {
		panic(sxError{err})
	}
}

// Do runs the function f in a transaction.  Within f, if Fail() is invoked or if any Must*** method encounters
// an error, then the transaction is rolled back and Do returns the error.  If f runs to completion, then the
// transaction is committed, and Do returns nil.
//
// Internally, the Must*** methods panic on error, and Fail() always panics.  The panic aborts execution of f.
// f should not attempt to recover from the panic.  Instead, Do will catch the panic and return it as an error.
//
// A TxOptions may be provided to specify isolation level and/or read-only status.  If no TxOptions is provided,
// then the default oprtions are used.  Extra TxOptions are ignored.
func Do(db *sql.DB, f func(*Tx), opts ...sql.TxOptions) error {
	return DoContext(context.Background(), db, f, opts...)
}

// DoContext runs the function f in a transaction.  Within f, if Fail() is invoked or if any Must*** method encounters
// an error, then the transaction is rolled back and Do returns the error.  If f runs to completion, then the
// transaction is committed, and DoContext returns nil.
//
// Internally, the Must*** methods panic on error, and Fail() always panics.  The panic aborts execution of f.
// f should not attempt to recover from the panic.  Instead, Do will catch the panic and return it as an error.
//
// A TxOptions may be provided to specify isolation level and/or read-only status.  If no TxOptions is provided,
// then the default oprtions are used.  Extra TxOptions are ignored.
func DoContext(ctx context.Context, db *sql.DB, f func(*Tx), opts ...sql.TxOptions) (err error) {

	var opt *sql.TxOptions
	if len(opts) > 0 {
		opt = &opts[0]
	}

	var tx *sql.Tx
	tx, err = db.BeginTx(ctx, opt)
	if err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			if ourerr, ok := r.(sxError); ok {
				// Our panic.  Unwrap it and return it as an error code.
				tx.Rollback()
				err = ourerr.err
			} else {
				// Not our panic, so propagate it.
				panic(r)
			}
		}
	}()

	// This runs the queries.
	f(&Tx{tx})

	err = tx.Commit()
	return
}
