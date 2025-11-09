package rules

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRule(t *testing.T) {
	type args struct {
		raw string
	}
	tests := []struct {
		name    string
		args    args
		want    Rule
		wantErr bool
	}{
		{
			name:    "Valid required rule",
			args:    args{raw: "required"},
			want:    &requiredRule{},
			wantErr: false,
		},
		{
			name:    "Valid min rule",
			args:    args{raw: "min=5"},
			want:    &minRule{Min: 5},
			wantErr: false,
		},
		{
			name:    "Invalid min rule with non-numeric value",
			args:    args{raw: "min=abc"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Unknown rule",
			args:    args{raw: "unknown=123"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterDefaultRules()
			got, err := GetRule(tt.args.raw)
			require.Equal(t, tt.wantErr, err != nil, fmt.Sprintf("GetRule() error = %v, wantErr %v", err, tt.wantErr))
			require.True(t, reflect.DeepEqual(got, tt.want), fmt.Sprintf("GetRule() got = %v, want %v", got, tt.want))
		})
	}
}

func TestRegisterDefaultRules(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Register default rules"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear the registry before each test
			regMu.Lock()
			registry = make(map[string]RuleFactory)
			defaultRulesLoaded = false
			regMu.Unlock()

			RegisterDefaultRules()

			// List of expected default rules
			expectedRules := []string{"required", "min", "max", "len", "oneof", "pattern"}

			// Check if each expected rule is registered
			for _, rule := range expectedRules {
				regMu.RLock()
				_, exists := registry[rule]
				regMu.RUnlock()
				require.True(t, exists, fmt.Sprintf("expected rule %q to be registered", rule))
			}
		})
	}
}

// customRule is a simple implementation for testing custom rule registration
type customRule struct{}

func (*customRule) Name() string { return "custom" }
func (*customRule) Validate(_ reflect.Value, _ reflect.StructField) error {
	// This is a placeholder implementation for testing purposes
	// In a real scenario, this would contain actual validation logic
	return nil
}

func TestRegisterRule(t *testing.T) {
	type args struct {
		name    string
		factory RuleFactory
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Register required rule",
			args: args{
				name: "required",
				factory: func(param string) (Rule, error) {
					return &requiredRule{}, nil
				},
			},
		},
		{
			name: "Register min rule",
			args: args{
				name: "min",
				factory: func(param string) (Rule, error) {
					return &minRule{Min: 5}, nil
				},
			},
		},
		{
			name: "Register custom rule",
			args: args{
				name: "custom",
				factory: func(param string) (Rule, error) {
					return &customRule{}, nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				RegisterRule(tt.args.name, tt.args.factory)
			})
		})
	}
}

// Basic unit tests for individual rules
func TestRequiredRule_Basics(t *testing.T) {
	r := &requiredRule{}

	// string
	var s string
	require.Error(t, r.Validate(reflect.ValueOf(s), reflect.StructField{}))
	s = "x"
	require.NoError(t, r.Validate(reflect.ValueOf(s), reflect.StructField{}))

	// slice
	var sl []int
	require.Error(t, r.Validate(reflect.ValueOf(sl), reflect.StructField{}))
	sl = []int{}
	require.Error(t, r.Validate(reflect.ValueOf(sl), reflect.StructField{}))
	sl = []int{1}
	require.NoError(t, r.Validate(reflect.ValueOf(sl), reflect.StructField{}))

	// map
	var m map[string]int
	require.Error(t, r.Validate(reflect.ValueOf(m), reflect.StructField{}))
	m = map[string]int{}
	require.Error(t, r.Validate(reflect.ValueOf(m), reflect.StructField{}))
	m["a"] = 1
	require.NoError(t, r.Validate(reflect.ValueOf(m), reflect.StructField{}))

	// pointer
	var p *int
	require.Error(t, r.Validate(reflect.ValueOf(p), reflect.StructField{}))
	x := 10
	p = &x
	require.NoError(t, r.Validate(reflect.ValueOf(p), reflect.StructField{}))
}

func TestLenRule_UnsupportedType(t *testing.T) {
	r := &lenRule{Length: 1}
	require.Error(t, r.Validate(reflect.ValueOf(123), reflect.StructField{}))
}

func TestPatternRule_NonString(t *testing.T) {
	r := &patternRule{Re: regexp.MustCompile("^x$")}
	require.Error(t, r.Validate(reflect.ValueOf(123), reflect.StructField{}))
}

func TestOneOfRule_Basics(t *testing.T) {
	r := &oneOfRule{Options: []string{"a", "b"}}
	// success
	require.NoError(t, r.Validate(reflect.ValueOf("a"), reflect.StructField{}))
	// failure
	require.Error(t, r.Validate(reflect.ValueOf("c"), reflect.StructField{}))
	// wrong type
	require.Error(t, r.Validate(reflect.ValueOf(1), reflect.StructField{}))
}

func TestMinMaxRule_Floats(t *testing.T) {
	min := &minRule{Min: 1.5}
	max := &maxRule{Max: 3.0}

	require.Error(t, min.Validate(reflect.ValueOf(1.0), reflect.StructField{}))
	require.NoError(t, min.Validate(reflect.ValueOf(1.5), reflect.StructField{}))
	require.NoError(t, min.Validate(reflect.ValueOf(2.0), reflect.StructField{}))

	require.NoError(t, max.Validate(reflect.ValueOf(3.0), reflect.StructField{}))
	require.NoError(t, max.Validate(reflect.ValueOf(2.5), reflect.StructField{}))
	require.Error(t, max.Validate(reflect.ValueOf(3.5), reflect.StructField{}))
}
