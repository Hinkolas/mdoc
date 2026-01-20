package preview

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher   *fsnotify.Watcher
	OnChanged func()
}

// NewWatcher creates a new file watcher for the given paths
func NewWatcher(paths ...string) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			watcher.Close()
			return nil, fmt.Errorf("failed to get absolute path for %s: %w", path, err)
		}

		// Check if file exists
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			log.Printf("Warning: file does not exist, skipping watch: %s", absPath)
			continue
		}

		if err := watcher.Add(absPath); err != nil {
			log.Printf("Warning: failed to watch %s: %v", absPath, err)
			continue
		}
		log.Printf("Watching: %s", absPath)
	}

	return &Watcher{
		watcher: watcher,
	}, nil
}

// Watch starts watching for file changes and calls OnChanged when detected
// This method blocks until Close is called
func (w *Watcher) Watch() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			// Only react to write events
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Printf("File changed: %s", event.Name)
				if w.OnChanged != nil {
					w.OnChanged()
				}
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

// Close stops the watcher
func (w *Watcher) Close() error {
	return w.watcher.Close()
}
