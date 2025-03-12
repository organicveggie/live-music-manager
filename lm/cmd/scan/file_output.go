package scan

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type FileAddFileOut struct {
	filename string
	file     *os.File
	writer   *bufio.Writer
}

func newFileAddFileOut(filename string, overwrite bool) (*FileAddFileOut, error) {
	fileOut := &FileAddFileOut{
		filename: filename,
	}

	var err error
	if _, err = os.Stat(filename); err != nil {
		if errors.Is(err, fs.ErrExist) && !overwrite {
			return nil, fmt.Errorf("output file %s exists: %v", filename, err)
		}
	}

	if fileOut.file, err = os.Create(fileOut.filename); err != nil {
		return nil, err
	}

	fileOut.writer = bufio.NewWriter(fileOut.file)

	return fileOut, nil
}

func (f *FileAddFileOut) Close() error {
	if err := f.writer.Flush(); err != nil {
		return fmt.Errorf("error flushing output file %s: %v", f.filename, err)
	}
	if err := f.file.Close(); err != nil {
		return fmt.Errorf("error closing output file %s: %v", f.filename, err)
	}
	return nil
}

func (f *FileAddFileOut) AddFile(filename string) error {
	fmt.Fprintf(f.writer, "%s\n", filename)
	return nil
}
