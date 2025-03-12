package analyze

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type FileSourceHandler struct {
	sourceFileName string
	fileNames      []string
	index          int
}

func newFileSourceHandler(sourceFileName string) (*FileSourceHandler, error) {
	fs := FileSourceHandler{
		sourceFileName: sourceFileName,
		index:          0,
	}

	var f *os.File
	var err error
	if f, err = os.Open(sourceFileName); err != nil {
		return nil, fmt.Errorf("error opening source file %q: %v", fs.sourceFileName, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fs.fileNames = append(fs.fileNames, line)
	}

	return &fs, nil
}

func (fs *FileSourceHandler) Close() error {
	return nil
}

func (fs *FileSourceHandler) GetFilename() (string, error) {
	if fs.index > len(fs.fileNames) {
		return "", io.EOF
	}

	filename := fs.fileNames[fs.index]
	fs.index++

	return filename, nil
}

func (fs *FileSourceHandler) AllFiles() func(yield func(s string) bool) {
	return func(yield func(s string) bool) {
		for i := range fs.fileNames {
			if !yield(fs.fileNames[i]) {
				return
			}
		}
	}
}
