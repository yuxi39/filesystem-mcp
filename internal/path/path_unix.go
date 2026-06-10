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

func (pm *PathManager) Resolve(input string) *Path {
	switch {
	case strings.HasPrefix(input, "/"):

	}
	return nil
}

func (pm *PathManager) resolveAbsolute(input string) *Path {
	// 先实现前缀树, 思路:使用hash算法映射到32位的数上,将这个数转成int32类型,然后使用map嵌套结构
	return nil
}
