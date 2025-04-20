package rules

import (
	"fmt"
	"github.com/ykhdr/kdl-config/internal/reflectutils"
	"reflect"
)

type minRule struct{ Min float64 }

func (r *minRule) Name() string { return "min" }
func (r *minRule) Validate(fv reflect.Value, _ reflect.StructField) error {
	v, err := reflectutils.GetNumericValue(fv)
	if err != nil {
		return err
	}
	if v < r.Min {
		return fmt.Errorf("value %v < min %v", v, r.Min)
	}
	return nil
}
