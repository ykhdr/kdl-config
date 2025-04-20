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

	// Load and validate config.kdl in current directory
	if err := loader.Load(cfg, "config.kdl"); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	fmt.Printf("Loaded configuration: %+v\n", cfg)
}
