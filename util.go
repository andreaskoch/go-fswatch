// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fswatch

// sliceContainsElement returns true if the specified list of strings contains the given text element.
func sliceContainsElement(listOfStrings []string, textElement string) bool {
	for _, t := range listOfStrings {
		if t == textElement {
			return true
		}
	}

	return false
}
