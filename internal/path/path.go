package path

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
)

// TODO First All scope roots and bypass and ignore
// Roots: map[string]string		key:lowercase directory name value: absolutely path
// Bypass: map[string]bool 		key:lowercase absolutely path value: readable
// ignore: []string				value:lowercase directory name just not tree
// Path model should resolve uri, relative path, absolutely path, filesystem's path, windows path, unix path

// inner path format: [<root>:<path>/<to>/<directory>]

// feature:
// 		platform free
// 		path security

type PathManager struct {
	mu     sync.RWMutex
	roots  map[string]string
	bypass map[string]bool
	ignore map[string]struct{}
	match  prefixTree
}

type Path struct {
	FSPath string
}

func decodeFileURI(uri string) (string, error) {
	if !strings.HasPrefix(uri, "file://") {
		return "", fmt.Errorf("unsupported URI scheme: %s", uri)
	}
	rawPath := strings.TrimPrefix(uri, "file://")

	decoded, err := url.PathUnescape(rawPath)
	if err != nil {
		return "", fmt.Errorf("decoding URI path %q: %w", uri, err)
	}
	return decoded, nil
}
