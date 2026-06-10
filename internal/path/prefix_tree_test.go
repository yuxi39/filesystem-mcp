package path

import "testing"

func newPrefixTreeForTest() *prefixTree {
	return &prefixTree{subPrefix: make(map[string]*prefixTree)}
}

func TestPrefixTreeInsertAndMatch(t *testing.T) {
	tree := newPrefixTreeForTest()

	if replaced := tree.InsertTree("root", []string{"f:", "odds"}); replaced != "" {
		t.Fatalf("unexpected replacement: %q", replaced)
	}

	if !tree.Match([]string{"f:", "odds"}) {
		t.Fatal("expected exact path to match")
	}
	if !tree.Match([]string{"f:", "odds", "hello"}) {
		t.Fatal("expected child path to match parent prefix")
	}
	if tree.Match([]string{"f:", "other"}) {
		t.Fatal("unexpected match outside registered prefix")
	}
}

func TestPrefixTreeInsertChildUnderExistingRoot(t *testing.T) {
	tree := newPrefixTreeForTest()
	tree.InsertTree("root", []string{"f:", "odds"})

	if replaced := tree.InsertTree("child", []string{"f:", "odds", "hello"}); replaced != "child" {
		t.Fatalf("expected child insert to be rejected as itself, got %q", replaced)
	}

	if !tree.Match([]string{"f:", "odds", "hello"}) {
		t.Fatal("existing root should still match child paths")
	}
}

func TestPrefixTreeInsertParentOverChildren(t *testing.T) {
	tree := newPrefixTreeForTest()
	tree.InsertTree("hello", []string{"f:", "odds", "hello"})
	tree.InsertTree("world", []string{"f:", "odds", "world"})

	replaced := tree.InsertTree("odds", []string{"f:", "odds"})
	if replaced == "" {
		t.Fatal("expected child namespaces to be returned")
	}

	if !tree.Match([]string{"f:", "odds"}) {
		t.Fatal("new parent root should match exact path")
	}
	if !tree.Match([]string{"f:", "odds", "hello", "README.md"}) {
		t.Fatal("new parent root should match old child path")
	}
}

func TestPrefixTreeDeleteOnlyNode(t *testing.T) {
	tree := newPrefixTreeForTest()
	tree.InsertTree("root", []string{"f:", "odds"})

	if !tree.Delete("root", []string{"f:", "odds"}) {
		t.Fatal("expected delete to succeed")
	}
	if tree.Match([]string{"f:", "odds"}) {
		t.Fatal("deleted prefix should not match")
	}
}

func TestPrefixTreeDeleteOneBranchKeepsSibling(t *testing.T) {
	tree := newPrefixTreeForTest()
	tree.InsertTree("hello", []string{"f:", "odds", "hello"})
	tree.InsertTree("world", []string{"f:", "odds", "world"})

	if !tree.Delete("hello", []string{"f:", "odds", "hello"}) {
		t.Fatal("expected delete to succeed")
	}
	if tree.Match([]string{"f:", "odds", "hello"}) {
		t.Fatal("deleted branch should not match")
	}
	if !tree.Match([]string{"f:", "odds", "world"}) {
		t.Fatal("sibling branch should still match")
	}
}

