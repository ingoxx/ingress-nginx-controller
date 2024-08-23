package file

import (
	"github.com/fsnotify/fsnotify"
	"log"
)

type WatcherFile struct {
	dir     string
	watcher *fsnotify.Watcher
	onEvent func()
}

func NewFileWatcher(dir string, onEvent func()) (*WatcherFile, error) {
	w := &WatcherFile{
		onEvent: onEvent,
		dir:     dir,
	}

	return w, w.watch()
}

func (w *WatcherFile) watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	w.watcher = watcher

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					w.onEvent()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	return watcher.Add(w.dir)
}
