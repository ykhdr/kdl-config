package kdlconfig

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors changes to a KDL file and automatically reloads it.
type Watcher struct {
	path       string
	prototype  any
	onChange   func(newCfg any)
	loader     *Loader
	currentMux sync.RWMutex
	current    any
	watcher    *fsnotify.Watcher
	stopCh     chan struct{}
}

// Watch creates and starts a Watcher.
//   - path: path to the config file.
//   - prototype: pointer to an empty struct of the same shape that will be loaded.
//   - onChange: callback that is called on the first successful load and after each successful reload.
func Watch(path string, prototype any, onChange func(newCfg any)) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	watcher := &Watcher{
		path:      path,
		prototype: prototype,
		onChange:  onChange,
		loader:    NewLoader(),
		stopCh:    make(chan struct{}),
	}

	// Immediately load the config and call the callback
	if err := watcher.reload(); err != nil {
		_ = w.Close()
		return nil, err
	}

	if err := w.Add(path); err != nil {
		_ = w.Close()
		return nil, fmt.Errorf("failed to watch file %q: %w", path, err)
	}
	watcher.watcher = w

	go watcher.loop()
	return watcher, nil
}

// loop listens for fsnotify events and reloads the config when the file changes.
func (w *Watcher) loop() {
	defer func() { _ = w.watcher.Close() }()
	debounce := time.NewTimer(0)
	<-debounce.C

	for {
		select {
		case event := <-w.watcher.Events:
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				// debouncing rapid events
				debounce.Reset(100 * time.Millisecond)
			}
		case <-debounce.C:
			if err := w.reload(); err != nil {
				// here we can log the error, but not stop the watcher
				fmt.Printf("config reload error: %v\n", err)
			}
		case err := <-w.watcher.Errors:
			fmt.Printf("watcher error: %v\n", err)
		case <-w.stopCh:
			return
		}
	}
}

// reload reads and validates the new config, saves it and calls onChange.
func (w *Watcher) reload() error {
	// Clone the prototype to avoid overwriting the old instance
	newCfg, err := clonePrototype(w.prototype)
	if err != nil {
		return fmt.Errorf("failed to clone prototype: %w", err)
	}

	if err := w.loader.Load(newCfg, w.path); err != nil {
		return err
	}

	w.currentMux.Lock()
	w.current = newCfg
	w.currentMux.Unlock()

	// callback in a separate goroutine to avoid blocking watcher.loop
	go w.onChange(newCfg)
	return nil
}

// Stop stops the watching and frees resources.
func (w *Watcher) Stop() {
	close(w.stopCh)
}

// clonePrototype creates a new pointer to the same structure as the prototype.
// Returns an error if the prototype is not a pointer to a struct.
func clonePrototype(prototype any) (any, error) {
	t := reflect.TypeOf(prototype)
	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("prototype must be a pointer to struct")
	}
	// create a new pointer to the struct
	newPtr := reflect.New(t.Elem())
	return newPtr.Interface(), nil
}
