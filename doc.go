// Package sx provides some simple extensions to the database/sql package to reduce the amount of boilerplate code.
//
// Transactions and error handling
//
// Package sx provides a function called Do, which runs a callback function inside a transaction.  The callback
// function is provided with a Tx object, which is an sql.Tx object that has been extended with some Must*** methods.
// When a Must*** method encounters an error, it panics, and the panic is caught by Do and returned to the caller
// as an error value.
//
// Do automatically commits or rolls back the transaction based on whether or not the callback function completed
// successfuly.
//
// Query helpers and struct matching
//
// Package sx provides functions to generate frequently-used queries, based on a simple matching between struct
// fields and database columns.
//
// By default, every field in a struct corresponds to the database column whose name is the snake-cased version of
// the field name, i.e. the field HelloWorld corresponds to the "hello_world" column.  Acronyms are treated as words,
// so HelloRPCWorld becomes "hello_rpc_world".
//
// The column name can also be specified explicitly by tagging the field with the desired name, and fields can be
// excluded altogether by tagging with "-".
//
// Fields that should be used for scanning but exluded for inserts and updates are additionally tagged "readonly".
//
// Examples:
//
//     // Field is called "field" in the database.
//     Field int
//
//     // Field is called "hage" in the database.
//     Field int `sx:"hage"`
//
//     // Field is called "hage" in the database and should be skipped for inserts and updates.
//     Field int `sx:"hage,readonly"`
//
//     // Field is called "field" in the database and should be skipped for inserts and updates.
//     Field int `sx:",readonly"`
//
//     // Field should be ignored by sx.
//     Field int `sx:"-"`
package sx
