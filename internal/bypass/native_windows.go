//go:build windows

package bypass

import "path/filepath"

func joinNative(base string, segments []string) string {
	parts := append([]string{base}, segments...)
	return filepath.Clean(filepath.Join(parts...))
}
