package innerpath

import (
	"strings"
	"unicode"
)

func isWindowsDrivePath(s string) bool {
	return len(s) >= 3 &&
		unicode.IsLetter(rune(s[0])) &&
		s[1] == ':' &&
		(s[2] == '\\' || s[2] == '/')
}

func isFileURIWindowsDrivePath(s string) bool {
	return len(s) >= 4 &&
		s[0] == '/' &&
		unicode.IsLetter(rune(s[1])) &&
		s[2] == ':' &&
		s[3] == '/'
}

func isWindowsUNC(s string) bool {
	return strings.HasPrefix(s, `\\`) || strings.HasPrefix(s, `//`)
}

func parseWindowsDrive(input string) (Path, error) {
	s := strings.ReplaceAll(input, `\`, "/")
	drive := strings.ToLower(s[:2])
	rest := strings.TrimLeft(s[2:], "/")
	segments, err := splitClean(rest, false)
	if err != nil {
		return Path{Kind: PathWinDrive}, err
	}
	return Path{Kind: PathWinDrive, Segments: append([]string{drive}, segments...)}, nil
}

func parseWindowsUNC(input string) (Path, error) {
	s := strings.ReplaceAll(input, `\`, "/")
	s = strings.TrimLeft(s, "/")
	segments, err := splitClean(s, false)
	if err != nil {
		return Path{Kind: PathWinUNC}, err
	}
	if len(segments) < 2 {
		return Path{Kind: PathWinUNC}, unsupported(input, "UNC paths must include a server and share")
	}
	return Path{Kind: PathWinUNC, Segments: append([]string{"unc"}, segments...)}, nil
}
