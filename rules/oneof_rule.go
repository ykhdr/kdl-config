package rules

import (
	"fmt"
	"reflect"
)

type oneOfRule struct{ Options []string }

func (r *oneOfRule) Name() string { return "oneof" }
func (r *oneOfRule) Validate(fv reflect.Value, _ reflect.StructField) error {
	if fv.Kind() != reflect.String {
		return fmt.Errorf("oneof supports only string, got %s", fv.Kind())
	}
	s := fv.String()
	for _, o := range r.Options {
		if s == o {
			return nil
		}
	}
	return fmt.Errorf("value %q does not match any of the options: %v", s, r.Options)
}
