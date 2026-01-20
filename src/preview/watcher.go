package preview

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/hinkolas/mdoc/src/core"
)

type Watcher struct {
	watcher *fsnotify.Watcher
}

func NewWatcher(document *core.Document) (*Watcher, error) {

	// Set up file watcher for hot reload
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()

	// Watch the source file
	absInputPath, err := filepath.Abs(document.DocumentPath)
	if err != nil {
		fmt.Println("Failed to get absolute path for input file:", err)
		os.Exit(1)
	}
	if err := watcher.Add(absInputPath); err != nil {
		fmt.Println("Failed to watch input file:", err)
		os.Exit(1)
	}

	// Watch the theme file
	themePath := filepath.Join(os.ExpandEnv(core.THEME_DIR), document.Config.Theme+".html")
	absThemePath, err := filepath.Abs(themePath)
	if err != nil {
		fmt.Println("Failed to get absolute path for theme file:", err)
		os.Exit(1)
	}
	if err := watcher.Add(absThemePath); err != nil {
		fmt.Printf("Warning: Failed to watch theme file %s: %v\n", absThemePath, err)
		// Continue anyway - theme watching is optional
	}

	return &Watcher{
		watcher: watcher,
	}, nil

}
