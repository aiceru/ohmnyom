package path

import (
	"path/filepath"
	"runtime"
)

func Root() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "..", "..")
}
