// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fswatch

import (
	"fmt"
	"time"
)

// newFolderChange creates a new FolderChange instance from the given list if new, moved and modified items.
func newFolderChange(newItems, movedItems, modifiedItems []string) *FolderChange {
	return &FolderChange{
		timeStamp:     time.Now(),
		newItems:      newItems,
		movedItems:    movedItems,
		modifiedItems: modifiedItems,
	}
}

// FolderChange represents changes (new, moved and modified items) of a folder at a given time.
type FolderChange struct {
	timeStamp     time.Time
	newItems      []string
	movedItems    []string
	modifiedItems []string
}

func (folderChange *FolderChange) String() string {
	return fmt.Sprintf("Folderchange (timestamp: %s, new: %d, moved: %d)", folderChange.timeStamp, len(folderChange.New()), len(folderChange.Moved()))
}

// TimeStamp retunrs the time stamp of the current folder change.
func (folderChange *FolderChange) TimeStamp() time.Time {
	return folderChange.timeStamp
}

// New returns the new items of the current folder change.
func (folderChange *FolderChange) New() []string {
	return folderChange.newItems
}

// Moved returns the moved items of the current folder change.
func (folderChange *FolderChange) Moved() []string {
	return folderChange.movedItems
}

// Modified returns the modified items of the current folder change.
func (folderChange *FolderChange) Modified() []string {
	return folderChange.modifiedItems
}
