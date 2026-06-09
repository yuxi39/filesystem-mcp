//go:build !windows

package path

import (
	"net/url"
	"path/filepath"
	"strings"
)

func URIToPath(uri string) (string, error) {
	decoded, err := decodeFileURI(uri)
	if err != nil {
		return "", err
	}
	return filepath.Clean(decoded), nil
}

func PathToURI(path string) string {
	segments := strings.Split(path, "/")
	for i, seg := range segments {
		segments[i] = url.PathEscape(seg)
	}
	return "file://" + strings.Join(segments, "/")
}
