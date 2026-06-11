package innerpath

import (
	"errors"
	"reflect"
	"runtime"
	"testing"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  PathKind
	}{
		{"namespace", "odds:hello/README.md", PathNamespace},
		{"relative slash", "./hello/README.md", PathRelative},
		{"relative backslash", `.\hello\README.md`, PathRelative},
		{"windows drive backslash", `F:\ODDS&ENDS\hello`, PathWinDrive},
		{"windows drive slash", "F:/ODDS&ENDS/hello", PathWinDrive},
		{"windows unc", `\\wsl$\Ubuntu\home\yuxi`, PathWinUNC},
		{"unix absolute", "/etc/cron.d", PathUnixAbs},
		{"file uri", "file:///f%3A/ODDS%26ENDS/hello%20world", PathFileURI},
		{"home shorthand unsupported", "~/project", PathUnsupported},
		{"powershell env unsupported", `$env:USERPROFILE\Desktop`, PathUnsupported},
		{"cmd env unsupported", `%USERPROFILE%\Desktop`, PathUnsupported},
		{"glob unsupported", "*.go", PathUnsupported},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Classify(tt.input); got != tt.want {
				t.Fatalf("Classify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseWindowsDrivePath(t *testing.T) {
	got, err := Parse(`F:\ODDS&ENDS\hello world\中文\日本語`)
	if err != nil {
		t.Fatal(err)
	}
	want := Path{
		Kind:     PathWinDrive,
		Segments: []string{"f:", "ODDS&ENDS", "hello world", "中文", "日本語"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseWindowsUNCPath(t *testing.T) {
	got, err := Parse(`\\wsl$\Ubuntu\home\yuxi\项目`)
	if err != nil {
		t.Fatal(err)
	}
	want := Path{
		Kind:     PathWinUNC,
		Segments: []string{"unc", "wsl$", "Ubuntu", "home", "yuxi", "项目"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseUnixAbsPath(t *testing.T) {
	got, err := Parse("/etc/cron.d")
	if err != nil {
		t.Fatal(err)
	}
	want := Path{Kind: PathUnixAbs, Segments: []string{"etc", "cron.d"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseNamespacePath(t *testing.T) {
	got, err := Parse("odds:hello world/中文/日本語")
	if err != nil {
		t.Fatal(err)
	}
	want := Path{
		Kind:      PathNamespace,
		Namespace: "odds",
		Segments:  []string{"hello world", "中文", "日本語"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseFileURI(t *testing.T) {
	got, err := Parse("file:///f%3A/ODDS%26ENDS/hello%20world/%E4%B8%AD%E6%96%87")
	if err != nil {
		t.Fatal(err)
	}
	want := Path{
		Kind:     PathWinDrive,
		Segments: []string{"f:", "ODDS&ENDS", "hello world", "中文"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestUnsupportedShellSyntax(t *testing.T) {
	for _, input := range []string{
		"~/project",
		`$env:USERPROFILE\Desktop`,
		`%USERPROFILE%\Desktop`,
		"*.go",
	} {
		_, err := Parse(input)
		if !errors.Is(err, ErrUnsupportedPathSyntax) {
			t.Fatalf("Parse(%q) error = %v, want ErrUnsupportedPathSyntax", input, err)
		}
	}
}

func TestPathTraversalRejected(t *testing.T) {
	for _, input := range []string{
		"odds:../secret",
		"./../secret",
		`F:\ODDS&ENDS\..\secret`,
		"/etc/../secret",
	} {
		_, err := Parse(input)
		if !errors.Is(err, ErrPathTraversal) {
			t.Fatalf("Parse(%q) error = %v, want ErrPathTraversal", input, err)
		}
	}
}

func TestToNativePath(t *testing.T) {
	if runtime.GOOS == "windows" {
		got, err := ToNativePath(Path{Kind: PathWinDrive, Segments: []string{"f:", "ODDS&ENDS", "hello"}})
		if err != nil {
			t.Fatal(err)
		}
		if got != `f:\ODDS&ENDS\hello` {
			t.Fatalf("got %q", got)
		}
		return
	}

	got, err := ToNativePath(Path{Kind: PathUnixAbs, Segments: []string{"etc", "cron.d"}})
	if err != nil {
		t.Fatal(err)
	}
	if got != "/etc/cron.d" {
		t.Fatalf("got %q", got)
	}
}

