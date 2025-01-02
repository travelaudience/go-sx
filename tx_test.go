package sx_test

import (
	"context"
	"database/sql"
	"errors"
	"math/rand"
	"os"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	sx "github.com/travelaudience/go-sx"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

// helper functions

func newMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("error creating mock database: %v", err)
	}
	return db, mock
}

func endMock(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	err := mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("mocked expectations were not met: %v", err)
	}
}

func TestMustExec(t *testing.T) {

	t.Run("MustExec with result", func(t *testing.T) {
		db, mock := newMock(t)
		a, b := rand.Int63(), rand.Int63()
		const query = "SELECT alpha"

		mock.ExpectBegin()
		mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(a, b))
		mock.ExpectCommit()

		err := sx.Do(db, func(tx *sx.Tx) {
			res := tx.MustExec(query)
			a0, _ := res.LastInsertId()
			b0, _ := res.RowsAffected()
			if a0 != a || b0 != b {
				t.Errorf("Expected result (%d, %d), got (%d, %d)", a, b, a0, b0)
			}
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		endMock(t, mock)
	})

	t.Run("MustExec with error", func(t *testing.T) {
		db, mock := newMock(t)
		const query = "SELECT bravo"
		err0 := errors.New("bravo error")

		mock.ExpectBegin()
		mock.ExpectExec(query).WillReturnError(err0)
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustExec(query)
		})
		if err != err0 {
			t.Errorf("expected error %v, got %v", err0, err)
		}

		endMock(t, mock)
	})

	t.Run("MustExec with 1 argument and result", func(t *testing.T) {
		db, mock := newMock(t)
		x, a, b := rand.Int63(), rand.Int63(), rand.Int63()
		const query = "SELECT charlie"

		mock.ExpectBegin()
		mock.ExpectExec(query).WithArgs(x).WillReturnResult(sqlmock.NewResult(a, b))
		mock.ExpectCommit()

		err := sx.Do(db, func(tx *sx.Tx) {
			res := tx.MustExec(query, x)
			a0, _ := res.LastInsertId()
			b0, _ := res.RowsAffected()
			if a0 != a || b0 != b {
				t.Errorf("Expected result (%d, %d), got (%d, %d)", a, b, a0, b0)
			}
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		endMock(t, mock)
	})

	t.Run("MustExec with 2 arguments and error", func(t *testing.T) {
		db, mock := newMock(t)
		x, y := rand.Int63(), rand.Int63()
		const query = "SELECT delta"
		err0 := errors.New("delta error")

		mock.ExpectBegin()
		mock.ExpectExec(query).WithArgs(x, y).WillReturnError(err0)
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustExec(query, x, y)
		})
		if err != err0 {
			t.Errorf("expected error %v, got %v", err0, err)
		}

		endMock(t, mock)
	})

	t.Run("MustExecContext with result", func(t *testing.T) {
		db, mock := newMock(t)
		a, b := rand.Int63(), rand.Int63()
		const query = "SELECT alpha_context"

		mock.ExpectBegin()
		mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(a, b))
		mock.ExpectCommit()

		err := sx.Do(db, func(tx *sx.Tx) {
			res := tx.MustExecContext(context.Background(), query)
			a0, _ := res.LastInsertId()
			b0, _ := res.RowsAffected()
			if a0 != a || b0 != b {
				t.Errorf("Expected result (%d, %d), got (%d, %d)", a, b, a0, b0)
			}
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		endMock(t, mock)
	})

	t.Run("MustExec with isolation level and error", func(t *testing.T) {
		db, mock := newMock(t)
		const query = "SELECT bravissimo"
		err0 := errors.New("bravissimo error")

		mock.ExpectBegin()
		mock.ExpectExec(query).WillReturnError(err0)
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustExec(query)
		}, sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != err0 {
			t.Errorf("expected error %v, got %v", err0, err)
		}

		endMock(t, mock)
	})
}

func TestMustQueryRow(t *testing.T) {

	t.Run("MustQueryRow with result", func(t *testing.T) {
		db, mock := newMock(t)
		a, b := rand.Int63(), rand.Int63()
		const query = "SELECT echo"
		rows := sqlmock.NewRows([]string{"a", "b"}).AddRow(a, b)

		mock.ExpectBegin()
		mock.ExpectQuery(query).WillReturnRows(rows)
		mock.ExpectCommit()

		err := sx.Do(db, func(tx *sx.Tx) {
			var a0, b0 int64
			tx.MustQueryRow(query).MustScan(&a0, &b0)
			if a0 != a || b0 != b {
				t.Errorf("Expected result (%d, %d), got (%d, %d)", a, b, a0, b0)
			}
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		endMock(t, mock)
	})

	t.Run("MustQueryRow with no rows", func(t *testing.T) {
		db, mock := newMock(t)
		const query = "SELECT foxtrot"
		rows := sqlmock.NewRows([]string{"a", "b"})

		mock.ExpectBegin()
		mock.ExpectQuery(query).WillReturnRows(rows)
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			var a0, b0 int64
			tx.MustQueryRow(query).MustScan(&a0, &b0)
		})
		if err != sql.ErrNoRows {
			t.Errorf("expected error %v, got %v", sql.ErrNoRows, err)
		}

		endMock(t, mock)
	})

	t.Run("MustQueryRow with error", func(t *testing.T) {
		db, mock := newMock(t)
		const query = "SELECT golf"
		err0 := errors.New("golf error")

		mock.ExpectBegin()
		mock.ExpectQuery(query).WillReturnError(err0)
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			var a0, b0 int64
			tx.MustQueryRow(query).MustScan(&a0, &b0)
		})
		if err != err0 {
			t.Errorf("expected error %v, got %v", err0, err)
		}

		endMock(t, mock)
	})

	t.Run("MustQueryRow with 1 argument and error", func(t *testing.T) {
		db, mock := newMock(t)
		x := rand.Int63()
		const query = "SELECT hotel"
		err0 := errors.New("hotel error")

		mock.ExpectBegin()
		mock.ExpectQuery(query).WithArgs(x).WillReturnError(err0)
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			var a0, b0 int64
			tx.MustQueryRow(query, x).MustScan(&a0, &b0)
		})
		if err != err0 {
			t.Errorf("expected error %v, got %v", err0, err)
		}

		endMock(t, mock)
	})

	t.Run("MustQueryRow with 3 arguments and struct result", func(t *testing.T) {
		db, mock := newMock(t)
		a, b, x, y, z := rand.Int63(), rand.Int63(), rand.Int63(), rand.Int63(), rand.Int63()
		const query = "SELECT indigo"
		rows := sqlmock.NewRows([]string{"a", "b"}).AddRow(a, b)

		mock.ExpectBegin()
		mock.ExpectQuery(query).WithArgs(x, y, z).WillReturnRows(rows)
		mock.ExpectCommit()

		err := sx.Do(db, func(tx *sx.Tx) {
			var res struct{ A, B int64 }
			tx.MustQueryRow(query, x, y, z).MustScans(&res)
			if res.A != a || res.B != b {
				t.Errorf("Expected result (%d, %d), got (%d, %d)", a, b, res.A, res.B)
			}
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		endMock(t, mock)
	})

	t.Run("MustQueryRowContext with result", func(t *testing.T) {
		db, mock := newMock(t)
		a, b := rand.Int63(), rand.Int63()
		const query = "SELECT echo_context"
		rows := sqlmock.NewRows([]string{"a", "b"}).AddRow(a, b)

		mock.ExpectBegin()
		mock.ExpectQuery(query).WillReturnRows(rows)
		mock.ExpectCommit()

		err := sx.Do(db, func(tx *sx.Tx) {
			var a0, b0 int64
			tx.MustQueryRowContext(context.TODO(), query).MustScan(&a0, &b0)
			if a0 != a || b0 != b {
				t.Errorf("Expected result (%d, %d), got (%d, %d)", a, b, a0, b0)
			}
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		endMock(t, mock)
	})
}

func TestMustQuery(t *testing.T) {

	t.Run("MustQuery with error", func(t *testing.T) {
		db, mock := newMock(t)
		const query = "SELECT juliett"
		err0 := errors.New("juliett error")

		mock.ExpectBegin()
		mock.ExpectQuery(query).WillReturnError(err0)
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustQuery(query)
		})
		if err != err0 {
			t.Errorf("expected error %v, got %v", err0, err)
		}

		endMock(t, mock)
	})

	t.Run("MustQuery with 1 argument and error", func(t *testing.T) {
		db, mock := newMock(t)
		x := rand.Int63()
		const query = "SELECT kilo"
		err0 := errors.New("kilo error")

		mock.ExpectBegin()
		mock.ExpectQuery(query).WithArgs(x).WillReturnError(err0)
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustQuery(query, x)
		})
		if err != err0 {
			t.Errorf("expected error %v, got %v", err0, err)
		}

		endMock(t, mock)
	})

	t.Run("MustQuery with 1 argument and 1 result row", func(t *testing.T) {
		db, mock := newMock(t)
		a, b, x := rand.Int63(), rand.Int63(), rand.Int63()
		const query = "SELECT lima"
		rows := sqlmock.NewRows([]string{"a", "b"}).AddRow(a, b)

		mock.ExpectBegin()
		mock.ExpectQuery(query).WithArgs(x).WillReturnRows(rows)
		mock.ExpectCommit()

		var a0, b0 int64
		n := 0
		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustQuery(query, x).Each(func(r *sx.Rows) {
				r.MustScan(&a0, &b0)
				n++
			})
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if n != 1 {
			t.Errorf("Expected 1 row, got %d", n)
		} else if a0 != a || b0 != b {
			t.Errorf("Expected result (%d, %d), got (%d, %d)", a, b, a0, b0)
		}

		endMock(t, mock)
	})

	t.Run("MustQuery with 1 argument and 1 result row with error", func(t *testing.T) {
		db, mock := newMock(t)
		x := rand.Int63()
		const query = "SELECT mike"
		rows := sqlmock.NewRows([]string{"a", "b"}).AddRow("scan", "error")

		mock.ExpectBegin()
		mock.ExpectQuery(query).WithArgs(x).WillReturnRows(rows)
		mock.ExpectRollback()

		var a0, b0 int64
		n := 0
		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustQuery(query, x).Each(func(r *sx.Rows) {
				r.MustScan(&a0, &b0)
				n++
			})
		})
		if n != 0 {
			t.Errorf("Expected no rows, got %d", n)
		} else if err == nil || !strings.Contains(err.Error(), "Scan error") {
			t.Errorf("unexpected error: %v", err)
		}

		endMock(t, mock)
	})

	t.Run("MustQuery with 1 argument and 2 struct result rows", func(t *testing.T) {
		type ab struct{ A, B int64 }

		db, mock := newMock(t)
		dat, x := [2]ab{{A: rand.Int63(), B: rand.Int63()}, {A: rand.Int63(), B: rand.Int63()}}, rand.Int63()
		const query = "SELECT november"
		rows := sqlmock.NewRows([]string{"a", "b"}).AddRow(dat[0].A, dat[0].B).AddRow(dat[1].A, dat[1].B)

		mock.ExpectBegin()
		mock.ExpectQuery(query).WithArgs(x).WillReturnRows(rows)
		mock.ExpectCommit()

		var res [2]ab
		n := 0
		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustQuery(query, x).Each(func(r *sx.Rows) {
				r.MustScans(&res[n])
				n++
			})
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if n != 2 {
			t.Errorf("Expected 2 rows, got %d", n)
		} else if res != dat {
			t.Errorf("Expected results (%d, %d), (%d, %d) got (%d, %d), (%d, %d)",
				dat[0].A, dat[0].B, dat[1].A, dat[1].B, res[0].A, res[0].B, res[1].A, res[1].B)
		}

		endMock(t, mock)
	})

	t.Run("MustQuery with 2 arguments, 2 result rows and row error", func(t *testing.T) {
		db, mock := newMock(t)
		a, b, x, y := [2]int64{rand.Int63(), 0}, [2]int64{rand.Int63(), 0}, rand.Int63(), rand.Int63()
		const query = "SELECT oscar"
		err0 := errors.New("oscar error")
		rows := sqlmock.NewRows([]string{"a", "b"}).AddRow(a[0], b[0]).AddRow(a[1], b[1]).RowError(1, err0)

		mock.ExpectBegin()
		mock.ExpectQuery(query).WithArgs(x, y).WillReturnRows(rows)
		mock.ExpectRollback()

		var aa, bb [2]int64
		n := 0
		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustQuery(query, x, y).Each(func(r *sx.Rows) {
				r.MustScan(&aa[n], &bb[n])
				n++
			})
		})
		if n != 1 {
			t.Errorf("Expected 1 row before the row error, got %d", n)
		} else if err != err0 {
			t.Errorf("unexpected error: %v", err)
		} else if aa != a || bb != b {
			t.Errorf("Expected result (%d, %d) before the row error, got (%d, %d)",
				a[0], b[0], aa[0], bb[0])
		}

		endMock(t, mock)
	})

	t.Run("MustQuery with 2 arguments, 2 result rows and scan error", func(t *testing.T) {
		db, mock := newMock(t)
		a, b := [3]int64{rand.Int63(), rand.Int63(), 0}, [3]int64{rand.Int63(), rand.Int63(), 0}
		x, y := rand.Int63(), rand.Int63()
		const query = "SELECT papa"
		rows := sqlmock.NewRows([]string{"a", "b"}).AddRow(a[0], b[0]).AddRow(a[1], b[1]).AddRow("scan", "error")

		mock.ExpectBegin()
		mock.ExpectQuery(query).WithArgs(x, y).WillReturnRows(rows)
		mock.ExpectRollback()

		var aa, bb [3]int64
		n := 0
		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustQuery(query, x, y).Each(func(r *sx.Rows) {
				r.MustScan(&aa[n], &bb[n])
				n++
			})
		})
		if n != 2 {
			t.Errorf("Expected 2 rows before the scan error, got %d", n)
		} else if err == nil || !strings.Contains(err.Error(), "Scan error") {
			t.Errorf("unexpected error: %v", err)
		} else if aa != a || bb != b {
			t.Errorf("Expected results (%d, %d), (%d, %d) before the scan error, got (%d, %d), (%d, %d)",
				a[0], b[0], a[1], b[1], aa[0], bb[0], aa[1], bb[1])
		}

		endMock(t, mock)
	})

	t.Run("MustQueryContext with error", func(t *testing.T) {
		db, mock := newMock(t)
		const query = "SELECT juliett_context"
		err0 := errors.New("juliett_context error")

		mock.ExpectBegin()
		mock.ExpectQuery(query).WillReturnError(err0)
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustQueryContext(context.TODO(), query)
		})
		if err != err0 {
			t.Errorf("expected error %v, got %v", err0, err)
		}

		endMock(t, mock)
	})
}

func TestMustPrepare(t *testing.T) {

	t.Run("MustPrepare with MustExec and result", func(t *testing.T) {
		db, mock := newMock(t)
		a, b, x := rand.Int63(), rand.Int63(), rand.Int63()
		const query = "SELECT quebec"

		mock.ExpectBegin()
		mock.ExpectPrepare(query).ExpectExec().WithArgs(x).WillReturnResult(sqlmock.NewResult(a, b))
		mock.ExpectCommit()

		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustPrepare(query).Do(func(stmt *sx.Stmt) {
				res := stmt.MustExec(x)
				a0, _ := res.LastInsertId()
				b0, _ := res.RowsAffected()
				if a0 != a || b0 != b {
					t.Errorf("Expected result (%d, %d), got (%d, %d)", a, b, a0, b0)
				}
			})
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		endMock(t, mock)
	})

	t.Run("MustPrepare with MustExec and error", func(t *testing.T) {
		db, mock := newMock(t)
		const query = "SELECT romeo"
		err0 := errors.New("romeo error")

		mock.ExpectBegin()
		mock.ExpectPrepare(query).ExpectExec().WillReturnError(err0)
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustPrepare(query).Do(func(stmt *sx.Stmt) {
				stmt.MustExec()
			})
		})
		if err != err0 {
			t.Errorf("expected error %v, got %v", err0, err)
		}

		endMock(t, mock)
	})

	t.Run("MustPrepare with MustQueryRow and result", func(t *testing.T) {
		db, mock := newMock(t)
		a, b, x := rand.Int63(), rand.Int63(), rand.Int63()
		const query = "SELECT sierra"
		rows := sqlmock.NewRows([]string{"a", "b"}).AddRow(a, b)

		mock.ExpectBegin()
		mock.ExpectPrepare(query).ExpectQuery().WithArgs(x).WillReturnRows(rows)
		mock.ExpectCommit()

		err := sx.Do(db, func(tx *sx.Tx) {
			var a0, b0 int64
			tx.MustPrepare(query).Do(func(stmt *sx.Stmt) {
				stmt.MustQueryRow(x).MustScan(&a0, &b0)
				if a0 != a || b0 != b {
					t.Errorf("Expected result (%d, %d), got (%d, %d)", a, b, a0, b0)
				}
			})
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		endMock(t, mock)
	})

	t.Run("MustPrepare with MustQueryRow and error", func(t *testing.T) {
		db, mock := newMock(t)
		const query = "SELECT tango"
		err0 := errors.New("tango error")

		mock.ExpectBegin()
		mock.ExpectPrepare(query).ExpectQuery().WillReturnError(err0)
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			var a0, b0 int64
			tx.MustPrepare(query).Do(func(stmt *sx.Stmt) {
				stmt.MustQueryRow().MustScan(&a0, &b0)
			})
		})
		if err != err0 {
			t.Errorf("expected error %v, got %v", err0, err)
		}

		endMock(t, mock)
	})

	t.Run("MustPrepare with MustQuery and result", func(t *testing.T) {
		db, mock := newMock(t)
		a, b, x := rand.Int63(), rand.Int63(), rand.Int63()
		const query = "SELECT uniform"
		rows := sqlmock.NewRows([]string{"a", "b"}).AddRow(a, b)

		mock.ExpectBegin()
		mock.ExpectPrepare(query).ExpectQuery().WithArgs(x).WillReturnRows(rows)
		mock.ExpectCommit()

		var a0, b0 int64
		n := 0
		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustPrepare(query).Do(func(stmt *sx.Stmt) {
				stmt.MustQuery(x).Each(func(r *sx.Rows) {
					r.MustScan(&a0, &b0)
					n++
				})
			})
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if n != 1 {
			t.Errorf("Expected 1 row, got %d", n)
		} else if a0 != a || b0 != b {
			t.Errorf("Expected result (%d, %d), got (%d, %d)", a, b, a0, b0)
		}

		endMock(t, mock)
	})

	t.Run("MustPrepare with MustQuery and error", func(t *testing.T) {
		db, mock := newMock(t)
		const query = "SELECT victor"
		err0 := errors.New("victor error")

		mock.ExpectBegin()
		mock.ExpectPrepare(query).ExpectQuery().WillReturnError(err0)
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustPrepare(query).Do(func(stmt *sx.Stmt) {
				stmt.MustQuery()
			})
		})
		if err != err0 {
			t.Errorf("expected error %v, got %v", err0, err)
		}

		endMock(t, mock)
	})

	t.Run("MustPrepare with error", func(t *testing.T) {
		db, mock := newMock(t)
		const query = "SELECT whiskey"
		err0 := errors.New("whiskey error")

		mock.ExpectBegin()
		mock.ExpectPrepare(query).WillReturnError(err0)
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustPrepare(query)
		})
		if err != err0 {
			t.Errorf("expected error %v, got %v", err0, err)
		}

		endMock(t, mock)
	})

	t.Run("MustPrepareContext with MustQueryRowContext and result", func(t *testing.T) {
		db, mock := newMock(t)
		a, b, x := rand.Int63(), rand.Int63(), rand.Int63()
		const query = "SELECT sierra_context"
		rows := sqlmock.NewRows([]string{"a", "b"}).AddRow(a, b)

		mock.ExpectBegin()
		mock.ExpectPrepare(query).ExpectQuery().WithArgs(x).WillReturnRows(rows)
		mock.ExpectCommit()

		err := sx.Do(db, func(tx *sx.Tx) {
			var a0, b0 int64
			tx.MustPrepareContext(context.TODO(), query).Do(func(stmt *sx.Stmt) {
				stmt.MustQueryRowContext(context.TODO(), x).MustScan(&a0, &b0)
				if a0 != a || b0 != b {
					t.Errorf("Expected result (%d, %d), got (%d, %d)", a, b, a0, b0)
				}
			})
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		endMock(t, mock)
	})

	t.Run("MustPrepareContext with MustQueryContext and error", func(t *testing.T) {
		db, mock := newMock(t)
		const query = "SELECT victor_context"
		err0 := errors.New("victor_context error")

		mock.ExpectBegin()
		mock.ExpectPrepare(query).ExpectQuery().WillReturnError(err0)
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			tx.MustPrepareContext(context.TODO(), query).Do(func(stmt *sx.Stmt) {
				stmt.MustQueryContext(context.TODO())
			})
		})
		if err != err0 {
			t.Errorf("expected error %v, got %v", err0, err)
		}

		endMock(t, mock)
	})
}

func TestFail(t *testing.T) {

	t.Run("explicit fail", func(t *testing.T) {
		db, mock := newMock(t)
		err0 := errors.New("x-ray error")

		mock.ExpectBegin()
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			tx.Fail(err0)
		})
		if err != err0 {
			t.Errorf("expected error %v, got %v", err0, err)
		}

		endMock(t, mock)
	})

	t.Run("begin transaction fail", func(t *testing.T) {
		db, mock := newMock(t)
		err0 := errors.New("yankee error")

		mock.ExpectBegin().WillReturnError(err0)

		err := sx.Do(db, func(tx *sx.Tx) {
			tx.Fail(errors.New("should never happen"))
		})

		if err != err0 {
			t.Errorf("expected error %v, got %v", err0, err)
		}

		endMock(t, mock)
	})

	t.Run("panic inside transaction", func(t *testing.T) {
		// This test ensures that an arbitrary panic inside a transaction is not erroneously caught by us and instead
		// gets propagated back up as a panic.
		db, mock := newMock(t)
		err0 := errors.New("zulu error")

		mock.ExpectBegin()

		var err error
		func() {
			defer func() {
				if r := recover(); r != nil {
					if e, ok := r.(error); ok {
						err = e
					}
				}
			}()
			sx.Do(db, func(tx *sx.Tx) {
				panic(err0)
			})
		}()
		if err != err0 {
			t.Errorf("expected panic %v, got %v", err0, err)
		}

		endMock(t, mock)
	})

	t.Run("explicit nil fail", func(t *testing.T) {
		db, mock := newMock(t)

		mock.ExpectBegin()
		mock.ExpectRollback()

		err := sx.Do(db, func(tx *sx.Tx) {
			// This should roll back the transaction and return a nil error
			tx.Fail(nil)
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		endMock(t, mock)
	})
}
