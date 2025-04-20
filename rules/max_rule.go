package rules

import (
	"fmt"
	"github.com/ykhdr/kdl-config/internal/reflectutils"
	"reflect"
)

type maxRule struct{ Max float64 }

func (r *maxRule) Name() string { return "max" }
func (r *maxRule) Validate(fv reflect.Value, _ reflect.StructField) error {
	v, err := reflectutils.GetNumericValue(fv)
	if err != nil {
		return err
	}
	if v > r.Max {
		return fmt.Errorf("value %v > max %v", v, r.Max)
	}
	return nil
}
