//go:build !windows

package innerpath

import (
	"path/filepath"
)

func ToNativePath(p Path) (string, error) {
	switch p.Kind {
	case PathUnixAbs:
		return "/" + filepath.Join(p.Segments...), nil
	case PathRelative:
		return filepath.Clean(filepath.Join(p.Segments...)), nil
	case PathWinDrive, PathWinUNC:
		return "", ErrUnsupportedOnPlatform
	default:
		return "", ErrUnsupportedOnPlatform
	}
}

func FromNativeSeparators(path string) string {
	return path
}

