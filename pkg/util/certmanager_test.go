// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
	"time"
)

func TestNewCertManager(t *testing.T) {
	cert, key, err := NewCertManager(
		[]string{"Org1"},
		time.Hour*100,
		[]string{"a1.com", "a2.com"},
		"a1.com",
	).GenerateSelfSignedCerts()
	if err != nil {
		t.Fatal()
	}
	if cert.String() == "" {
		t.Fatal()
	}
	if key.String() == "" {
		t.Fatal()
	}
}
