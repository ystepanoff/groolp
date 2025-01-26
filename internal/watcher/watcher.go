package watcher

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/ystepanoff/groolp/internal/core"
)

type Watcher struct {
	watcher     *fsnotify.Watcher
	taskManager *core.TaskManager
	watchPaths  []string
	taskName    string
}

func NewWatcher(
	tm *core.TaskManager,
	paths []string,
	taskName string,
) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		if err := w.Add(path); err != nil {
			return nil, err
		}
	}

	return &Watcher{
		watcher:     w,
		taskManager: tm,
		watchPaths:  paths,
		taskName:    taskName,
	}, nil
}

func (w *Watcher) Start() {
	defer w.watcher.Close()

	log.Println("Starting file watcher...")
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			for _, op := range []fsnotify.Op{
				fsnotify.Create,
				fsnotify.Remove,
				fsnotify.Write,
				fsnotify.Chmod,
				fsnotify.Rename,
			} {
				if event.Op&op == op {
					log.Printf("Detected change in: %s\n", event.Name)
					if err := w.taskManager.Run(w.taskName); err != nil {
						log.Printf(
							"Error running task '%s': %v\n",
							w.taskName,
							err,
						)
					}
					break
				}
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}
