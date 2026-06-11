//go:build windows

package innerpath

import (
	"path/filepath"
	"strings"
)

func ToNativePath(p Path) (string, error) {
	switch p.Kind {
	case PathWinDrive:
		if len(p.Segments) == 0 {
			return "", ErrEmptyPath
		}
		parts := append([]string{p.Segments[0] + `\`}, p.Segments[1:]...)
		return filepath.Clean(filepath.Join(parts...)), nil
	case PathWinUNC:
		if len(p.Segments) < 3 || p.Segments[0] != "unc" {
			return "", ErrEmptyPath
		}
		return `\\` + filepath.Join(p.Segments[1:]...), nil
	case PathUnixAbs:
		return "", ErrUnsupportedOnPlatform
	case PathRelative:
		return filepath.Clean(filepath.Join(p.Segments...)), nil
	default:
		return "", ErrUnsupportedOnPlatform
	}
}

func FromNativeSeparators(path string) string {
	return strings.ReplaceAll(path, "/", `\`)
}

