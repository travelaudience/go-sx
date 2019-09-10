package sx_test

import (
	"testing"

	sx "github.com/travelaudience/go-sx"
)

// The tests in this file test that the correct panics are generated.  The tests in helpers_test.go test for
// the correct results.

func TestMatching(t *testing.T) {

	type test1 struct {
		A           int
		Lollipop    bool
		ChocolateID float64
		FOOBarBAZ   string
	}

	type test2 struct {
		A int `sx:"-"`
	}

	type test3 struct {
		a int
	}

	type test4 struct {
		A int `sx:"-"`
		B int `sx:"foo"`
		C int `sx:"bar"`
		d int `sx:"baz"`
	}

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
				data:      &test2{},
				wantPanic: "sx: struct test2 has no usable fields",
			},
			{
				name:      "no exported fields",
				data:      &test3{},
				wantPanic: "sx: struct test3 has no usable fields",
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
				// this calls matchingOf(c.data) straight away
				sx.Values(c.data)
			}()
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
				data:      &test4{},
				field:     "A",
				wantPanic: "sx: struct test4 has no usable field A",
			},
			{
				name:      "unexported field",
				data:      &test4{},
				field:     "d",
				wantPanic: "sx: struct test4 has no usable field d",
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
							t.Errorf("case %s: expected %q, got %q", c.name, c.wantPanic, s)
						}
						return
					}
					panic(r)
				}()
				// this calls matchingOf(c.data).ColumnOf(c.field)
				sx.ValueOf(c.data, c.field)
			}()
		}
	})
}
