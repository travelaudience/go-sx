package sx

import (
	"reflect"
	"testing"
)

func TestMatching(t *testing.T) {

	type test1 struct {
		A           int
		Lollipop    bool
		ChocolateID float64
		FOOBarBAZ   string
	}

	type test2 struct {
		A int `sx:"-"`
		B int `sx:"foo"`
		C int `sx:"bar"`
		d int `sx:"baz"`
	}

	type test3 struct {
		Zz int `sx:"zzz"`
		Yy int `sx:"yyy,readonly"`
		Xx int `sx:"-"`
		Ww int `sx:",readonly"`
	}

	type test4 struct {
		A int `sx:"-"`
	}

	type test5 struct {
		a int
	}

	t.Run("gets the correct results", func(t *testing.T) {

		var testCases = []struct {
			name                    string
			data                    interface{}
			wantColumnList          []string
			wantColumnWriteableList []string
		}{
			{
				name:                    "test1",
				data:                    &test1{},
				wantColumnList:          []string{"a", "lollipop", "chocolate_id", "foo_bar_baz"},
				wantColumnWriteableList: []string{"a", "lollipop", "chocolate_id", "foo_bar_baz"},
			},
			{
				name:                    "test2",
				data:                    &test2{},
				wantColumnList:          []string{"foo", "bar"},
				wantColumnWriteableList: []string{"foo", "bar"},
			},
			{
				name:                    "test3",
				data:                    &test3{},
				wantColumnList:          []string{"zzz", "yyy", "ww"},
				wantColumnWriteableList: []string{"zzz"},
			},
		}

		for _, c := range testCases {
			m := matchingOf(c.data)
			if a, b := c.wantColumnList, m.columnList(); !reflect.DeepEqual(a, b) {
				t.Errorf("case %s: expected columns %v, got %v", c.name, a, b)
			}
			if a, b := c.wantColumnWriteableList, m.columnWriteableList(); !reflect.DeepEqual(a, b) {
				t.Errorf("case %s: expected columnsWriteable %v, got %v", c.name, a, b)
			}
		}
	})

	t.Run("panics on bad input", func(t *testing.T) {

		var testCases = []struct {
			name      string
			data      interface{}
			wantPanic string
		}{

			{
				name:      "pass a struct, not a pointer",
				data:      test1{},
				wantPanic: "sx: expected a pointer to a struct",
			},
			{
				name:      "pass nil",
				data:      nil,
				wantPanic: "sx: expected a pointer to a struct",
			},
			{
				name:      "pass something else",
				data:      "hello",
				wantPanic: "sx: expected a pointer to a struct",
			},
			{
				name:      "no usable fields",
				data:      &test4{},
				wantPanic: "sx: struct test4 has no usable fields",
			},
			{
				name:      "no exported fields",
				data:      &test5{},
				wantPanic: "sx: struct test5 has no usable fields",
			},
		}

		for _, c := range testCases {
			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("case %s: expected a panic", c.name)
						return
					}
					if s, ok := r.(string); ok {
						if s != c.wantPanic {
							t.Errorf("case %s: expected \"%s\", got \"%s\"", c.name, c.wantPanic, s)
						}
						return
					}
					panic(r)
				}()
				matchingOf(c.data)
			}()
		}
	})

	t.Run("ColumnOf gets the correct results", func(t *testing.T) {

		var testCases = []struct {
			name           string
			data           interface{}
			field          string
			wantColumnName *column
		}{
			{
				name:           "test1-A",
				data:           &test1{},
				field:          "A",
				wantColumnName: &column{index: 0, name: "a"},
			},
			{
				name:           "test2-B",
				data:           &test2{},
				field:          "B",
				wantColumnName: &column{index: 1, name: "foo"},
			},
			{
				name:           "test3-Ww",
				data:           &test3{},
				field:          "Ww",
				wantColumnName: &column{index: 3, name: "ww", readonly: true},
			},
		}

		for _, c := range testCases {
			m := matchingOf(c.data)
			if a, b := c.wantColumnName, m.columnOf(c.field); !reflect.DeepEqual(a, b) {
				t.Errorf("case %s: expected columns %v, got %v", c.name, a, b)
			}
		}
	})

	t.Run("ColumnOf panics on unknown field", func(t *testing.T) {

		var testCases = []struct {
			name      string
			data      interface{}
			field     string
			wantPanic string
		}{
			{
				name:      "unknown field",
				data:      &test1{},
				field:     "Zzzzz",
				wantPanic: "sx: struct test1 has no usable field Zzzzz",
			},
			{
				name:      "ignored field",
				data:      &test2{},
				field:     "A",
				wantPanic: "sx: struct test2 has no usable field A",
			},
			{
				name:      "unexported field",
				data:      &test2{},
				field:     "d",
				wantPanic: "sx: struct test2 has no usable field d",
			},
		}

		for _, c := range testCases {

			m := matchingOf(c.data)

			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("case %s: expected a panic", c.name)
						return
					}
					if s, ok := r.(string); ok {
						if s != c.wantPanic {
							t.Errorf("case %s: expected %q, got %q", c.name, c.wantPanic, s)
						}
						return
					}
					panic(r)
				}()
				m.columnOf(c.field)
			}()
		}
	})
}
