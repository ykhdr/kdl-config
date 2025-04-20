package kdlconfig

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// watcherConfig is the struct used for KDL unmarshaling in tests.
type watcherConfig struct {
	Foo int `kdl:"foo"`
}

func TestWatcher_Reload(t *testing.T) {
	tempDir := t.TempDir()
	file := filepath.Join(tempDir, "config.kdl")

	// Write initial valid config
	require.NoError(t, os.WriteFile(file, []byte("foo 1\n"), 0644))

	// Channel to capture callbacks
	ch := make(chan watcherConfig, 2)

	watcher, err := Watch(file, &watcherConfig{}, func(newCfg any) {
		cfg := newCfg.(*watcherConfig)
		ch <- *cfg
	})
	require.NoError(t, err)
	defer watcher.Stop()

	// Expect initial load with Foo == 1
	require.Eventually(t, func() bool {
		select {
		case c := <-ch:
			return c.Foo == 1
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond, "initial config not loaded")

	// Update config
	require.NoError(t, os.WriteFile(file, []byte("foo 2\n"), 0644))

	// Expect reload with Foo == 2
	require.Eventually(t, func() bool {
		select {
		case c := <-ch:
			return c.Foo == 2
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond, "updated config not reloaded")
}

func TestWatcher_Debounce(t *testing.T) {
	tempDir := t.TempDir()
	file := filepath.Join(tempDir, "config.kdl")
	require.NoError(t, os.WriteFile(file, []byte("foo 3\n"), 0644))

	ch := make(chan watcherConfig, 1)
	watcher, err := Watch(file, &watcherConfig{}, func(newCfg any) {
		cfg := newCfg.(*watcherConfig)
		ch <- *cfg
	})
	require.NoError(t, err)
	defer watcher.Stop()

	// Drain initial event
	require.Eventually(t, func() bool {
		select {
		case <-ch:
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)

	// Rapid updates: write foo=4 then foo=5
	require.NoError(t, os.WriteFile(file, []byte("foo 4\n"), 0644))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, os.WriteFile(file, []byte("foo 5\n"), 0644))

	// Expect only one reload with the last value (5)
	require.Eventually(t, func() bool {
		select {
		case c := <-ch:
			return c.Foo == 5
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}

func TestWatcher_InvalidConfigNoCallback(t *testing.T) {
	tempDir := t.TempDir()
	file := filepath.Join(tempDir, "config.kdl")
	require.NoError(t, os.WriteFile(file, []byte("foo 6\n"), 0644))

	ch := make(chan watcherConfig, 2)
	watcher, err := Watch(file, &watcherConfig{}, func(newCfg any) {
		cfg := newCfg.(*watcherConfig)
		ch <- *cfg
	})
	require.NoError(t, err)
	defer watcher.Stop()

	// Initial load
	require.Eventually(t, func() bool {
		select {
		case c := <-ch:
			return c.Foo == 6
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)

	// Write invalid content (should not trigger callback)
	require.NoError(t, os.WriteFile(file, []byte("invalid content\n"), 0644))

	// Wait and ensure no new callback
	time.Sleep(200 * time.Millisecond)
	select {
	case c := <-ch:
		t.Fatalf("unexpected callback for invalid config: %+v", c)
	default:
	}
}

func TestWatcher_Stop(t *testing.T) {
	tempDir := t.TempDir()
	file := filepath.Join(tempDir, "config.kdl")
	require.NoError(t, os.WriteFile(file, []byte("foo 7\n"), 0644))

	ch := make(chan watcherConfig, 2)
	watcher, err := Watch(file, &watcherConfig{}, func(newCfg any) {
		cfg := newCfg.(*watcherConfig)
		ch <- *cfg
	})
	require.NoError(t, err)

	// Initial load
	require.Eventually(t, func() bool {
		select {
		case <-ch:
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)

	// Stop the watcher
	watcher.Stop()

	// Write new valid config (should not trigger callback)
	require.NoError(t, os.WriteFile(file, []byte("foo 8\n"), 0644))
	time.Sleep(200 * time.Millisecond)
	select {
	case c := <-ch:
		t.Fatalf("unexpected callback after Stop: %+v", c)
	default:
	}
}
