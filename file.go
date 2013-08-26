// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fswatch

import (
	"fmt"
	"os"
	"time"
)

type FileWatcher struct {
	Modified chan bool
	Moved    chan bool
	Stopped  chan bool

	hash    string
	debug   bool
	file    string
	running bool
}

func NewFileWatcher(filePath string) *FileWatcher {
	hash, err := getHashFromFile(filePath)
	if err != nil {
		return nil
	}

	return &FileWatcher{
		Modified: make(chan bool),
		Moved:    make(chan bool),
		Stopped:  make(chan bool),
		hash:     hash,
		debug:    false,
		file:     filePath,
	}
}

func (fileWatcher *FileWatcher) String() string {
	return fmt.Sprintf("Filewatcher %q", fileWatcher.file)
}

func (fileWatcher *FileWatcher) SetFile(filePath string) {
	fileWatcher.file = filePath
}

func (fileWatcher *FileWatcher) Start() *FileWatcher {
	fileWatcher.running = true
	sleepInterval := time.Second * 2

	go func() {

		for fileWatcher.running {

			if newHash, err := getHashFromFile(fileWatcher.file); err == nil {

				// check if file has been modified
				if newHash != fileWatcher.hash {

					fileWatcher.log("Item was modified")

					// save the new hash
					fileWatcher.hash = newHash

					// send out the notification
					go func() {
						fileWatcher.Modified <- true
					}()
				}

			} else if os.IsNotExist(err) {

				// send out the notification
				fileWatcher.log("Item was removed")
				go func() {
					fileWatcher.Moved <- true
				}()

				// stop this file watcher
				fileWatcher.Stop()
			}

			time.Sleep(sleepInterval)

		}

		go func() {
			fileWatcher.Stopped <- true
		}()

		fileWatcher.log("Stopped")
	}()

	return fileWatcher
}

func (fileWatcher *FileWatcher) Stop() *FileWatcher {
	fileWatcher.log("Stopping")
	fileWatcher.running = false
	return fileWatcher
}

func (fileWatcher *FileWatcher) IsRunning() bool {
	return fileWatcher.running
}

func (fileWatcher *FileWatcher) log(message string) *FileWatcher {
	if fileWatcher.debug {
		fmt.Printf("%s - %s\n", fileWatcher, message)
	}

	return fileWatcher
}
