package sx

import (
	"reflect"

	"github.com/google/go-cmp/cmp"
)

// StructDiff finds all the differences between a and b that are representable in patch.  A and b must be structs
// of the same type, and patch is a pointer to a struct that will be updated with all values from b that differ
// from those of a.
//
// Fields are matched by name.  If the patch field is a pointer field, the corresponding field in b must be
// assignable to the type pointed to.  Otherwise, the corresponding field in b must be must be assignable to
// the patch field itself.
//
// StructDiff panics if the patch field is not a pointer field and the value in b is the zero value and different
// from the value in a.  It is left as an exercise to the reader to understand why this panic is necessary.
//
// StructDiff returns the number of fields set in patch.
func StructDiff(a, b, patch interface{}) int {

	av, bv, pv := reflect.ValueOf(a), reflect.ValueOf(b), reflect.ValueOf(patch)
	if av.Kind() != reflect.Struct || bv.Kind() != reflect.Struct || av.Type() != bv.Type() {
		panic("cannot generate patch from non-structs or structs of different types")
	}
	if pv.Kind() != reflect.Ptr || pv.Elem().Kind() != reflect.Struct {
		panic("patch must point to a struct")
	}
	pv = pv.Elem()
	pt := pv.Type()

	var count int
	for i := 0; i < pt.NumField(); i++ {
		fieldName := pt.Field(i).Name
		aval, bval := av.FieldByName(fieldName), bv.FieldByName(fieldName)
		// Using cmp.Equal here so that time fields get compared as expected.
		if aval.IsValid() && bval.IsValid() && !cmp.Equal(aval.Interface(), bval.Interface()) {
			pval := pv.Field(i)
			if pval.Kind() == reflect.Ptr {
				x := reflect.New(bval.Type())
				x.Elem().Set(bval)
				pval.Set(x)
			} else if bval.IsZero() {
				panic("cannot set field " + fieldName + " to its zero value")
			} else {
				pval.Set(bval)
			}
			count++
		}
	}
	return count
}
