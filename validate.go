package kdlconfig

import (
	"fmt"
	"github.com/ykhdr/kdl-config/rules"
	"reflect"
	"strings"
)

// ValidationError describes a single validation error.
type ValidationError struct {
	Field string
	Msg   string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed on %q: %s", e.Field, e.Msg)
}

// ValidationErrors aggregates multiple ValidationError instances.
type ValidationErrors []ValidationError

func (es ValidationErrors) Error() string {
	var b strings.Builder
	for i, e := range es {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(e.Error())
	}
	return b.String()
}

// validateStruct iterates over the fields of the cfg structure (pointer to struct),
// reads from the struct tag rawTag := ft.Tag.Get("validate"),
// then for each rule calls rules.GetRule(rawRule) and invokes rule.Validate().
func validateStruct(cfg any) error {
	// First, register all built-in rules (with a single call, idempotent).
	rules.RegisterDefaultRules()

	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("validateStruct: expected pointer to struct, got %T", cfg)
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("validateStruct: expected struct, got %T", v.Interface())
	}

	var allErrs ValidationErrors
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		rawTag := sf.Tag.Get("validate")
		if rawTag == "" {
			continue
		}
		fv := v.Field(i)

		for _, rawRule := range strings.Split(rawTag, ",") {
			rawRule = strings.TrimSpace(rawRule)
			rule, err := rules.GetRule(rawRule)
			if err != nil {
				allErrs = append(allErrs, ValidationError{Field: sf.Name, Msg: err.Error()})
				continue
			}
			if err := rule.Validate(fv, sf); err != nil {
				allErrs = append(allErrs, ValidationError{Field: sf.Name, Msg: err.Error()})
			}
		}
	}

	if len(allErrs) > 0 {
		return allErrs
	}
	return nil
}
