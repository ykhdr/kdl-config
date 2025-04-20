package rules

import (
	"fmt"
	"reflect"
	"regexp"
)

type patternRule struct{ Re *regexp.Regexp }

func (r *patternRule) Name() string { return "pattern" }
func (r *patternRule) Validate(fv reflect.Value, _ reflect.StructField) error {
	if fv.Kind() != reflect.String {
		return fmt.Errorf("pattern only works with strings, got %s", fv.Kind())
	}
	s := fv.String()
	if !r.Re.MatchString(s) {
		return fmt.Errorf("value %q does not match pattern %q", s, r.Re.String())
	}
	return nil
}
