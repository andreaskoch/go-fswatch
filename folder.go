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
	timeStamp     time.Time
	newItems      []string
	movedItems    []string
	modifiedItems []string
}

func newFolderChange(newItems, movedItems, modifiedItems []string) *FolderChange {
	return &FolderChange{
		timeStamp:     time.Now(),
		newItems:      newItems,
		movedItems:    movedItems,
		modifiedItems: modifiedItems,
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
	sleepInterval := time.Second * 2

	go func() {

		// get existing entries
		directory := folderWatcher.folder
		entryList := getFolderEntries(directory, folderWatcher.recurse, folderWatcher.skipFile)

		for folderWatcher.IsRunning() {

			// get new entries
			updatedEntryList := getFolderEntries(directory, folderWatcher.recurse, folderWatcher.skipFile)

			// check for new items
			newItems := make([]string, 0)
			modifiedItems := make([]string, 0)

			for entry, hash := range updatedEntryList {

				// check for new entries
				if _, exists := entryList[entry]; !exists {

					// entry is new
					newItems = append(newItems, entry)
					continue
				}

				// check if the file changed
				if oldHash := entryList[entry]; oldHash != hash {

					// existing entry has been modified
					modifiedItems = append(modifiedItems, entry)

				}
			}

			// check for moved items
			movedItems := make([]string, 0)
			for _, entry := range entryList {
				if _, exists := updatedEntryList[entry]; !exists {
					movedItems = append(movedItems, entry)
				}
			}

			// assign the new list
			entryList = updatedEntryList

			// sleep
			time.Sleep(sleepInterval)

			// check if something happened
			if len(newItems) > 0 || len(movedItems) > 0 || len(modifiedItems) > 0 {

				// send out change
				go func() {
					folderWatcher.Change <- newFolderChange(newItems, movedItems, modifiedItems)
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

func getFolderEntries(directory string, recurse bool, skipFile func(path string) bool) map[string]string {

	// the return array
	entries := make(map[string]string)

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
			for filepath, hash := range subFolderEntries {
				entries[filepath] = hash
			}

		} else {

			// append entry
			if hash, err := getHashFromFile(subEntryPath); err == nil {
				entries[subEntryPath] = hash
			}
		}

	}

	return entries
}
