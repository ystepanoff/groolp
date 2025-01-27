package watcher

import "github.com/fsnotify/fsnotify"

// WatcherInterface defines the methods and channels used by Watcher
type WatcherInterface interface {
	Add(name string) error
	Close() error
	Events() <-chan fsnotify.Event
	Errors() <-chan error
}

// FSNotifyWrapper wraps fsnotify.Watcher to implement WatcherInterface
type FSNotifyWrapper struct {
	*fsnotify.Watcher
}

func (f *FSNotifyWrapper) Events() <-chan fsnotify.Event {
	return f.Watcher.Events
}

func (f *FSNotifyWrapper) Errors() <-chan error {
	return f.Watcher.Errors
}
