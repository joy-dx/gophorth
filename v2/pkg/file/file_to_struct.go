package file

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"

	"github.com/spf13/afero"
)

func FileToStruct(filePath string, structPointer interface{}) error {
	var appFS = afero.NewOsFs()
	if fileHandle, fileOpenErr := appFS.Open(filePath); fileOpenErr != nil {
		return fileOpenErr
	} else {
		defer fileHandle.Close()
		if fileContent, fileReadErr := afero.ReadAll(fileHandle); fileReadErr != nil {
			return fileReadErr
		} else {
			switch filepath.Ext(filePath) {
			case ".yaml", ".yml":
				if unmarshallErr := yaml.Unmarshal(fileContent, structPointer); unmarshallErr != nil {
					return unmarshallErr
				}
			case ".json":
				if unmarshallErr := json.Unmarshal(fileContent, structPointer); unmarshallErr != nil {
					return unmarshallErr
				}
			case ".toml":
				if unmarshallErr := toml.Unmarshal(fileContent, structPointer); unmarshallErr != nil {
					return unmarshallErr
				}
			default:
				return fmt.Errorf("unhandled organisaton environment extension: %s", filePath)
			}
		}
	}
	return nil
}
