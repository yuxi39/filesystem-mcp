//go:build windows

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

	trimmed := strings.TrimPrefix(decoded, "/")
	if len(trimmed) >= 2 && trimmed[1] == ':' {
		return filepath.FromSlash(trimmed), nil
	}
	return filepath.Clean(decoded), nil
}

func PathToURI(path string) string {
	slashed := filepath.ToSlash(path)

	if len(slashed) >= 2 && slashed[1] == ':' {
		slashed = "/" + slashed
	}

	segments := strings.Split(slashed, "/")
	for i, seg := range segments {
		segments[i] = url.PathEscape(seg)
	}
	return "file://" + strings.Join(segments, "/")
}
