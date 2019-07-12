package sx

import (
	"reflect"
	"testing"
)

// Test structs

type menagerie0 struct {
	Platypus   string
	Rhinoceros float64
}

type menagerie1 struct {
	Chimpanzee int64  `sx:"human"`
	Flamingo   string `sx:",readonly"`
	Warthog    string `sx:"-"`
}

func TestSelectInsertUpdateAll(t *testing.T) {

	var testCases = []struct {
		name                 string
		table                string
		datatype             interface{}
		numberedPlaceholders bool
		wantSelect           string
		wantInsert           string
		wantUpdate           string
	}{
		{
			name:                 "menagerie0",
			table:                "zoo",
			datatype:             &menagerie0{},
			numberedPlaceholders: false,
			wantSelect:           "SELECT platypus,rhinoceros FROM zoo",
			wantInsert:           "INSERT INTO zoo (platypus,rhinoceros) VALUES (?,?)",
			wantUpdate:           "UPDATE zoo SET platypus=?,rhinoceros=?",
		},
		{
			name:                 "menagerie0 numbered",
			table:                "zoo",
			datatype:             &menagerie0{},
			numberedPlaceholders: true,
			wantSelect:           "SELECT platypus,rhinoceros FROM zoo",
			wantInsert:           "INSERT INTO zoo (platypus,rhinoceros) VALUES ($1,$2)",
			wantUpdate:           "UPDATE zoo SET platypus=$2,rhinoceros=$3",
		},
		{
			name:                 "menagerie1",
			table:                "jungle",
			datatype:             &menagerie1{},
			numberedPlaceholders: false,
			wantSelect:           "SELECT human,flamingo FROM jungle",
			wantInsert:           "INSERT INTO jungle (human) VALUES (?)",
			wantUpdate:           "UPDATE jungle SET human=?",
		},
		{
			name:                 "menagerie1 numbered",
			table:                "jungle",
			datatype:             &menagerie1{},
			numberedPlaceholders: true,
			wantSelect:           "SELECT human,flamingo FROM jungle",
			wantInsert:           "INSERT INTO jungle (human) VALUES ($1)",
			wantUpdate:           "UPDATE jungle SET human=$2",
		},
	}

	for _, c := range testCases {
		SetNumberedPlaceholders(c.numberedPlaceholders)
		if a, b := c.wantSelect, SelectQuery(c.table, c.datatype); a != b {
			t.Errorf("case %s select: expected \"%s\", got \"%s\"", c.name, a, b)
		}
		if a, b := c.wantInsert, InsertQuery(c.table, c.datatype); a != b {
			t.Errorf("case %s insert: expected \"%s\", got \"%s\"", c.name, a, b)
		}
		if a, b := c.wantUpdate, UpdateAllQuery(c.table, c.datatype); a != b {
			t.Errorf("case %s update all: expected \"%s\", got \"%s\"", c.name, a, b)
		}
	}
}

func TestSelectAlias(t *testing.T) {

	var testCases = []struct {
		name       string
		table      string
		alias      string
		datatype   interface{}
		wantSelect string
	}{
		{
			name:       "menagerie0",
			table:      "zoo",
			alias:      "home",
			datatype:   &menagerie0{},
			wantSelect: "SELECT home.platypus,home.rhinoceros FROM zoo home",
		},
		{
			name:       "menagerie1",
			table:      "jungle",
			alias:      "a",
			datatype:   &menagerie1{},
			wantSelect: "SELECT a.human,a.flamingo FROM jungle a",
		},
	}

	for _, c := range testCases {
		if a, b := c.wantSelect, SelectAliasQuery(c.table, c.alias, c.datatype); a != b {
			t.Errorf("case %s select alias: expected \"%s\", got \"%s\"", c.name, a, b)
		}
	}
}

func TestWhere(t *testing.T) {

	var testCases = []struct {
		name       string
		conditions []string
		want       string
	}{
		{
			name: "empty",
		},
		{
			name:       "one condition",
			conditions: []string{"a=5"},
			want:       " WHERE (a=5)",
		},
		{
			name:       "two conditions",
			conditions: []string{"a=5", "b=6"},
			want:       " WHERE (a=5) AND (b=6)",
		},
		{
			name:       "three conditions",
			conditions: []string{"a=5", "b=6", "c=7"},
			want:       " WHERE (a=5) AND (b=6) AND (c=7)",
		},
	}

	for _, c := range testCases {
		if a, b := c.want, Where(c.conditions...); a != b {
			t.Errorf("case %s: expected \"%s\", got \"%s\"", c.name, a, b)
		}
	}
}

func TestLimitOffset(t *testing.T) {

	var testCases = []struct {
		name   string
		limit  int64
		offset int64
		want   string
	}{
		{
			name: "empty",
		},
		{
			name:  "limit only",
			limit: 100,
			want:  " LIMIT 100",
		},
		{
			name:   "offset only",
			offset: 200,
			want:   " OFFSET 200",
		},
		{
			name:   "limit and offset",
			limit:  123,
			offset: 456,
			want:   " LIMIT 123 OFFSET 456",
		},
		{
			name:   "negative numbers", // no practical use for this, but the function should still process it
			limit:  -3,
			offset: -999999,
			want:   " LIMIT -3 OFFSET -999999",
		},
	}

	for _, c := range testCases {
		if a, b := c.want, LimitOffset(c.limit, c.offset); a != b {
			t.Errorf("case %s: expected \"%s\", got \"%s\"", c.name, a, b)
		}
	}
}

func TestInsertPanic(t *testing.T) {
	// InsertQuery should panic if all of a struct's fields are tagged readonly.
	type ohNo struct {
		field1 int `sx:",readonly"`
		field2 int `sx:",readonly"`
	}
	const wantPanic = "sx: struct ohNo has no usable fields"

	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("expected a panic")
				return
			}
			if s, ok := r.(string); ok {
				if s != wantPanic {
					t.Errorf("expected \"%s\", got \"%s\"", wantPanic, s)
				}
				return
			}
			panic(r)
		}()
		InsertQuery("zoo", &ohNo{})
	}()
}

func TestUpdate(t *testing.T) {

	type menagerie2 struct {
		Cougar      string
		Cheetah     *string `sx:"cat"`
		Grizzly     int32
		Wilderbeast []int32
		Ant         bool `sx:"-"`
		Bee         bool `sx:",readonly"`
	}

	var (
		x = []int32{1, 2, 3}
		y = int32(5)
		s = "abcde"
	)

	var testCases = []struct {
		name                 string
		table                string
		data                 interface{}
		numberedPlaceholders bool
		wantQuery            string
		wantValues           []interface{}
	}{
		{
			name:                 "empty",
			table:                "africa",
			data:                 &menagerie2{},
			numberedPlaceholders: false,
			wantQuery:            "",
			wantValues:           nil,
		},
		{
			name:                 "ignore ant and bee",
			table:                "asia",
			data:                 &menagerie2{Ant: true, Bee: true},
			numberedPlaceholders: true,
			wantQuery:            "",
			wantValues:           nil,
		},
		{
			name:                 "cougar and wilderbeast",
			table:                "australia",
			data:                 &menagerie2{Cougar: "abc", Wilderbeast: x},
			numberedPlaceholders: false,
			wantQuery:            "UPDATE australia SET cougar=?,wilderbeast=?",
			wantValues:           []interface{}{"abc", x},
		},
		{
			name:                 "cougar and wilderbeast numbered",
			table:                "australia",
			data:                 &menagerie2{Cougar: "abc", Wilderbeast: x},
			numberedPlaceholders: true,
			wantQuery:            "UPDATE australia SET cougar=$2,wilderbeast=$3",
			wantValues:           []interface{}{"abc", x},
		},
		{
			name:                 "cheetah and grizzly",
			table:                "siberia",
			data:                 &menagerie2{Cheetah: &s, Grizzly: y},
			numberedPlaceholders: false,
			wantQuery:            "UPDATE siberia SET cat=?,grizzly=?",
			wantValues:           []interface{}{s, y},
		},
		{
			name:                 "cheetah and grizzly numbered",
			table:                "siberia",
			data:                 &menagerie2{Cheetah: &s, Grizzly: y},
			numberedPlaceholders: true,
			wantQuery:            "UPDATE siberia SET cat=$2,grizzly=$3",
			wantValues:           []interface{}{s, y},
		},
		{
			name:                 "everyone except ant",
			table:                "berlin",
			data:                 &menagerie2{Cougar: "roar", Cheetah: &s, Grizzly: y, Wilderbeast: x, Bee: true},
			numberedPlaceholders: false,
			wantQuery:            "UPDATE berlin SET cougar=?,cat=?,grizzly=?,wilderbeast=?",
			wantValues:           []interface{}{"roar", s, y, x},
		},
		{
			name:                 "everyone except bee",
			table:                "berlin",
			data:                 &menagerie2{Cougar: "roar", Cheetah: &s, Grizzly: y, Wilderbeast: x, Ant: true},
			numberedPlaceholders: true,
			wantQuery:            "UPDATE berlin SET cougar=$2,cat=$3,grizzly=$4,wilderbeast=$5",
			wantValues:           []interface{}{"roar", s, y, x},
		},
	}

	for _, c := range testCases {
		SetNumberedPlaceholders(c.numberedPlaceholders)
		query, values := UpdateQuery(c.table, c.data)
		if a, b := c.wantQuery, query; a != b {
			t.Errorf("case %s query: expected \"%s\", got \"%s\"", c.name, a, b)
		}
		if a, b := c.wantValues, values; !reflect.DeepEqual(a, b) {
			t.Errorf("case %s values: expected %v, got %v", c.name, a, b)
		}
	}
}

func TestUpdateFields(t *testing.T) {

	var testCases = []struct {
		name                 string
		table                string
		data                 interface{}
		fields               []string
		numberedPlaceholders bool
		wantQuery            string
		wantValues           []interface{}
		wantPanic            string
	}{
		{
			name:      "no fields",
			table:     "irrelevant",
			data:      &menagerie0{},
			wantPanic: "UpdateFieldsQuery requires at least one field",
		},
		{
			name:      "invalid field",
			table:     "irrelevant",
			data:      &menagerie0{},
			fields:    []string{"Research"},
			wantPanic: "struct menagerie0 has no usable field Research",
		},
		{
			name:      "ignored field",
			table:     "irrelevant",
			data:      &menagerie1{},
			fields:    []string{"Warthog"},
			wantPanic: "struct menagerie1 has no usable field Warthog",
		},
		{
			name:                 "rhinoceros",
			table:                "swamp",
			data:                 &menagerie0{Platypus: "abc", Rhinoceros: -5.0},
			fields:               []string{"Rhinoceros"},
			numberedPlaceholders: false,
			wantQuery:            "UPDATE swamp SET rhinoceros=?",
			wantValues:           []interface{}{float64(-5.0)},
		},
		{
			name:                 "rhinoceros numbered",
			table:                "swamp",
			data:                 &menagerie0{Platypus: "abc", Rhinoceros: -10.0},
			fields:               []string{"Rhinoceros"},
			numberedPlaceholders: true,
			wantQuery:            "UPDATE swamp SET rhinoceros=$2",
			wantValues:           []interface{}{float64(-10.0)},
		},
		{
			name:                 "rhinoceros and platypus",
			table:                "swamp",
			data:                 &menagerie0{Platypus: "abc", Rhinoceros: -5.0},
			fields:               []string{"Rhinoceros", "Platypus"},
			numberedPlaceholders: false,
			wantQuery:            "UPDATE swamp SET rhinoceros=?,platypus=?",
			wantValues:           []interface{}{float64(-5.0), "abc"},
		},
		{
			name:                 "rhinoceros and platypus numbered",
			table:                "swamp",
			data:                 &menagerie0{Platypus: "abc", Rhinoceros: 10.0},
			fields:               []string{"Rhinoceros", "Platypus"},
			numberedPlaceholders: true,
			wantQuery:            "UPDATE swamp SET rhinoceros=$2,platypus=$3",
			wantValues:           []interface{}{float64(10.0), "abc"},
		},
		{
			name:                 "can update readonly field", // TODO: should we allow this?
			table:                "forest",
			data:                 &menagerie1{Chimpanzee: 123, Flamingo: "foo"},
			fields:               []string{"Chimpanzee", "Flamingo"},
			numberedPlaceholders: false,
			wantQuery:            "UPDATE forest SET human=?,flamingo=?",
			wantValues:           []interface{}{int64(123), "foo"},
		},
		{
			name:                 "can update readonly field numbered",
			table:                "forest",
			data:                 &menagerie1{Chimpanzee: 123, Flamingo: "foo"},
			fields:               []string{"Chimpanzee", "Flamingo"},
			numberedPlaceholders: true,
			wantQuery:            "UPDATE forest SET human=$2,flamingo=$3",
			wantValues:           []interface{}{int64(123), "foo"},
		},
	}

	type menagerie1 struct {
		Chimpanzee int64  `sx:"human"`
		Flamingo   string `sx:",readonly"`
		Warthog    string `sx:"-"`
	}

	for _, c := range testCases {

		var (
			query    string
			values   []interface{}
			gotPanic string
		)

		SetNumberedPlaceholders(c.numberedPlaceholders)

		func() {
			gotPanic = ""
			defer func() {
				r := recover()
				if r == nil {
					return
				}
				if s, ok := r.(string); ok {
					gotPanic = s
					return
				}
				panic(r)
			}()
			query, values = UpdateFieldsQuery(c.table, c.data, c.fields...)
		}()

		if gotPanic != c.wantPanic {
			if c.wantPanic == "" {
				t.Errorf("case %s: unexpected panic %q", c.name, gotPanic)
			} else if gotPanic == "" {
				t.Errorf("case %s: expected panic %q but got none", c.name, c.wantPanic)
			} else {
				t.Errorf("case %s: expected panic %q, got %q", c.name, c.wantPanic, gotPanic)
			}
			continue
		}

		if a, b := c.wantQuery, query; a != b {
			t.Errorf("case %s query: expected \"%s\", got \"%s\"", c.name, a, b)
		}
		if a, b := c.wantValues, values; !reflect.DeepEqual(a, b) {
			t.Errorf("case %s values: expected %v, got %v", c.name, a, b)
		}
	}
}

func TestAddrsValues(t *testing.T) {

	var (
		data0 = menagerie0{Platypus: "yes", Rhinoceros: 1.0}
		data1 = menagerie1{Chimpanzee: 64, Flamingo: "maybe", Warthog: "no"}
	)

	var testCases = []struct {
		name       string
		data       interface{}
		wantAddrs  []interface{}
		wantValues []interface{}
	}{
		{
			name:       "menagerie0",
			data:       &data0,
			wantAddrs:  []interface{}{&data0.Platypus, &data0.Rhinoceros},
			wantValues: []interface{}{"yes", float64(1.0)},
		},
		{
			name:       "menagerie1",
			data:       &data1,
			wantAddrs:  []interface{}{&data1.Chimpanzee, &data1.Flamingo},
			wantValues: []interface{}{int64(64)},
		},
	}

	// What's returned from Addrs is a slice of pointers, and we need to test that these are the exact pointers
	// that we want.  We can get a false true from DeepEqual, since DeepEqual considers two distinct pointers whose
	// pointed-at values are the same to be equal.
	shallowEqual := func(a, b []interface{}) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}

	for _, c := range testCases {
		if a, b := c.wantAddrs, Addrs(c.data); !shallowEqual(a, b) {
			t.Errorf("case %s addrs: expected %v, got %v", c.name, a, b)
		}
		if a, b := c.wantValues, Values(c.data); !reflect.DeepEqual(a, b) {
			t.Errorf("case %s values: expected %v, got %v", c.name, a, b)
		}
	}
}

func TestValueOf(t *testing.T) {

	var testCases = []struct {
		name      string
		data      interface{}
		field     string
		wantValue interface{}
	}{
		{
			name:      "platypus",
			data:      &menagerie0{Platypus: "pus", Rhinoceros: 2.0},
			field:     "Platypus",
			wantValue: "pus",
		},
		{
			name:      "rhinoceros",
			data:      &menagerie0{Platypus: "pus", Rhinoceros: 2.0},
			field:     "Rhinoceros",
			wantValue: float64(2.0),
		},
		{
			name:      "chimpanzee",
			data:      &menagerie1{Chimpanzee: 641, Flamingo: "goth", Warthog: "hog"},
			field:     "Chimpanzee",
			wantValue: int64(641),
		},
		{
			name:      "flamingo",
			data:      &menagerie1{Chimpanzee: 641, Flamingo: "goth", Warthog: "hog"},
			field:     "Flamingo",
			wantValue: "goth",
		},
	}

	for _, c := range testCases {
		if a, b := c.wantValue, ValueOf(c.data, c.field); a != b {
			t.Errorf("case %s: expected %v, got %v", c.name, a, b)
		}
	}
}

func TestColumnsColumnsWriteable(t *testing.T) {

	var testCases = []struct {
		name                 string
		datatype             interface{}
		wantColumns          []string
		wantColumnsWriteable []string
	}{
		{
			name:                 "menagerie0",
			datatype:             &menagerie0{},
			wantColumns:          []string{"platypus", "rhinoceros"},
			wantColumnsWriteable: []string{"platypus", "rhinoceros"},
		},
		{
			name:                 "menagerie1",
			datatype:             &menagerie1{},
			wantColumns:          []string{"human", "flamingo"},
			wantColumnsWriteable: []string{"human"},
		},
	}

	for _, c := range testCases {
		if a, b := c.wantColumns, Columns(c.datatype); !reflect.DeepEqual(a, b) {
			t.Errorf("case %s columns: expected %v, got %v", c.name, a, b)
		}
		if a, b := c.wantColumnsWriteable, ColumnsWriteable(c.datatype); !reflect.DeepEqual(a, b) {
			t.Errorf("case %s columns writeable: expected %v, got %v", c.name, a, b)
		}
	}
}

func TestColumnOf(t *testing.T) {

	var testCases = []struct {
		name       string
		datatype   interface{}
		field      string
		wantColumn string
		wantErr    string
	}{
		{
			name:       "platypus",
			datatype:   &menagerie0{},
			field:      "Platypus",
			wantColumn: "platypus",
		},
		{
			name:       "rhinoceros",
			datatype:   &menagerie0{},
			field:      "Rhinoceros",
			wantColumn: "rhinoceros",
		},
		{
			name:     "unknown column",
			datatype: &menagerie0{},
			field:    "Hippopotamus",
			wantErr:  "struct menagerie0 has no usable field Hippopotamus",
		},
		{
			name:       "chimpanzee",
			datatype:   &menagerie1{},
			field:      "Chimpanzee",
			wantColumn: "human",
		},
		{
			name:       "flamingo",
			datatype:   &menagerie1{},
			field:      "Flamingo",
			wantColumn: "flamingo",
		},
		{
			name:     "ignored field",
			datatype: &menagerie1{},
			field:    "Warthog",
			wantErr:  "struct menagerie1 has no usable field Warthog",
		},
	}

	for _, c := range testCases {
		column, err := ColumnOf(c.datatype, c.field)
		if err != nil {
			if c.wantErr == "" {
				t.Errorf("case %s: unexpected error %q", c.name, err.Error())
			} else if a, b := c.wantErr, err.Error(); a != b {
				t.Errorf("case %s: expected error %q, got %q", c.name, a, b)
			}
			continue
		}
		if a, b := c.wantColumn, column; a != b {
			t.Errorf("case %s: expected %v, got %v", c.name, a, b)
		}
	}
}
