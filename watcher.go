// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fswatch

// Watcher defines the base set of of functions for a file and folder watcher.
type Watcher interface {

	// Modified returns a boolean channel which sends a flag indicating
	// whether a filesystem item has been modified.
	Modified() chan bool

	// Moved returns a boolean channel which sends a flag indicating
	// whether a filesystem item has moved.
	Moved() chan bool

	// Stopped returns a boolean channel which sends a flag indicating
	// whether the filesystem watcher has stopped.
	Stopped() chan bool

	// Start starts the watch-process.
	Start()

	// Stop stops any the watch-process.
	Stop()

	// IsRunning returns a flag indicating whether this
	// filesystem watcher is running or not.
	IsRunning() bool
}
