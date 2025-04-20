package kdlconfig

import (
	"fmt"
	"os"

	"github.com/sblinch/kdl-go"
)

// Loader is responsible for reading, parsing and validating KDL configs into Go structures.
// First, kdl.Unmarshal is used to parse and map the data,
// then validation is performed using struct tags (`min`, `max`, `required`, etc.).
type Loader struct{}

// NewLoader creates a new Loader instance.
func NewLoader() *Loader {
	return &Loader{}
}

// Load reads the config file at path, unmarshals it using kdl.Unmarshal
// into the provided Go structure cfg (pointer), then performs validation.
func (l *Loader) Load(cfg interface{}, path string) error {
	// Reading the file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %q: %w", path, err)
	}

	// Unmarshaling via kdl.Unmarshal: checks KDL syntax correctness
	if err := kdl.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to unmarshal KDL: %w", err)
	}

	// Validation using struct tags (min, max, required, etc.)
	if err := validateStruct(cfg); err != nil {
		return err
	}
	return nil
}
