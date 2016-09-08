/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package template

import (
	"log"
	"path"

	"gopkg.in/fsnotify.v1"
)

type fileWatcher struct {
	file    string
	watcher *fsnotify.Watcher
	onEvent func()
}

func newFileWatcher(file string, onEvent func()) (fileWatcher, error) {
	fw := fileWatcher{
		file:    file,
		onEvent: onEvent,
	}

	err := fw.watch()
	return fw, err
}

func (f fileWatcher) close() error {
	return f.watcher.Close()
}

// watch creates a fsnotify watcher for a file and create of write events
func (f *fileWatcher) watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	f.watcher = watcher

	dir, file := path.Split(f.file)
	go func(file string) {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create &&
						event.Name == file {
					f.onEvent()
				}
			case err := <-watcher.Errors:
				if err != nil {
					log.Printf("error watching file: %v\n", err)
				}
			}
		}
	}(file)
	return watcher.Add(dir)
}
