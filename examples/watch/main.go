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
	// Start watching config.kdl; callback fires on initial load and on each valid change
	watcher, err := kdlconfig.Watch("config.kdl", &Config{}, func(newCfg any) {
		cfg := newCfg.(*Config)
		fmt.Printf("Config updated: %+v\n", cfg)
	})
	if err != nil {
		panic(err)
	}
	defer watcher.Stop()

	// Keep the program running
	select {
	case <-time.After(10 * time.Minute):
	}
}
