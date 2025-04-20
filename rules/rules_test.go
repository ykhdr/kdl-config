package rules

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
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
