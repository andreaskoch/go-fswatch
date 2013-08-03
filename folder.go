// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fswatch

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"
)

type FolderChange struct {
	timeStamp  time.Time
	newItems   []string
	movedItems []string
}

func newFolderChange(newItems []string, movedItems []string) *FolderChange {
	return &FolderChange{
		timeStamp:  time.Now(),
		newItems:   newItems,
		movedItems: movedItems,
	}
}

func (folderChange *FolderChange) String() string {
	return fmt.Sprintf("Folderchange (timestamp: %s, new: %d, moved: %d)", folderChange.timeStamp, len(folderChange.New()), len(folderChange.Moved()))
}

func (folderChange *FolderChange) TimeStamp() time.Time {
	return folderChange.timeStamp
}

func (folderChange *FolderChange) New() []string {
	return folderChange.newItems
}

func (folderChange *FolderChange) Moved() []string {
	return folderChange.movedItems
}

type FolderWatcher struct {
	Change  chan *FolderChange
	Stopped chan bool

	recurse  bool
	skipFile func(path string) bool

	debug   bool
	folder  string
	running bool
}

func NewFolderWatcher(folderPath string, recurse bool, skipFile func(path string) bool) *FolderWatcher {
	return &FolderWatcher{
		Change:  make(chan *FolderChange),
		Stopped: make(chan bool),

		recurse:  recurse,
		skipFile: skipFile,

		debug:  false,
		folder: folderPath,
	}
}

func (folderWatcher *FolderWatcher) String() string {
	return fmt.Sprintf("Folderwatcher %q", folderWatcher.folder)
}

func (folderWatcher *FolderWatcher) Start() *FolderWatcher {
	folderWatcher.running = true
	sleepTime := time.Second * 2

	go func() {

		// get existing entries
		directory := folderWatcher.folder
		existingEntries := getFolderEntries(directory, folderWatcher.recurse, folderWatcher.skipFile)

		for folderWatcher.IsRunning() {

			// get new entries
			newEntries := getFolderEntries(directory, folderWatcher.recurse, folderWatcher.skipFile)

			// check for new items
			newItems := make([]string, 0)
			for _, newEntry := range newEntries {
				isNewItem := !sliceContainsElement(existingEntries, newEntry)
				if isNewItem {
					newItems = append(newItems, newEntry)
				}
			}

			// check for moved items
			movedItems := make([]string, 0)
			for _, existingEntry := range existingEntries {
				isMoved := !sliceContainsElement(newEntries, existingEntry)
				if isMoved {
					movedItems = append(movedItems, existingEntry)
				}
			}

			// assign the new list
			existingEntries = newEntries

			// sleep
			time.Sleep(sleepTime)

			// check if something happened
			if len(newItems) > 0 || len(movedItems) > 0 {

				// send out change
				go func() {
					folderWatcher.Change <- newFolderChange(newItems, movedItems)
				}()
			}

		}

		go func() {
			folderWatcher.Stopped <- true
		}()

		folderWatcher.log("Stopped")
	}()

	return folderWatcher
}

func (folderWatcher *FolderWatcher) Stop() *FolderWatcher {
	folderWatcher.log("Stopping")
	folderWatcher.running = false
	return folderWatcher
}

func (folderWatcher *FolderWatcher) IsRunning() bool {
	return folderWatcher.running
}

func (folderWatcher *FolderWatcher) log(message string) *FolderWatcher {
	if folderWatcher.debug {
		fmt.Printf("%s - %s\n", folderWatcher, message)
	}

	return folderWatcher
}

func getFolderEntries(directory string, recurse bool, skipFile func(path string) bool) []string {

	// the return array
	entries := make([]string, 0)

	// read the entries of the specified directory
	directoryEntries, err := ioutil.ReadDir(directory)
	if err != nil {
		return entries
	}

	for _, entry := range directoryEntries {

		// get the full path
		subEntryPath := filepath.Join(directory, entry.Name())

		// check if the enty shall be ignored
		if skipFile(subEntryPath) {
			continue
		}

		// recurse or append
		if recurse && entry.IsDir() {

			// recurse
			subFolderEntries := getFolderEntries(subEntryPath, recurse, skipFile)
			entries = append(entries, subFolderEntries...)

		} else {

			// append entry
			entries = append(entries, subEntryPath)
		}

	}

	return entries
}
