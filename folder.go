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

var numberOfFolderWatchers int

func init() {
	numberOfFolderWatchers = 0
}

func NumberOfFolderWatchers() int {
	return numberOfFolderWatchers
}

type FolderWatcher struct {
	changeDetails chan *FolderChange

	modified chan bool
	moved    chan bool
	stopped  chan bool

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

		modified: make(chan bool),
		moved:    make(chan bool),
		stopped:  make(chan bool),

		changeDetails: make(chan *FolderChange),

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

func (folderWatcher *FolderWatcher) Modified() chan bool {
	return folderWatcher.modified
}

func (folderWatcher *FolderWatcher) Moved() chan bool {
	return folderWatcher.moved
}

func (folderWatcher *FolderWatcher) Stopped() chan bool {
	return folderWatcher.stopped
}

func (folderWatcher *FolderWatcher) ChangeDetails() chan *FolderChange {
	return folderWatcher.changeDetails
}

func (folderWatcher *FolderWatcher) Start() {
	folderWatcher.running = true
	sleepInterval := time.Second * folderWatcher.checkInterval

	go func() {

		// get existing entries
		directory := folderWatcher.folder
		entryList, err := getFolderEntries(directory, folderWatcher.recurse, folderWatcher.skipFile)
		if err != nil {
			// the folder does no longer exist or is not accessible
			go func() {
				folderWatcher.moved <- true
			}()
		}

		numberOfFolderWatchers++

		for folderWatcher.IsRunning() {

			// get new entries
			updatedEntryList, _ := getFolderEntries(directory, folderWatcher.recurse, folderWatcher.skipFile)

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
					folderWatcher.modified <- true
					folderWatcher.changeDetails <- newFolderChange(newItems, movedItems, modifiedItems)
				}()
			}
		}

		go func() {
			folderWatcher.stopped <- true
		}()

		numberOfFolderWatchers--
		folderWatcher.log("Stopped")
	}()
}

func (folderWatcher *FolderWatcher) Stop() {
	folderWatcher.log("Stopping")
	folderWatcher.running = false
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

func getFolderEntries(directory string, recurse bool, skipFile func(path string) bool) ([]string, error) {

	// the return array
	entries := make([]string, 0)

	// read the entries of the specified directory
	directoryEntries, err := ioutil.ReadDir(directory)
	if err != nil {
		return entries, err
	}

	for _, entry := range directoryEntries {

		// get the full path
		subEntryPath := filepath.Join(directory, entry.Name())

		// recurse or append
		if recurse && entry.IsDir() {

			// recurse (ignore errors, unreadable sub directories don't hurt much)
			subFolderEntries, _ := getFolderEntries(subEntryPath, recurse, skipFile)
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

	return entries, nil
}
