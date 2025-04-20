package rules

import (
	"fmt"
	"reflect"
)

type lenRule struct{ Length int }

func (r *lenRule) Name() string { return "len" }
func (r *lenRule) Validate(fv reflect.Value, _ reflect.StructField) error {
	var l int
	switch fv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		l = fv.Len()
	default:
		return fmt.Errorf("len rule is not supported for %s", fv.Kind())
	}
	if l != r.Length {
		return fmt.Errorf("length %d != %d", l, r.Length)
	}
	return nil
}
