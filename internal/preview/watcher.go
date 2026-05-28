package preview

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches a set of files and invokes onChange (debounced) when any
// of them is written.
type Watcher struct {
	w        *fsnotify.Watcher
	onChange func()
}

// NewWatcher creates a watcher and adds each given path. Missing paths are
// skipped with a warning rather than treated as fatal — themes may be
// inlined into the document, in which case there's nothing to watch.
func NewWatcher(onChange func(), paths ...string) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	for _, p := range paths {
		if p == "" {
			continue
		}
		if err := w.Add(p); err != nil {
			log.Printf("watch %s: %v", p, err)
		}
	}
	return &Watcher{w: w, onChange: onChange}, nil
}

// Run blocks until Close is called. A 100ms debounce coalesces editor
// save bursts (some editors emit several writes in quick succession).
func (w *Watcher) Run() {
	const debounce = 100 * time.Millisecond
	var pending *time.Timer
	fire := func() {
		if w.onChange != nil {
			w.onChange()
		}
	}
	for {
		select {
		case ev, ok := <-w.w.Events:
			if !ok {
				return
			}
			if ev.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}
			if pending != nil {
				pending.Stop()
			}
			pending = time.AfterFunc(debounce, fire)
		case err, ok := <-w.w.Errors:
			if !ok {
				return
			}
			log.Printf("watcher: %v", err)
		}
	}
}

// Close stops the watcher.
func (w *Watcher) Close() error { return w.w.Close() }
