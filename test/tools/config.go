package tools

import (
	//"fmt"
	//"io/ioutil"
	//"log"
	//"os"
	"path/filepath"
	"runtime"
	//"gopkg.in/yaml.v2"
)

var _, currentFile, _, _ = runtime.Caller(0)
var basepath = filepath.Dir(currentFile)

func Path(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(basepath, rel)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
