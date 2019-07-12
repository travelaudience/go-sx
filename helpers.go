package sx

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

// SelectQuery returns a query string of the form
//     SELECT <columns> FROM <table>
// where <columns> is the list of columns defined by the struct pointed at by datatype, and <table> is the table name
// given.
func SelectQuery(table string, datatype interface{}) string {
	bob := strings.Builder{}
	bob.WriteString("SELECT")
	var sep byte = ' '
	for _, c := range matchingOf(datatype).columns {
		bob.WriteByte(sep)
		bob.WriteString(c.name)
		sep = ','
	}
	bob.WriteString(" FROM ")
	bob.WriteString(table)
	return bob.String()
}

// SelectAliasQuery returns a query string like that of SelectQuery except that a table alias is included, e.g.
//     SELECT <alias>.<col0>, <alias>.<col1>, ..., <alias>.<coln> FROM <table> <alias>
func SelectAliasQuery(table, alias string, datatype interface{}) string {
	bob := strings.Builder{}
	bob.WriteString("SELECT")
	var sep byte = ' '
	for _, c := range matchingOf(datatype).columns {
		bob.WriteByte(sep)
		bob.WriteString(alias)
		bob.WriteByte('.')
		bob.WriteString(c.name)
		sep = ','
	}
	bob.WriteString(" FROM ")
	bob.WriteString(table)
	bob.WriteByte(' ')
	bob.WriteString(alias)
	return bob.String()
}

// Where returns a string of the form
//     WHERE (<condition>) AND (<condition>) ...
// with a leading space.
//
// If no conditions are given, then Where returns the empty string.
func Where(conditions ...string) string {
	if len(conditions) == 0 {
		return ""
	}
	return " WHERE (" + strings.Join(conditions, ") AND (") + ")"
}

// LimitOffset returns a string of the form
//     LIMIT <limit> OFFSET <offset>
// with a leading space.
//
// If either limit or offset are zero, then that part of the string is omitted.  If both limit and offset are zero,
// then LimitOffset returns the empty string.
func LimitOffset(limit, offset int64) string {
	x := ""
	if limit != 0 {
		x = " LIMIT " + strconv.FormatInt(limit, 10)
	}
	if offset != 0 {
		x += " OFFSET " + strconv.FormatInt(offset, 10)
	}
	return x
}

// InsertQuery returns a query string of the form
//     INSERT INTO <table> (<columns>) VALUES (?,?,...)
//     INSERT INTO <table> (<columns>) VALUES ($1,$2,...)  (numbered placeholders)
// where <table> is the table name given, and <columns> is the list of the columns defined by the struct pointed at by
// datatype.  Struct fields tagged "readonly" are skipped.
//
// Panics if all fields are tagged "readonly".
func InsertQuery(table string, datatype interface{}) string {
	columns := matchingOf(datatype).columns
	bob := strings.Builder{}
	bob.WriteString("INSERT INTO ")
	bob.WriteString(table)
	bob.WriteByte(' ')
	var sep byte = '('
	var n int
	for _, c := range columns {
		if !c.readonly {
			bob.WriteByte(sep)
			bob.WriteString(c.name)
			sep = ','
			n++
		}
	}
	if n == 0 {
		panic("sx: struct " + matchingOf(datatype).reflectType.Name() + " has no writeable fields")
	}
	bob.WriteString(") VALUES ")
	sep = '('
	for p := Placeholder(0); p < Placeholder(n); {
		bob.WriteByte(sep)
		bob.WriteString(p.Next())
		sep = ','
	}
	bob.WriteByte(')')
	return bob.String()
}

// UpdateQuery returns a query string and a list of values from the struct pointed at by data.  This is the prefferred
// way to do updates, as it allows pointer fields in the struct and automatically skips zero values.
//
// The query string of the form
//     UPDATE <table> SET <column>=?,<column>=?,...
//     UPDATE <table> SET <column>=$2,<column>=$3,...  (numbered placeholders)
// where <table> is the table name given, and each <column> is a column name defined by the struct pointed at by data.
//
// With numbered placeholders, numbering starts at $2.  This allows $1 to be used in the WHERE clause.
//
// The list of values contains values from the struct to match the placeholders.  For pointer fields, the values
// pointed at are used.
//
// UpdateQuery takes all the writeable fields (not tagged "readonly") from the struct, looks up their values, and if
// it finds a zero value, the field is skipped.  This allows the caller to set only those values that need updating.
// If it is necessary to update a field to a zero value, then a pointer field should be used.  A pointer to a zero
// value will force an update, and a nil pointer will be skipped.
//
// The struct used for UpdateQuery will normally be a different struct from that used for select or insert on the
// same table.  This is okay.
//
// If there are no applicable fields, Update returns ("", nil).
func UpdateQuery(table string, data interface{}) (string, []interface{}) {
	m := matchingOf(data)
	instance := reflect.ValueOf(data).Elem()

	columns := make([]string, 0)
	values := make([]interface{}, 0)
	var p Placeholder = 1 // start from 2

	for _, c := range m.columns {
		if !c.readonly {
			if val := instance.Field(c.index); !valueIsZero(val) {
				columns = append(columns, c.name+"="+p.Next())
				if val.Kind() == reflect.Ptr {
					val = val.Elem()
				}
				values = append(values, val.Interface())
			}
		}
	}
	if len(columns) == 0 {
		return "", nil
	}

	return "UPDATE " + table + " SET " + strings.Join(columns, ","), values
}

// UpdateAllQuery returns a query string of the form
//     UPDATE <table> SET <column>=?,<column>=?,...
//     UPDATE <table> SET <column>=$2,<column>=$3,...  (numbered placeholders)
// where <table> is the table name given, and each <column> is a column name defined by the struct pointed at by
// data.  All writeable fields (those not tagged "readonly") are included.  Fields are in the order of the struct.
//
// With numbered placeholders, numbering starts at $2.  This allows $1 to be used in the WHERE clause.
//
// Use with the Values function to write to all writeable feilds.
func UpdateAllQuery(table string, data interface{}) string {
	m := matchingOf(data)
	columns := make([]string, 0)
	var p Placeholder = 1 // start from 2

	for _, c := range m.columns {
		if !c.readonly {
			columns = append(columns, c.name+"="+p.Next())
		}
	}

	return "UPDATE " + table + " SET " + strings.Join(columns, ",")
}

// UpdateFieldsQuery returns a query string and a list of values for the specified fields of the struct pointed at by data.
//
// The query string is of the form
//     UPDATE <table> SET <column>=?,<column>=?,...
//     UPDATE <table> SET <column>=$2,<column>=$3,...  (numbered placeholders)
// where <table> is the table name given, and each <column> is a column name defined by the struct pointed at by data.
//
// The list of values contains values from the struct to match the placeholders.  The order matches the the order of
// fields provided by the caller.
//
// With numbered placeholders, numbering starts at $2.  This allows $1 to be used in the WHERE clause.
//
// UpdateFieldsQuery panics if no field names are provided or if any of the requested fields do not exist.  If it is
// necessary to validate field names, use ColumnOf.
func UpdateFieldsQuery(table string, data interface{}, fields ...string) (string, []interface{}) {
	m := matchingOf(data)
	instance := reflect.ValueOf(data).Elem()

	columns := make([]string, 0)
	values := make([]interface{}, 0)
	var p Placeholder = 1 // start from 2

	if len(fields) == 0 {
		panic("UpdateFieldsQuery requires at least one field")
	}
	for _, field := range fields {
		if c, ok := m.columnMap[field]; ok {
			columns = append(columns, c.name+"="+p.Next())
			values = append(values, instance.Field(c.index).Interface())
		} else {
			panic("struct " + m.reflectType.Name() + " has no usable field " + field)
		}
	}

	return "UPDATE " + table + " SET " + strings.Join(columns, ","), values
}

// Addrs returns a slice of pointers to the fields of the struct pointed at by dest.  Use for scanning rows from a
// SELECT query.
//
// Panics if dest does not point at a struct.
func Addrs(dest interface{}) []interface{} {
	m := matchingOf(dest)
	val := reflect.ValueOf(dest).Elem()
	addrs := make([]interface{}, 0, len(m.columns))
	for _, c := range m.columns {
		addrs = append(addrs, val.Field(c.index).Addr().Interface())
	}
	return addrs
}

// Values returns a slice of values from the struct pointed at by data, excluding those from fields tagged "readonly".
// Use for providing values to an INSERT query.
//
// Panics if data does not point at a struct.
func Values(data interface{}) []interface{} {
	m := matchingOf(data)
	val := reflect.ValueOf(data).Elem()
	values := make([]interface{}, 0, len(m.columns))
	for _, c := range m.columns {
		if !c.readonly {
			values = append(values, val.Field(c.index).Interface())
		}
	}
	return values
}

// ValueOf returns the value of the specified field of the struct pointed at by data.  Panics if data does not
// point at a struct, or if the requested field doesn't exist.
func ValueOf(data interface{}, field string) interface{} {
	// This step verifies data and field and might panic.
	c := matchingOf(data).columnOf(field)
	// If there is a panic, then the reflection here will not be attempted.
	return reflect.ValueOf(data).Elem().Field(c.index).Interface()
}

// Columns returns the names of the database columns that correspond to the fields in the struct pointed at by
// datatype.  The order of returned fields matches the order of the struct.
func Columns(datatype interface{}) []string {
	return matchingOf(datatype).columnList()
}

// ColumnsWriteable returns the names of the database columns that correspond to the fields in the struct pointed at
// by datatype, excluding those tagged "readonly".  The order of returned fields matches the order of the struct.
func ColumnsWriteable(datatype interface{}) []string {
	return matchingOf(datatype).columnWriteableList()
}

// ColumnOf returns the name of the database column that corresponds to the specified field of the struct pointed
// at by datatype.
//
// ColumnOf returns an error if the provided field name is missing from the struct.
func ColumnOf(datatype interface{}, field string) (string, error) {
	m := matchingOf(datatype)
	if c, ok := m.columnMap[field]; ok {
		return c.name, nil
	}
	return "", errors.New("struct " + m.reflectType.Name() + " has no usable field " + field)
}
