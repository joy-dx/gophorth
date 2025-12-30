package file

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/spf13/afero"
)

func ToBytes(path string) ([]byte, error) {
	fileHandle, fileOpenErr := os.Open(os.ExpandEnv(path))
	if fileOpenErr != nil {
		return nil, fileOpenErr
	}
	defer fileHandle.Close()
	stat, err := fileHandle.Stat()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fileAsBytes := make([]byte, stat.Size())
	_, err = bufio.NewReader(fileHandle).Read(fileAsBytes)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return fileAsBytes, nil
}

func BytesToFile(contents []byte, path string) error {
	var appFs = afero.NewOsFs()

	// Make sure the destination folder exists
	fsPath := path[:strings.LastIndex(path, "/")]
	if err := appFs.MkdirAll(fsPath, fs.ModePerm); err != nil {
		return fmt.Errorf("mkdir '%s': %s", fsPath, err)
	}

	// Create and write the file contents
	fh, fileOpenErr := appFs.Create(path)
	if fileOpenErr != nil {
		return fileOpenErr
	}
	defer fh.Close()
	if _, writeErr := fh.Write(contents); writeErr != nil {
		return writeErr
	}
	return nil
}
