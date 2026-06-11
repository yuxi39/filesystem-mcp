package root

import "testing"

func TestRootTreeInsertAndMatch(t *testing.T) {
	tree := NewRootTree(false)

	ok, removed := tree.Insert("workspace", []string{"f:", "odds"})
	if !ok {
		t.Fatal("expected insert to succeed")
	}
	if removed != "" {
		t.Fatalf("unexpected removed namespaces: %q", removed)
	}

	matched, namespace := tree.Match([]string{"f:", "odds", "hello"})
	if !matched || namespace != "workspace" {
		t.Fatalf("expected workspace match, got matched=%v namespace=%q", matched, namespace)
	}
}

func TestRootTreeRejectsChildUnderExistingRoot(t *testing.T) {
	tree := NewRootTree(false)
	tree.Insert("workspace", []string{"f:", "odds"})

	ok, removed := tree.Insert("child", []string{"f:", "odds", "hello"})
	if ok {
		t.Fatal("expected child insert under existing root to be rejected")
	}
	if removed != "" {
		t.Fatalf("unexpected removed namespaces: %q", removed)
	}
}

func TestRootTreeParentReplacesChildren(t *testing.T) {
	tree := NewRootTree(false)
	tree.Insert("hello-root", []string{"f:", "odds", "hello"})
	tree.Insert("world-root", []string{"f:", "odds", "world"})

	ok, removed := tree.Insert("workspace", []string{"f:", "odds"})
	if !ok {
		t.Fatal("expected parent insert to succeed")
	}
	if removed == "" {
		t.Fatal("expected removed child namespaces")
	}

	matched, namespace := tree.Match([]string{"f:", "odds", "hello"})
	if !matched || namespace != "workspace" {
		t.Fatalf("expected parent workspace to match child path, got matched=%v namespace=%q", matched, namespace)
	}
}

func TestRootTreeDeleteLeafKeepsSiblingUnderSameParent(t *testing.T) {
	tree := NewRootTree(false)
	tree.Insert("hello-root", []string{"f:", "odds", "hello"})
	tree.Insert("world-root", []string{"f:", "odds", "world"})

	if !tree.Delete("hello-root", []string{"f:", "odds", "hello"}) {
		t.Fatal("expected delete to succeed")
	}

	matched, _ := tree.Match([]string{"f:", "odds", "hello"})
	if matched {
		t.Fatal("deleted branch should not match")
	}

	matched, namespace := tree.Match([]string{"f:", "odds", "world"})
	if !matched || namespace != "world-root" {
		t.Fatalf("expected sibling world-root to remain, got matched=%v namespace=%q", matched, namespace)
	}
}

func TestRootTreeDeleteDeepLeafKeepsSiblingBelowSingleChain(t *testing.T) {
	tree := NewRootTree(false)
	tree.Insert("c-root", []string{"f:", "a", "b", "c"})
	tree.Insert("d-root", []string{"f:", "a", "b", "d"})

	if !tree.Delete("c-root", []string{"f:", "a", "b", "c"}) {
		t.Fatal("expected delete to succeed")
	}

	matched, _ := tree.Match([]string{"f:", "a", "b", "c"})
	if matched {
		t.Fatal("deleted branch should not match")
	}

	matched, namespace := tree.Match([]string{"f:", "a", "b", "d"})
	if !matched || namespace != "d-root" {
		t.Fatalf("expected sibling d-root to remain, got matched=%v namespace=%q", matched, namespace)
	}
}

func TestRootTreeEmptySegmentsAreInvalid(t *testing.T) {
	tree := NewRootTree(false)

	if ok, _ := tree.Insert("empty", nil); ok {
		t.Fatal("empty insert should be rejected")
	}
	if ok, _ := tree.Insert("", []string{"f:"}); ok {
		t.Fatal("empty namespace insert should be rejected")
	}
	if tree.Delete("empty", nil) {
		t.Fatal("empty delete should be rejected")
	}
	if tree.Delete("", []string{"f:"}) {
		t.Fatal("empty namespace delete should be rejected")
	}
	if matched, _ := tree.Match(nil); matched {
		t.Fatal("empty match should not match")
	}
}
