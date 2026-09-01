package loom

import (
	"fmt"
	"os"
)

const MaxInputBytes int64 = 2 * 1024 * 1024

func readSourceFile(path, label string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > MaxInputBytes {
		return nil, fmt.Errorf("%s exceeds maximum supported input size of %d bytes", label, MaxInputBytes)
	}
	return os.ReadFile(path)
}
