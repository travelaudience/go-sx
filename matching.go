package sx

import (
	"reflect"
	"regexp"
	"strings"
	"sync"
)

// A matching is between struct fields and database columns.
type matching struct {
	reflectType reflect.Type
	columns     []*column          // an ordered list of columns
	columnMap   map[string]*column // columns keyed by field name
}

type column struct {
	index    int    // index of this field in the struct
	name     string // name of the corresponding db column
	readonly bool   // flag to skip this column on insert/update operations (e.g. for primary key or automatic timestamp)
}

// ColumnList returns the names of the database columns in the order of the struct.
func (m *matching) columnList() []string {
	list := make([]string, 0, len(m.columns))
	for _, c := range m.columns {
		list = append(list, c.name)
	}
	return list
}

// ColumnWriteableList returns the names of the database columns in the order of the struct, excluding read-only columns.
func (m *matching) columnWriteableList() []string {
	list := make([]string, 0, len(m.columns))
	for _, c := range m.columns {
		if !c.readonly {
			list = append(list, c.name)
		}
	}
	return list
}

// ColumnOf returns the column which matches the named field.  Panics if the field doesn't exist.
func (m *matching) columnOf(field string) *column {
	if c, ok := m.columnMap[field]; ok {
		return c
	}
	panic("sx: struct " + m.reflectType.Name() + " has no usable field " + field)
}

// MatchingOf returns a matching for the given struct type, generating it if necessary.  MatchingOf looks only at the
// structure of datatype and ignore its values.
//
// Panics if datatype does not point at a struct, or if the struct has no usable fields.
func matchingOf(datatype interface{}) *matching {
	matchingCacheMu.Lock()
	defer matchingCacheMu.Unlock()

	v := reflect.ValueOf(datatype)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		panic("sx: expected a pointer to a struct")
	}

	// First look for a cached matching.
	reflectType := v.Elem().Type()
	if m, ok := matchingCache[reflectType]; ok {
		return m
	}

	// Nothing cached, generate a new matching and cache it.
	n := reflectType.NumField()
	cols := make([]*column, 0)
	colmap := make(map[string]*column)
	for i := 0; i < n; i++ {
		field := reflectType.Field(i)
		tags := strings.Split(field.Tag.Get("sx"), ",")
		colname := tags[0]
		if colname == "-" || field.PkgPath != "" {
			continue // skip excluded and unexported fields.
		}
		if colname == "" {
			colname = snakeCase(field.Name) // default column name based on field name
		}
		col := &column{
			index: i,
			name:  colname,
		}
		// See if there's a readonly tag.  A readonly tag would have to be in at least the second position, since
		// the first position is always interpreted as a column name.
		for _, tag := range tags[1:] {
			if tag == "readonly" {
				col.readonly = true
				break
			}
		}
		cols = append(cols, col)
		colmap[field.Name] = col
	}
	if len(cols) == 0 {
		panic("sx: struct " + reflectType.Name() + " has no usable fields")
	}

	m := &matching{
		reflectType: reflectType,
		columns:     cols,
		columnMap:   colmap,
	}
	matchingCache[reflectType] = m
	return m
}

// Cache to keep track of struct types that have been seen and therefore analyzed.
var matchingCache = make(map[reflect.Type]*matching)
var matchingCacheMu sync.Mutex

// Snake-casing logic.

var (
	matchWord    = regexp.MustCompile(`(.)([A-Z][a-z]+)`)
	matchAcronym = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

func snakeCase(in string) string {
	const r = `${1}_${2}`
	return strings.ToLower(matchAcronym.ReplaceAllString(matchWord.ReplaceAllString(in, r), r))
}
