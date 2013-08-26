// Copyright 2013 Andreas Koch. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fswatch

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
)

func getHashFromFile(path string) (string, error) {

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	sha1 := sha1.New()
	sha1.Write(bytes)

	return fmt.Sprintf("%x", string(sha1.Sum(nil)[0:8])), nil
}
