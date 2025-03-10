package scan

import "fmt"

type FileAddStdOut struct{}

func newFileAddStdOut() *FileAddStdOut {
	return &FileAddStdOut{}
}

func (fas *FileAddStdOut) AddFile(filename string) error {
	fmt.Printf("%q\n", filename)
	return nil
}

func (f *FileAddStdOut) Close() error {
	return nil
}
