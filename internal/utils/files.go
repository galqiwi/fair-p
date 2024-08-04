package utils

import (
	"errors"
	"os"
)

// FileExists is not correct, don't use it in mission-critical applications
// https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
func FileExists(filename string) bool {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
