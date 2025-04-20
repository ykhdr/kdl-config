// validate_test.go
package kdlconfig

import (
	"fmt"
	"github.com/ykhdr/kdl-config/internal/reflectutils"
	"github.com/ykhdr/kdl-config/rules"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func setup() {
	rules.RegisterDefaultRules()
}

type basicStruct struct {
	Value int    `validate:"required,min=1,max=10"`
	Name  string `validate:"len=5"`
	Env   string `validate:"oneof=dev|prod"`
}

func TestValidateStruct_Basic_Success(t *testing.T) {
	setup()
	cfg := &basicStruct{Value: 5, Name: "Hello", Env: "dev"}
	err := validateStruct(cfg)
	require.NoError(t, err)
}

func TestValidateStruct_Basic_Failures(t *testing.T) {
	tests := []struct {
		name   string
		cfg    *basicStruct
		expSub string
	}{
		{"missing required", &basicStruct{Value: 0, Name: "Hello", Env: "dev"}, "Value"},
		{"below min", &basicStruct{Value: -1, Name: "Hello", Env: "dev"}, "Value"},
		{"above max", &basicStruct{Value: 11, Name: "Hello", Env: "dev"}, "Value"},
		{"len fail", &basicStruct{Value: 5, Name: "Hi", Env: "dev"}, "Name"},
		{"oneof fail", &basicStruct{Value: 5, Name: "Hello", Env: "qa"}, "Env"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()
			err := validateStruct(tt.cfg)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expSub)
		})
	}
}

type patternStruct struct {
	Pattern string `validate:"pattern=^user_[0-9]+$"`
}

func TestValidateStruct_PatternRule(t *testing.T) {
	setup()
	valid := &patternStruct{Pattern: "user_123"}
	require.NoError(t, validateStruct(valid))

	invalid := &patternStruct{Pattern: "invalid"}
	err := validateStruct(invalid)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Pattern")
}

type sliceStruct struct {
	Items []int `validate:"len=3"`
}

func TestValidateStruct_LenSlice(t *testing.T) {
	setup()
	good := &sliceStruct{Items: []int{1, 2, 3}}
	require.NoError(t, validateStruct(good))

	bad := &sliceStruct{Items: []int{1, 2}}
	err := validateStruct(bad)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Items")
}

type mapStruct struct {
	M map[string]int `validate:"len=2"`
}

func TestValidateStruct_LenMap(t *testing.T) {
	setup()
	good := &mapStruct{M: map[string]int{"a": 1, "b": 2}}
	require.NoError(t, validateStruct(good))

	bad := &mapStruct{M: map[string]int{"a": 1}}
	err := validateStruct(bad)
	require.Error(t, err)
	require.Contains(t, err.Error(), "M")
}

type noTagStruct struct {
	A int
	B string
	C []float64
}

func TestValidateStruct_NoTags(t *testing.T) {
	setup()
	cfg := &noTagStruct{A: 0, B: "", C: nil}
	require.NoError(t, validateStruct(cfg))
}

type unknownRuleStruct struct {
	X int `validate:"foobar=123"`
}

func TestValidateStruct_UnknownRule(t *testing.T) {
	setup()
	cfg := &unknownRuleStruct{X: 1}
	err := validateStruct(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown validation rule")
}

type oneOfIntStruct struct {
	A int `validate:"oneof=1|2|3"`
}

func TestValidateStruct_OneOfTypeMismatch(t *testing.T) {
	setup()
	cfg := &oneOfIntStruct{A: 2}
	// oneof поддерживает только строки
	err := validateStruct(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "supports only string")
}

type customStruct struct {
	A int `validate:"isodd"`
}

type isOddRule struct{}

func (r *isOddRule) Name() string { return "isodd" }
func (r *isOddRule) Validate(fv reflect.Value, _ reflect.StructField) error {
	v, err := reflectutils.GetNumericValue(fv)
	if err != nil {
		return err
	}
	if int64(v)%2 == 0 {
		return fmt.Errorf("value %v is not odd", v)
	}
	return nil
}

func TestValidateStruct_CustomRule(t *testing.T) {
	setup()
	rules.RegisterRule("isodd", func(_ string) (rules.Rule, error) {
		return &isOddRule{}, nil
	})

	cfg := &customStruct{A: 3}
	require.NoError(t, validateStruct(cfg))

	cfg.A = 4
	err := validateStruct(cfg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not odd")
}
