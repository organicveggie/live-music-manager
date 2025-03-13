package analyze

import (
	"bufio"
	"fmt"
	"os"
)

type FileSourceHandler struct {
	sourceFileName string
	file           *os.File
	ch             AnalyzeChan
}

func newFileSourceHandler(ch AnalyzeChan, sourceFileName string) (*FileSourceHandler, error) {
	fs := FileSourceHandler{
		ch:             ch,
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

func (fs *FileSourceHandler) Analyze() error {
	scanner := bufio.NewScanner(fs.file)
	for scanner.Scan() {
		line := scanner.Text()
		fs.ch <- line
	}
	return nil
}
