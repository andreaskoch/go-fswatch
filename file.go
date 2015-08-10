// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fswatch

import (
	"fmt"
	"os"
	"time"
)

// numberOfFileWatchers contains the current number of active file watchers.
var numberOfFileWatchers int

func init() {
	numberOfFolderWatchers = 0
}

// NumberOfFileWatchers returns the number of currently active file watchers.
func NumberOfFileWatchers() int {
	return numberOfFileWatchers
}

// A FileWatcher can be used to determine if a given file has been modified or moved.
type FileWatcher struct {
	modified chan bool
	moved    chan bool
	stopped  chan bool

	file          string
	running       bool
	wasStopped    bool
	checkInterval time.Duration

	previousModTime time.Time
}

// NewFileWatcher creates a new file watcher for a given file path.
// The check interval in seconds defines how often the watcher shall check for changes  (recommended: 1 - n seconds).
func NewFileWatcher(filePath string, checkIntervalInSeconds int) *FileWatcher {

	if checkIntervalInSeconds < 1 {
		panic(fmt.Sprintf("Cannot create a file watcher with a check interval of %v seconds.", checkIntervalInSeconds))
	}

	return &FileWatcher{
		modified: make(chan bool),
		moved:    make(chan bool),
		stopped:  make(chan bool),

		file:          filePath,
		checkInterval: time.Duration(checkIntervalInSeconds),
	}
}

func (fileWatcher *FileWatcher) String() string {
	return fmt.Sprintf("Filewatcher %q", fileWatcher.file)
}

// SetFile sets the file file for this file watcher.
func (fileWatcher *FileWatcher) SetFile(filePath string) {
	fileWatcher.file = filePath
}

// Modified returns a channel indicating if the file has been modified.
func (filewatcher *FileWatcher) Modified() chan bool {
	return filewatcher.modified
}

// Moved returns a channel indicating if the file has been moved.
func (filewatcher *FileWatcher) Moved() chan bool {
	return filewatcher.moved
}

// Stopped returns a channel indicating if the file watcher stopped.
func (filewatcher *FileWatcher) Stopped() chan bool {
	return filewatcher.stopped
}

// Start starts the watch process.
func (fileWatcher *FileWatcher) Start() {
	fileWatcher.running = true
	sleepInterval := time.Second * fileWatcher.checkInterval

	go func() {

		// increment watcher count
		numberOfFileWatchers++

		var modTime time.Time
		previousModTime := fileWatcher.getPreviousModTime()

		if timeIsSet(previousModTime) {
			modTime = previousModTime
		} else {
			currentModTime, err := getLastModTimeFromFile(fileWatcher.file)
			if err != nil {

				// send out the notification
				log("File %q has been moved or is inaccessible.", fileWatcher.file)
				go func() {
					fileWatcher.moved <- true
				}()

				// stop this file watcher
				fileWatcher.Stop()

			} else {

				modTime = currentModTime
			}

		}

		for fileWatcher.wasStopped == false {

			newModTime, err := getLastModTimeFromFile(fileWatcher.file)
			if err != nil {

				// send out the notification
				log("File %q has been moved.", fileWatcher.file)
				go func() {
					fileWatcher.moved <- true
				}()

				// stop this file watcher
				fileWatcher.Stop()

				continue
			}

			// detect changes
			if modTime.Before(newModTime) {

				// send out the notification
				log("File %q has been modified.", fileWatcher.file)
				go func() {
					fileWatcher.modified <- true
				}()

			} else {

				log("File %q has not changed.", fileWatcher.file)

			}

			// assign the new modtime
			modTime = newModTime

			time.Sleep(sleepInterval)

		}

		fileWatcher.running = false

		// capture the entry list for a restart
		fileWatcher.captureModTime(modTime)

		// inform channel-subscribers
		go func() {
			fileWatcher.stopped <- true
		}()

		// decrement the watch counter
		numberOfFileWatchers--

		// final log message
		log("Stopped file watcher %q", fileWatcher.String())
	}()
}

// Stop stops the watch process.
func (fileWatcher *FileWatcher) Stop() {
	log("Stopping file watcher %q", fileWatcher.String())
	fileWatcher.wasStopped = true
}

// IsRunning returns a flag indicating whether the watcher is currently running.
func (fileWatcher *FileWatcher) IsRunning() bool {
	return fileWatcher.running
}

// getPreviousModTime returns the last known modification time of the file.
func (fileWatcher *FileWatcher) getPreviousModTime() time.Time {
	return fileWatcher.previousModTime
}

// Remember the last mod time for a later restart
func (fileWatcher *FileWatcher) captureModTime(modTime time.Time) {
	fileWatcher.previousModTime = modTime
}

// getLastModTimeFromFile returns the last modification time of the file with the given file path.
// If modifiction time cannot be determined getLastModTimeFromFile will return an error.
func getLastModTimeFromFile(file string) (time.Time, error) {
	fileInfo, err := os.Stat(file)
	if err != nil {
		return time.Time{}, err
	}

	return fileInfo.ModTime(), nil
}

// timeIsSet returns true if the supplied time is set / initialized.
func timeIsSet(t time.Time) bool {
	return time.Time{} == t
}
