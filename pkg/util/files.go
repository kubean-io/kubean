// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"
)

func WriteFile(filepath string, bytes []byte) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func IsExist(filepath string) bool {
	if _, err := os.Stat(filepath); err == nil {
		return true
	}
	return false
}
