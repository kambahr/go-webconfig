package webconfig

import (
	"io"
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
func fileOrDirExists(path string) bool {
	if path == "" {
		return false
	}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

// The config file is opened for read every few seconds... the original
// Go ReadFile() func closes the file on defer; since the refreshConfig()
// loops inside itself via goto, the file.Close() is never called,
// hence the too may open file error. The following is the same
// Go ReadFile() func with the exception of closing the file before
// return.
//
// ../src/os/file.go
// ReadFile reads the named file and returns the contents.
// A successful call returns err == nil, not err == EOF.
// Because ReadFile reads the whole file, it does not treat an EOF from Read
// as an error to be reported.
func ReadFile(name string) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		f.Close()
		return nil, err
	}

	var size int
	if info, err := f.Stat(); err == nil {
		size64 := info.Size()
		if int64(int(size64)) == size64 {
			size = int(size64)
		}
	}
	size++ // one byte for final read at EOF

	// If a file claims a small size, read at least 512 bytes.
	// In particular, files in Linux's /proc claim size 0 but
	// then do not work right if read in small pieces,
	// so an initial read of 1 byte would not work correctly.
	if size < 512 {
		size = 512
	}

	data := make([]byte, 0, size)
	for {
		if len(data) >= cap(data) {
			d := append(data[:cap(data)], 0)
			data = d[:len(data)]
		}
		n, err := f.Read(data[len(data):cap(data)])
		data = data[:len(data)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			f.Close()
			return data, err
		}
	}
}
