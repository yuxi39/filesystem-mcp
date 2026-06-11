package root

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootManagerAddReplacesSameNamespace(t *testing.T) {
	manager := NewRootManager()
	first := t.TempDir()
	second := t.TempDir()

	if _, _, err := manager.Add("workspace", first); err != nil {
		t.Fatal(err)
	}
	got, removed, err := manager.Add("workspace", second)
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 0 {
		t.Fatalf("unexpected removed namespaces: %v", removed)
	}
	if !sameNativePath(got.AbsolutePath, second) {
		t.Fatalf("got path %q, want %q", got.AbsolutePath, second)
	}
}

func TestRootManagerFailedReplaceRestoresOldNamespace(t *testing.T) {
	manager := NewRootManager()
	parent := t.TempDir()
	child := filepath.Join(parent, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	old := t.TempDir()

	if _, _, err := manager.Add("parent", parent); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.Add("workspace", old); err != nil {
		t.Fatal(err)
	}
	if _, _, err := manager.Add("workspace", child); err == nil {
		t.Fatal("expected replace to fail because child path is covered by parent root")
	}

	got := manager.ListTarget("workspace")
	if got == nil {
		t.Fatal("expected old workspace root to be restored")
	}
	if !sameNativePath(got.AbsolutePath, old) {
		t.Fatalf("got restored path %q, want %q", got.AbsolutePath, old)
	}
}

func sameNativePath(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}
