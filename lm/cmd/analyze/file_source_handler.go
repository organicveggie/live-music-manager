package analyze

import (
	"bufio"
	"fmt"
	"os"
)

type FileSourceHandler struct {
	sourceFileName string
	file           *os.File
}

func newFileSourceHandler(sourceFileName string) (*FileSourceHandler, error) {
	fs := FileSourceHandler{
		sourceFileName: sourceFileName,
	}

	var err error
	if fs.file, err = os.Open(sourceFileName); err != nil {
		return nil, fmt.Errorf("error opening source file %q: %v", fs.sourceFileName, err)
	}

	return &fs, nil
}

func (fs *FileSourceHandler) Close() error {
	return fs.file.Close()
}

func (fs *FileSourceHandler) AllFiles() func(yield func(string, error) bool) {
	return func(yield func(string, error) bool) {
		scanner := bufio.NewScanner(fs.file)
		for scanner.Scan() {
			line := scanner.Text()
			if !yield(line, nil) {
				return
			}
		}
	}
}

func (fs *FileSourceHandler) AnalyzeFiles(fn AnalyzerFn) error {
	scanner := bufio.NewScanner(fs.file)
	for scanner.Scan() {
		line := scanner.Text()
		if err := fn(line); err != nil {
			return err
		}
	}
	return nil
}
