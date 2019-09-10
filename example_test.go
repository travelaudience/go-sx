package sx_test

import (
	"fmt"

	sx "github.com/travelaudience/go-sx"
)

func ExampleSelectQuery() {
	type abc struct {
		Field1   int64
		FieldTwo string
		Field3   bool `sx:"gigo"`
	}
	query := sx.SelectQuery("sometable", &abc{})
	fmt.Println(query)
	// Output:
	// SELECT field1,field_two,gigo FROM sometable
}

func ExampleSelectAliasQuery() {
	type abc struct {
		Foo, Bar string
	}
	query := sx.SelectAliasQuery("sometable", "s", &abc{})
	fmt.Println(query)
	// Output:
	// SELECT s.foo,s.bar FROM sometable s
}

func ExampleWhere() {
	conditions := []string{
		"ordered",
		"NOT sent",
	}
	query := "SELECT * FROM sometable" + sx.Where(conditions...)
	fmt.Println(query)
	// Output:
	// SELECT * FROM sometable WHERE (ordered) AND (NOT sent)
}

func ExampleLimitOffset() {
	query := "SELECT * FROM sometable" + sx.LimitOffset(100, 0)
	fmt.Println(query)
	// Output:
	// SELECT * FROM sometable LIMIT 100
}

func ExampleInsertQuery() {
	sx.SetNumberedPlaceholders(true)
	type abc struct {
		Foo, Bar string
		Baz      int64 `sx:",readonly"`
	}
	query := sx.InsertQuery("sometable", &abc{})
	fmt.Println(query)
	// Output:
	// INSERT INTO sometable (foo,bar) VALUES ($1,$2)
}

func ExampleUpdateQuery() {
	sx.SetNumberedPlaceholders(true)
	type updateABC struct {
		Foo string  // cannot update to ""
		Bar *string // can update to ""
		Baz int64   // cannot update to 0
		Qux *int64  // can update to 0
	}

	s1, i1 := "hello", int64(0)
	x := updateABC{Bar: &s1, Baz: 42, Qux: &i1}
	query, values := sx.UpdateQuery("sometable", &x)
	query += " WHERE id=$1"
	fmt.Println(query)
	fmt.Println(values)

	query, values = sx.UpdateQuery("sometable", &updateABC{})
	fmt.Println(query == "", len(values))

	// Output:
	// UPDATE sometable SET bar=$2,baz=$3,qux=$4 WHERE id=$1
	// [hello 42 0]
	// true 0
}

func ExampleUpdateAllQuery() {
	sx.SetNumberedPlaceholders(true)
	type abc struct {
		Foo, Bar string
		Baz      int64 `sx:",readonly"`
	}
	query := sx.UpdateAllQuery("sometable", &abc{}) + " WHERE id=$1"
	fmt.Println(query)
	// Output:
	// UPDATE sometable SET foo=$2,bar=$3 WHERE id=$1
}

func ExampleUpdateFieldsQuery() {
	sx.SetNumberedPlaceholders(true)
	type abc struct {
		Foo, Bar string
		Baz      int64
	}
	x := abc{Foo: "hello", Bar: "Goodbye", Baz: 42}
	query, values := sx.UpdateFieldsQuery("sometable", &x, "Bar", "Baz")
	query += " WHERE id=$1"
	fmt.Println(query)
	fmt.Println(values)
	// Output:
	// UPDATE sometable SET bar=$2,baz=$3 WHERE id=$1
	// [Goodbye 42]
}
