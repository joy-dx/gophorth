package stringz

import (
	"net/url"
	"path/filepath"
)

func FilenameFromUrl(inputUrl string) (string, error) {
	u, err := url.Parse(inputUrl)
	if err != nil {
		return "", err
	}
	x, _ := url.QueryUnescape(u.EscapedPath())
	return filepath.Base(x), nil
}
