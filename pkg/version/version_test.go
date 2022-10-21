package version

import (
	"runtime"
	"testing"
)

func TestGet(t *testing.T) {
	infoObj := Get()
	infoObj.GoVersion = runtime.Version()
	if infoObj.GoVersion != runtime.Version() {
		t.Fatal()
	}
}
