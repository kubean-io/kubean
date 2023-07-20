// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"fmt"
	"runtime"
	"testing"
)

func TestGet(t *testing.T) {
	infoObj := Get()
	infoObj.GoVersion = runtime.Version()
	if infoObj.GoVersion != runtime.Version() {
		t.Fatal()
	}
	fmt.Println(infoObj.String())
}
