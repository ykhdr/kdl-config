package rules

import (
	"fmt"
	"reflect"
)

type requiredRule struct{}

func (*requiredRule) Name() string { return "required" }
func (*requiredRule) Validate(fv reflect.Value, _ reflect.StructField) error {
	// Semantics:
	// - slice/map: must be non-nil and non-empty
	// - array: length > 0
	// - ptr/interface: must be non-nil
	// - other kinds: must not be zero value
	switch fv.Kind() {
	case reflect.Slice, reflect.Map:
		if fv.IsNil() || fv.Len() == 0 {
			return fmt.Errorf("field is required (non-empty %s)", fv.Kind())
		}
	case reflect.Array:
		if fv.Len() == 0 {
			return fmt.Errorf("field is required (non-empty array)")
		}
	case reflect.Ptr, reflect.Interface:
		if fv.IsNil() {
			return fmt.Errorf("field is required (non-nil %s)", fv.Kind())
		}
	default:
		if fv.IsZero() {
			return fmt.Errorf("field is required")
		}
	}
	return nil
}
