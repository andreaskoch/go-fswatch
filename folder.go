// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fswatch

import (
	"fmt"
	"io/ioutil"
	"os"
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

func (folderChange *FolderChange) Modified() []string {
	return folderChange.modifiedItems
}

type FolderWatcher struct {
	Change  chan *FolderChange
	Stopped chan bool

	recurse  bool
	skipFile func(path string) bool

	debug         bool
	folder        string
	running       bool
	checkInterval time.Duration
}

func NewFolderWatcher(folderPath string, recurse bool, skipFile func(path string) bool, checkIntervalInSeconds int) *FolderWatcher {

	if checkIntervalInSeconds < 1 {
		panic(fmt.Sprintf("Cannot create a folder watcher with a check interval of %v seconds.", checkIntervalInSeconds))
	}

	return &FolderWatcher{
		Change:  make(chan *FolderChange),
		Stopped: make(chan bool),

		recurse:  recurse,
		skipFile: skipFile,

		debug:         false,
		folder:        folderPath,
		checkInterval: time.Duration(checkIntervalInSeconds),
	}
}

func (folderWatcher *FolderWatcher) String() string {
	return fmt.Sprintf("Folderwatcher %q", folderWatcher.folder)
}

func (folderWatcher *FolderWatcher) Start() *FolderWatcher {
	folderWatcher.running = true
	sleepInterval := time.Second * folderWatcher.checkInterval

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

			for _, entry := range updatedEntryList {

				if isNewItem := !sliceContainsElement(entryList, entry); isNewItem {
					// entry is new
					newItems = append(newItems, entry)
					continue
				}

				// check if the file changed
				if fileInfo, err := os.Stat(entry); err == nil {

					// check if file has been modified
					timeOfLastCheck := time.Now().Add(sleepInterval * -1)
					if fileHasChanged(fileInfo, timeOfLastCheck) {

						// existing entry has been modified
						modifiedItems = append(modifiedItems, entry)
					}

				}
			}

			// check for moved items
			movedItems := make([]string, 0)
			for _, entry := range entryList {
				isMoved := !sliceContainsElement(updatedEntryList, entry)
				if isMoved {
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

		// recurse or append
		if recurse && entry.IsDir() {

			// recurse
			subFolderEntries := getFolderEntries(subEntryPath, recurse, skipFile)
			entries = append(entries, subFolderEntries...)

		} else {

			// check if the enty shall be ignored
			if skipFile(subEntryPath) {
				continue
			}

			// append entry
			entries = append(entries, subEntryPath)
		}

	}

	return entries
}
