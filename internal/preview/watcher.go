package preview

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches a set of files and invokes onChange (debounced) when any
// of them is written. onChange receives the path of the file whose change
// triggered the (debounced) fire, so callers can log what reloaded.
type Watcher struct {
	w        *fsnotify.Watcher
	onChange func(changed string)

	mu        sync.Mutex
	themePath string // the single active theme file, updated via WatchTheme
}

// NewWatcher creates a watcher and adds each given path. A path that doesn't
// exist (e.g. a themes/ directory the project never created) is skipped
// silently — it's a normal, expected case, not an error. Other Add failures
// are logged but non-fatal.
func NewWatcher(onChange func(changed string), paths ...string) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	for _, p := range paths {
		if p == "" {
			continue
		}
		if _, statErr := os.Stat(p); errors.Is(statErr, fs.ErrNotExist) {
			continue
		}
		if err := w.Add(p); err != nil {
			log.Printf("watch %s: %v", p, err)
		}
	}
	return &Watcher{w: w, onChange: onChange}, nil
}

// WatchTheme follows the document's active theme file for content edits.
// The active theme can change mid-session (the user edits the frontmatter, or
// fixes a previously-missing theme), so this is called on every reload to keep
// the watch pointed at the right file. Passing "" — which is what the built-in
// default theme resolves to — just drops the previous theme watch.
//
// Theme directories are watched separately and statically (see NewWatcher) so
// that creating or switching theme files is noticed even on platforms where a
// directory watch doesn't report writes to existing files.
func (w *Watcher) WatchTheme(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if path == w.themePath {
		return
	}
	if w.themePath != "" {
		_ = w.w.Remove(w.themePath)
	}
	if path != "" {
		if err := w.w.Add(path); err != nil {
			log.Printf("watch theme %s: %v", path, err)
			w.themePath = ""
			return
		}
	}
	w.themePath = path
}

// Run blocks until Close is called. A 100ms debounce coalesces editor
// save bursts (some editors emit several writes in quick succession).
func (w *Watcher) Run() {
	const debounce = 100 * time.Millisecond
	var pending *time.Timer
	var lastPath string // the file from the most recent qualifying event
	fire := func() {
		if w.onChange != nil {
			w.onChange(lastPath)
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
			lastPath = ev.Name
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
