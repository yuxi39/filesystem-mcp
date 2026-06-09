package path

import (
	"runtime"
	"testing"
)

// --- Windows-specific tests ---

func TestURIToPath_WindowsBasic(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}
	uri := "file:///f%3A/ODDS%26ENDS/filesystem"
	got, err := URIToPath(uri)
	if err != nil {
		t.Fatal(err)
	}
	if got != `f:\ODDS&ENDS\filesystem` {
		t.Fatalf("got %q", got)
	}
}

func TestURIToPath_WindowsChinese(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}
	uri := "file:///f%3A/%E4%B8%AD%E6%96%87/%E6%B5%8B%E8%AF%95"
	got, err := URIToPath(uri)
	if err != nil {
		t.Fatal(err)
	}
	if got != `f:\中文\测试` {
		t.Fatalf("got %q", got)
	}
}

func TestURIToPath_WindowsJapanese(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}
	uri := "file:///f%3A/%E6%97%A5%E6%9C%AC%E8%AA%9E/%E3%83%95%E3%82%A1%E3%82%A4%E3%83%AB"
	got, err := URIToPath(uri)
	if err != nil {
		t.Fatal(err)
	}
	if got != `f:\日本語\ファイル` {
		t.Fatalf("got %q", got)
	}
}

func TestPathToURI_RoundTrip_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}
	orig := `f:\ODDS&ENDS\filesystem`
	uri := PathToURI(orig)
	got, err := URIToPath(uri)
	if err != nil {
		t.Fatal(err)
	}
	if got != orig {
		t.Fatalf("round trip: %q → %q → %q", orig, uri, got)
	}
}

// --- Unix / macOS tests ---

func TestURIToPath_UnixBasic(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-specific test")
	}
	uri := "file:///home/user/project"
	got, err := URIToPath(uri)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/home/user/project" {
		t.Fatalf("got %q", got)
	}
}

func TestURIToPath_Unicode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-specific test")
	}
	uri := "file:///home/%E4%B8%AD%E6%96%87/%E6%97%A5%E6%9C%AC%E8%AA%9E"
	got, err := URIToPath(uri)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/home/中文/日本語" {
		t.Fatalf("got %q", got)
	}
}

// --- Cross-platform tests ---

func TestURIToPath_InvalidScheme(t *testing.T) {
	_, err := URIToPath("https://example.com")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPathToURI_RoundTrip_Unicode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-specific test")
	}
	orig := "/home/中文/日本語"
	uri := PathToURI(orig)
	got, err := URIToPath(uri)
	if err != nil {
		t.Fatal(err)
	}
	if got != orig {
		t.Fatalf("round trip: %q → %q → %q", orig, uri, got)
	}
}
