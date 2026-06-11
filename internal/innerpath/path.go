// Package innerpath converts supported user path forms into platform-neutral
// path segments and back to native OS paths at the filesystem boundary.
//
// It accepts concrete filesystem paths, not shell expressions. Callers should
// resolve values such as $HOME, %USERPROFILE%, ~, globs, and PATH-like lists
// before calling this package.
package innerpath

import (
	"net/url"
	"strings"
)

func Parse(input string) (Path, error) {
	kind := Classify(input)
	switch kind {
	case PathNamespace:
		return parseNamespace(input)
	case PathRelative:
		return parseRelative(input)
	case PathWinDrive:
		return parseWindowsDrive(input)
	case PathWinUNC:
		return parseWindowsUNC(input)
	case PathUnixAbs:
		return parseUnixAbs(input)
	case PathFileURI:
		return parseFileURI(input)
	default:
		if strings.TrimSpace(input) == "" {
			return Path{Kind: PathUnsupported}, ErrEmptyPath
		}
		return Path{Kind: PathUnsupported}, unsupported(input, "pass a concrete filesystem path, not a shell expression or glob")
	}
}

func parseNamespace(input string) (Path, error) {
	namespace, rest, ok := strings.Cut(input, ":")
	if !ok || namespace == "" {
		return Path{Kind: PathUnsupported}, ErrInvalidNamespacePath
	}
	rest = strings.TrimLeft(rest, `/\`)
	segments, err := splitClean(rest, true)
	if err != nil {
		return Path{Kind: PathNamespace, Namespace: namespace}, err
	}
	return Path{
		Kind:      PathNamespace,
		Namespace: namespace,
		Segments:  segments,
	}, nil
}

func parseRelative(input string) (Path, error) {
	s := strings.TrimSpace(input)
	if s == "." {
		return Path{Kind: PathRelative}, nil
	}
	s = strings.TrimPrefix(s, "./")
	s = strings.TrimPrefix(s, `.\`)
	segments, err := splitClean(s, true)
	if err != nil {
		return Path{Kind: PathRelative}, err
	}
	return Path{Kind: PathRelative, Segments: segments}, nil
}

func parseUnixAbs(input string) (Path, error) {
	segments, err := splitClean(strings.TrimPrefix(input, "/"), false)
	if err != nil {
		return Path{Kind: PathUnixAbs}, err
	}
	return Path{Kind: PathUnixAbs, Segments: segments}, nil
}

func parseFileURI(input string) (Path, error) {
	u, err := url.Parse(input)
	if err != nil {
		return Path{Kind: PathFileURI}, err
	}
	if strings.ToLower(u.Scheme) != "file" {
		return Path{Kind: PathUnsupported}, unsupported(input, "only file:// URIs are supported")
	}

	if u.Host != "" && u.Host != "localhost" {
		return parseWindowsUNC(`\\` + u.Host + u.Path)
	}

	decoded := u.Path
	if isFileURIWindowsDrivePath(decoded) {
		return parseWindowsDrive(strings.TrimPrefix(decoded, "/"))
	}
	return parseUnixAbs(decoded)
}

func splitClean(s string, allowBackslashSeparator bool) ([]string, error) {
	if allowBackslashSeparator {
		s = strings.ReplaceAll(s, `\`, "/")
	}

	raw := strings.Split(s, "/")
	segments := make([]string, 0, len(raw))
	for _, seg := range raw {
		switch seg {
		case "", ".":
			continue
		case "..":
			return nil, ErrPathTraversal
		default:
			segments = append(segments, seg)
		}
	}
	return segments, nil
}

