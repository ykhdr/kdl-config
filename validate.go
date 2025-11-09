package kdlconfig

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ykhdr/kdl-config/rules"
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
	if v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("validateStruct: expected struct, got %T", v.Interface())
	}

	visited := make(map[uintptr]bool)
	// do not mark root struct address as visited for struct-field recursion
	if err := validateStructFields(v, visited); err != nil {
		return err
	}
	return nil
}

// validateStructFields validates a pointer to struct with cycle detection.
func validateStructFields(pv reflect.Value, visited map[uintptr]bool) error {
	// pv must be a non-nil pointer to struct
	if pv.Kind() != reflect.Ptr || pv.IsNil() || pv.Elem().Kind() != reflect.Struct {
		return nil
	}

	var allErrs ValidationErrors
	v := pv.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		fv := v.Field(i)

		// Recursive descent into nested structs while avoiding cycles
		switch fv.Kind() {
		case reflect.Struct:
			// Always recurse into embedded struct value; no cycle risk without pointers.
			addr := fv.Addr()
			if err := validateStructFields(addr, visited); err != nil {
				if verrs, ok := err.(ValidationErrors); ok {
					allErrs = append(allErrs, verrs...)
				} else {
					allErrs = append(allErrs, ValidationError{Field: sf.Name, Msg: err.Error()})
				}
			}
		case reflect.Ptr:
			if !fv.IsNil() && fv.Elem().Kind() == reflect.Struct {
				ptr := fv.Pointer()
				if !visited[ptr] {
					visited[ptr] = true
					if err := validateStructFields(fv, visited); err != nil {
						if verrs, ok := err.(ValidationErrors); ok {
							allErrs = append(allErrs, verrs...)
						} else {
							allErrs = append(allErrs, ValidationError{Field: sf.Name, Msg: err.Error()})
						}
					}
				}
			}
		}

		rawTag := sf.Tag.Get("validate")
		if rawTag != "" {
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
	}

	if len(allErrs) > 0 {
		return allErrs
	}
	return nil
}
