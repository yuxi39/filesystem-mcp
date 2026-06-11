package innerpath

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type PathKind string

const (
	PathNamespace   PathKind = "namespace"
	PathRelative    PathKind = "relative"
	PathWinDrive    PathKind = "win_drive"
	PathWinUNC      PathKind = "win_unc"
	PathUnixAbs     PathKind = "unix_abs"
	PathFileURI     PathKind = "file_uri"
	PathUnsupported PathKind = "unsupported"
)

var (
	ErrEmptyPath             = errors.New("empty path")
	ErrUnsupportedPathSyntax = errors.New("unsupported path syntax")
	ErrPathTraversal         = errors.New("path traversal is not allowed")
	ErrInvalidNamespacePath  = errors.New("invalid namespace path")
	ErrUnsupportedOnPlatform = errors.New("path kind is unsupported on this platform")
)

type Path struct {
	Kind      PathKind `json:"kind"`
	Namespace string   `json:"namespace,omitempty"`
	Segments  []string `json:"segments"`
}

type UnsupportedPathError struct {
	Input    string
	Reason   string
	Examples []string
}

func (e *UnsupportedPathError) Error() string {
	return fmt.Sprintf("%v: %s", ErrUnsupportedPathSyntax, e.Reason)
}

func (e *UnsupportedPathError) Unwrap() error {
	return ErrUnsupportedPathSyntax
}

func Classify(input string) PathKind {
	s := strings.TrimSpace(input)
	if s == "" {
		return PathUnsupported
	}
	if hasUnsupportedShellSyntax(s) {
		return PathUnsupported
	}
	if strings.HasPrefix(strings.ToLower(s), "file://") {
		return PathFileURI
	}
	if isWindowsUNC(s) {
		return PathWinUNC
	}
	if isWindowsDrivePath(s) {
		return PathWinDrive
	}
	if strings.HasPrefix(s, "/") {
		return PathUnixAbs
	}
	if strings.HasPrefix(s, "./") || strings.HasPrefix(s, `.\`) || s == "." {
		return PathRelative
	}
	if isNamespacePath(s) {
		return PathNamespace
	}
	return PathUnsupported
}

func hasUnsupportedShellSyntax(s string) bool {
	if strings.HasPrefix(s, "~") ||
		strings.HasPrefix(s, "$") ||
		strings.HasPrefix(s, "%") ||
		strings.HasPrefix(strings.ToLower(s), "$env:") ||
		strings.Contains(s, "$(") {
		return true
	}
	return strings.ContainsAny(s, "*?")
}

func isNamespacePath(s string) bool {
	namespace, _, ok := strings.Cut(s, ":")
	if !ok || namespace == "" {
		return false
	}
	for _, r := range namespace {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.' {
			continue
		}
		return false
	}
	return true
}

func unsupported(input, reason string) error {
	return &UnsupportedPathError{
		Input:  input,
		Reason: reason,
		Examples: []string{
			`C:\Users\User\Desktop`,
			`filesystem:README.md`,
			`./README.md with an explicit base root`,
			`/home/user`,
			`\\wsl$\`,
		},
	}
}
