package rules

import (
	"fmt"
	"reflect"
)

type requiredRule struct{}

func (*requiredRule) Name() string { return "required" }
func (*requiredRule) Validate(fv reflect.Value, _ reflect.StructField) error {
	if fv.IsZero() {
		return fmt.Errorf("field is required")
	}
	return nil
}
