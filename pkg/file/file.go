package file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

func FileToReader(filePath string) (*bufio.Reader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	return bufio.NewReader(file), nil
}

// PathExists Checks whether a given path exists as a node on the filesystem
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateSymlink(source string, destination string) error {
	appFs := &afero.OsFs{}

	symLink := func(l afero.Linker, source string, destination string) error {
		if err := l.SymlinkIfPossible(source, destination); err != nil {
			return err
		}
		return nil
	}
	return symLink(appFs, source, destination)
}

func StructToJSONFile(inputStruct interface{}, outputPath string) error {
	if serialized, err := json.Marshal(inputStruct); err != nil {
		return err
	} else {
		if err := os.MkdirAll(path.Dir(outputPath), fs.ModePerm); err != nil {
			return fmt.Errorf("problem creating output directory: %w", err)
		}
		return os.WriteFile(outputPath, serialized, 0644)
	}
}

func StructToIndentedJSONFile(inputStruct interface{}, outputPath string) error {

	if serialized, err := json.MarshalIndent(inputStruct, "", "  "); err != nil {
		return err
	} else {
		if err := os.MkdirAll(path.Dir(outputPath), fs.ModePerm); err != nil {
			return fmt.Errorf("problem creating output directory: %w", err)
		}
		return os.WriteFile(outputPath, serialized, 0644)
	}
}

func StructToYamlFile(inputStruct interface{}, outputPath string) error {
	if serialized, marshallErr := yaml.Marshal(inputStruct); marshallErr != nil {
		return marshallErr
	} else {
		if err := os.MkdirAll(path.Dir(outputPath), fs.ModePerm); err != nil {
			return fmt.Errorf("problem creating output directory: %w", err)
		}
		return os.WriteFile(outputPath, serialized, 0644)
	}
}
