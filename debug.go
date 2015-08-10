// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fswatch

import (
	"fmt"
)

var (
	debugIsEnabled = false
	debugMessages  chan string
)

// EnableDebug enables the debug mode for this package and returns a
// debug message channel.
func EnableDebug() chan string {
	debugIsEnabled = true
	debugMessages = make(chan string, 10)
	return debugMessages
}

// DisableDebug disables the debug mode for this package
// and closes the debug message channel.
func DisableDebug() {
	debugIsEnabled = false
	close(debugMessages)
}

// log sends a log message down the debug message channel if
// debugging is enabled.
func log(format string, v ...interface{}) {
	if !debugIsEnabled {
		return
	}

	debugMessages <- fmt.Sprint(fmt.Sprintf(format, v...))
}
