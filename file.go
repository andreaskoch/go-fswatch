// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fswatch

import (
	"fmt"
	"os"
	"time"
)

var numberOfFileWatchers int

func init() {
	numberOfFolderWatchers = 0
}

func NumberOfFileWatchers() int {
	return numberOfFileWatchers
}

type FileWatcher struct {
	modified chan bool
	moved    chan bool
	stopped  chan bool

	debug         bool
	file          string
	running       bool
	checkInterval time.Duration
}

func NewFileWatcher(filePath string, checkIntervalInSeconds int) *FileWatcher {

	if checkIntervalInSeconds < 1 {
		panic(fmt.Sprintf("Cannot create a file watcher with a check interval of %v seconds.", checkIntervalInSeconds))
	}

	return &FileWatcher{
		modified: make(chan bool),
		moved:    make(chan bool),
		stopped:  make(chan bool),

		debug:         false,
		file:          filePath,
		checkInterval: time.Duration(checkIntervalInSeconds),
	}
}

func (fileWatcher *FileWatcher) String() string {
	return fmt.Sprintf("Filewatcher %q", fileWatcher.file)
}

func (fileWatcher *FileWatcher) SetFile(filePath string) {
	fileWatcher.file = filePath
}

func (filewatcher *FileWatcher) Modified() chan bool {
	return filewatcher.modified
}

func (filewatcher *FileWatcher) Moved() chan bool {
	return filewatcher.moved
}

func (filewatcher *FileWatcher) Stopped() chan bool {
	return filewatcher.stopped
}

func (fileWatcher *FileWatcher) Start() {
	fileWatcher.running = true
	sleepInterval := time.Second * fileWatcher.checkInterval

	go func() {

		numberOfFileWatchers++

		for fileWatcher.running {

			if fileInfo, err := os.Stat(fileWatcher.file); err == nil {

				// check if file has been modified
				timeOfLastCheck := time.Now().Add(sleepInterval * -1)
				if fileHasChanged(fileInfo, timeOfLastCheck) {

					// send out the notification
					fileWatcher.log("Item was modified")
					go func() {
						fileWatcher.modified <- true
					}()
				}

			} else if os.IsNotExist(err) {

				// send out the notification
				fileWatcher.log("Item was removed")
				go func() {
					fileWatcher.moved <- true
				}()

				// stop this file watcher
				fileWatcher.Stop()
			}

			time.Sleep(sleepInterval)

		}

		go func() {
			fileWatcher.stopped <- true
		}()

		numberOfFileWatchers--
		fileWatcher.log("Stopped")
	}()
}

func (fileWatcher *FileWatcher) Stop() {
	fileWatcher.log("Stopping")
	fileWatcher.running = false
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

func fileHasChanged(fileInfo os.FileInfo, lastCheckTime time.Time) bool {
	modTime := fileInfo.ModTime()
	if lastCheckTime.Before(modTime) {
		return true
	}

	return false
}
