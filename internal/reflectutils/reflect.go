package reflectutils

import (
	"fmt"
	"reflect"
)

func GetNumericValue(fv reflect.Value) (float64, error) {
	switch fv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(fv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(fv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return fv.Float(), nil
	default:
		return 0, fmt.Errorf("the rule is applicable only to numbers, received %s", fv.Kind())
	}
}
