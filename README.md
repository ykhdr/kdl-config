# KDL config

A Go library for parsing, validating, and hot‑reloading [KDL](https://kdl.dev/) configuration files.

## Features

- **Struct‑based mapping** via `kdl.Unmarshal`.
- **Declarative validation** with struct tags:
    - `required`
    - `min=<number>`, `max=<number>`
    - `len=<number>` (strings, slices, arrays, maps)
    - `oneof=a|b|c` (string enums)
    - `pattern=<regexp>` (strings)
- **Custom rules**: register your own validation logic.
- **Hot reload**: watch file changes and automatically reload/validate.

## Requirements

- **Go** 1.20 or later

## Dependencies

- [github.com/sblinch/kdl-go](https://github.com/sblinch/kdl-go) — used under the hood for parsing and unmarshaling KDL documents
- [github.com/fsnotify/fsnotify](https://github.com/fsnotify/fsnotify) — for file‑watching (hot reload)

## Installation

```bash
go get github.com/ykhdr/kdl-config
```

## Quickstart
```go
package main

import (
	"fmt"
	"log"

	"github.com/ykhdr/kdl-config"
)

type Config struct {
	Port    int     `validate:"required,min=1,max=65535"`
	Scale   float64 `validate:"min=0.0,max=10.0"`
	Name    string  `validate:"required,len=4"`
	Env     string  `validate:"oneof=dev|prod|test"`
	Pattern string  `validate:"pattern=^user_[0-9]+$"`
}

func main() {
	loader := kdlconfig.NewLoader()
	cfg := &Config{}

	if err := loader.Load(cfg, "config.kdl"); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	fmt.Printf("Loaded configuration: %+v\n", cfg)
}
```

## Custom rules definition

```go
package main

import (
  "fmt"
  "reflect"

  "github.com/ykhdr/kdl-config"
)

// 1) Define a Rule implementation
type IsPositiveRule struct{}

func (r *IsPositiveRule) Name() string { return "ispositive" }
func (r *IsPositiveRule) Validate(fv reflect.Value, _ reflect.StructField) error {  
	switch fv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if fv.Int() <= 0 {
			return fmt.Errorf("must be positive")
		}
    case reflect.Float32, reflect.Float64:
        if fv.Float() <= 0 {
            return fmt.Errorf("must be positive")
        }
    default:
        return fmt.Errorf("unsupported type %s", fv.Kind())
    }
    return nil
}

// 2) Register it before calling Load
func init() {
    kdlconfig.RegisterRule("ispositive", func(_ string) (kdlconfig.Rule, error) {
        return &IsPositiveRule{}, nil
    })
}

// 3) Use in your config struct
type CustomConfig struct {
  Count float64 `validate:"ispositive"`
}
```



## Hot Reload
```go
package main

import (
  "fmt"
  "time"

  "github.com/ykhdr/kdl-config"
)

type Config struct {
  Foo int `kdl:"foo" validate:"min=0"`
}

func main() {
  watcher, err := kdlconfig.Watch("config.kdl", &Config{}, func(newCfg any) {
    cfg := newCfg.(*Config)
    fmt.Printf("Config updated: %+v\n", cfg)
  })
  if err != nil {
    panic(err)
  }
  defer watcher.Stop()

  select {
  case <-time.After(10 * time.Minute):
  }
}
```

## Examples 

See the [examples](./examples) directory for:
-	**basic/** — simple load & validate.
-	**watch/** — hot‑reloading.
