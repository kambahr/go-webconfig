package webconfig

import (
	"os"
	"strings"
)

// trimLine take out tab and spaces from both end of a linc.
func (c *Config) trimLine(l string) string {

	l = strings.Replace(l, "\t", " ", -1)
	l = strings.TrimLeft(l, " ")
	l = strings.TrimRight(l, " ")
	l = strings.Trim(l, " ")

	return l
}

// fileOrDirExists checks existance of file or directory.
func (c *Config) fileOrDirExists(path string) bool {
	if path == "" {
		return false
	}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return true
}
