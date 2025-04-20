package rules

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// ─────────────────────────────────────────────────────────────────────────────
// Rule interface and factory
// ─────────────────────────────────────────────────────────────────────────────

// Rule describes a single validation check.
type Rule interface {
	Name() string
	Validate(fv reflect.Value, sf reflect.StructField) error
}

// RuleFactory creates a Rule based on a parameter (for example "min=10" → MinRule{10}).
type RuleFactory func(param string) (Rule, error)

// ─────────────────────────────────────────────────────────────────────────────
// Registry and registration
// ─────────────────────────────────────────────────────────────────────────────

var (
	regMu              sync.RWMutex
	registry           = make(map[string]RuleFactory)
	defaultRulesLoaded bool
)

// RegisterRule thread-safely registers a new rule factory.
func RegisterRule(name string, factory RuleFactory) {
	regMu.Lock()
	defer regMu.Unlock()
	registry[name] = factory
}

// GetRule returns a Rule instance from a "raw" string, for example "min=5".
func GetRule(raw string) (Rule, error) {
	parts := strings.SplitN(raw, "=", 2)
	name := parts[0]

	regMu.RLock()
	factory, ok := registry[name]
	regMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown validation rule %q", name)
	}

	param := ""
	if len(parts) == 2 {
		param = parts[1]
	}
	return factory(param)
}

// RegisterDefaultRules registers built-in rules with **one** call.
// Can be called multiple times — actual registration will happen only once.
func RegisterDefaultRules() {
	regMu.Lock()
	defer regMu.Unlock()

	if defaultRulesLoaded {
		return
	}
	defaultRulesLoaded = true

	// required
	registry["required"] = func(_ string) (Rule, error) {
		return &requiredRule{}, nil
	}
	// min
	registry["min"] = func(param string) (Rule, error) {
		f, err := strconv.ParseFloat(param, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid min value %q: %w", param, err)
		}
		return &minRule{Min: f}, nil
	}
	// max
	registry["max"] = func(param string) (Rule, error) {
		f, err := strconv.ParseFloat(param, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid max value %q: %w", param, err)
		}
		return &maxRule{Max: f}, nil
	}
	// len
	registry["len"] = func(param string) (Rule, error) {
		n, err := strconv.Atoi(param)
		if err != nil {
			return nil, fmt.Errorf("invalid len value %q: %w", param, err)
		}
		return &lenRule{Length: n}, nil
	}
	// oneof
	registry["oneof"] = func(param string) (Rule, error) {
		opts := strings.Split(param, "|")
		if len(opts) == 0 {
			return nil, fmt.Errorf("oneof: at least one option must be specified")
		}
		return &oneOfRule{Options: opts}, nil
	}
	// pattern (regex)
	registry["pattern"] = func(param string) (Rule, error) {
		re, err := regexp.Compile(param)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", param, err)
		}
		return &patternRule{Re: re}, nil
	}
}
