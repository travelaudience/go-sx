# Some simple SQL extensions for Go

[![GoDoc](https://godoc.org/github.com/travelaudience/go-sx?status.svg)](http://godoc.org/github.com/travelaudience/go-sx)
[![CircleCI](https://circleci.com/gh/travelaudience/go-sx.svg?style=svg)](https://circleci.com/gh/travelaudience/go-sx)

**go-sx** provides some extensions to the standard library `database/sql` package.  It is designed for those who wish to use the full power of SQL without a heavy abstraction layer.

**UPDATE (July 2020):** This library is still actively maintained.  Contributions are welcome.

## Goals

The primary goal of **go-sx** is to eliminate boilerplate code.  Specifically, **go-sx** attempts to address the following pain points:

1. Transactions are clumsy.  It would be nice to have a simple function to run a callback in a transaction.
2. Error handling is clumsy.  It would be nice to have errors within a transaction automatically exit the transaction and trigger a rollback.  (This is nearly always what we want to do.)
3. Scanning multiple columns is clumsy.  It would be nice to have a simple way to scan into multiple struct fields at once.
4. Constructing queries is clumsy, especially when there are a lot of columns.
5. Iterating over result sets is clumsy.

## Non-goals

These are considered to be out of scope:

1. Be an ORM.
2. Write your queries for you.
3. Suggest that we need 1:1 relationship between struct types and tables.
4. Maintain database schemas.
5. Abstract away differences between SQL dialects.
6. Automatic type-manipulation.
7. Magic.

## Pain point #1:  Transactions are clumsy.

**go-sx** provides a function `Do` to run a transaction in a callback, automatically committing on success or rolling back on failure.

Here is some simple code to run two queries in a transaction.  The second query returns two values, which are read into variables `x` and `y`.

```go
tx, err := db.Begin()
if err != nil {
    return err
}
if _, err := tx.Exec(query0); err != nil {
    tx.Rollback()
    return err
}
if err := tx.QueryRow(query1).Scan(&x, &y); err != nil {
    tx.Rollback()
    return err
}
if err := tx.Commit(); err != nil {
    return err
}
```

Using the `Do` function, we put the business logic into a callback function and have **go-sx** take care of the transaction logic.

The `sx.Tx` object provided to the callback is the `sql.Tx` transaction object, extended with a few methods.  If we call `tx.Fail()`, then the transaction is immediately aborted and rolled back.

```go
err := sx.Do(db, func (tx *sx.Tx) {
    if _, err := tx.Exec(query0); err != nil {
        tx.Fail(err)
    }
    if err := tx.QueryRow(query1).Scan(&x, &y); err != nil {
        tx.Fail(err)
    }
})
```

Under the hood, `tx.Fail()` generates a panic which is recovered by `Do`.

## Pain point #2: Error handling is clumsy.

**go-sx** provides a collection of `Must***` methods which may be used inside of the callback to `Do`.  Any error encountered while in a `Must***` method causes the transaction to be aborted and rolled back.

Here is the code above, rewritten to use `Do`'s error handling.  It's simple and readable.

```go
err := sx.Do(db, func (tx *sx.Tx) {
    tx.MustExec(query0)
    tx.MustQueryRow(query1).MustScan(&x, &y)
})
```

## Pain point #3:  Scanning multiple columns is clumsy.

**go-sx** provides an `Addrs` function, which takes a struct and returns a slice of pointers to the elements.  So instead of:

```go
row.Scan(&a.Width, &a.Height, &a.Depth)
```

We can write:

```go
row.Scan(sx.Addrs(&a)...)
```

Or better yet, let **go-sx** handle the errors:

```go
row.MustScan(sx.Addrs(&a)...)
```

This is such a common pattern that we provide a shortcut to do this all in one step:

```go
row.MustScans(&a)
```

## Pain point #4:  Constructing queries is clumsy.

We would like **go-sx** to be able to construct some common queries for us.  To this end, we define a simple way to match struct fields with database columns, and then provide some helper functions that use this matching to construct queries.

By default, all exported struct fields match database columns whose name is the the field name snake_cased.  The default can be overridden by explicitly tagging fields, much like what is done with the standard json encoder.  Note that we don't care about the name of the table at this point.

Here is a struct that can be used to scan columns `violin`, `viola`, `cello` and `contrabass`.

```go
type orchestra struct {
    Violin string
    Viola  string
    Cello  string
    Bass   string `sx:"contrabass"`
}
```

We can use the helper function `SelectQuery` to construct a simple query.  Then we can add the WHERE clause that we need and scan the result set into our struct.

```go
var spo orchestra

wantID := 123
query := sx.SelectQuery("symphony", &spo) + " WHERE id=?"  // SELECT violin,viola,cello,contrabass FROM symphony WHERE id=?
tx.MustQueryRow(query, wantID).MustScans(&spo)
```

Note that a struct need not follow the database schema exactly.  It's entirely possible to have various structs mapped to different columns of the same table, or even one struct that maps to a query on joined tables.  On the other hand, it's essential that the columns in the query match the fields of the struct, and **go-sx** guarantees this, as we'll see below.

In some cases it's useful to have a struct that is used for both selects and inserts, with some of the fields being used just for selects.  This can be accomplished with the "readonly" tag.

```go
type orchestra1 struct {
    Violin string `sx:",readonly"`
    Viola  string
    Cello  string
    Bass   string `sx:"contrabass"`
}
```

It's also useful in some cases to have a struct field that is ignored by **go-sx**.  This can be accomplished with the "-" tag.

```go
type orchestra2 struct {
    Violin string `sx:",readonly"`
    Viola  string `sx:"-"`
    Cello  string
    Bass   string `sx:"contrabass"`
}
```

We can construct insert queries in a similar manner.  Violin is read-only and Viola is ignored, so we only need to provide values for Cello and Bass.  (If you need Postgres-style `$n` placeholders, see `sx.SetNumberedPlaceholders()`.)

```go
spo := orchestra2{Cello: "Strad", Bass: "Cecilio"}

query := sx.InsertQuery("symphony", &spo)  // INSERT INTO symphony (cello,contrabass) VALUES (?,?)
tx.MustExec(query, sx.Values(&spo)...)
```

We can contruct update queries this way too, and there is also an option to skip fields whose values are the zero values.  (The update structs support pointer fields, making this skip option rather useful.)

```go
spoChanges := orchestra2{Bass: "Strad"}

wantID := 123
query, values := sx.UpdateQuery("symphony", &spoChanges) + " WHERE id=?"  // UPDATE symphony SET contrabass=? WHERE id=?
tx.MustExec(query, append(values, wantID)...)
```

It is entirely possible to construct all of these queries by hand, and you're all welcome to do so.  Using the query generators, however, ensures that the fields match correctly, something that is particularly useful with a large number of columns.

## Pain point #5:  Iterating over result sets is clumsy.

**go-sx** provides an iterator called `Each` which runs a callback function on each row of a result set.  Using the iterator simplifies this code:

```go
var orchestras []orchestra

query := "SELECT violin,viola,cello,contrabass FROM symphony ORDER BY viola"  // Or we could use sx.SelectQuery()
rows := tx.MustQuery(query)
defer rows.Close()
for rows.Next() {
    var o orchestra
    rows.MustScans(&o)
    orchestras = append(orchestras, o)
}
if err := rows.Err(); err != nil {
    tx.Fail(err)
}
```

To this:

```go
var orchestras []orchestra

query := "SELECT violin,viola,cello,contrabass FROM symphony ORDER BY viola"
tx.MustQuery(query).Each(func (r *sx.Rows) {
    var o orchestra
    r.MustScans(&o)
    orchestras = append(orchestras, o)
})
```

## Contributing

Contributions are welcome! Read the [Contributing Guide](CONTRIBUTING.md) for more information.

## Licensing

This project is licensed under the MIT License - see the [LICENSE](LICENSE.txt) file for details
