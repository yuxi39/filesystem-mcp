package root

import "strings"

// RootTree stores registered root paths as a prefix tree over normalized path
// segments.
//
// The tree does not parse raw filesystem paths. Callers must pass cleaned,
// platform-normalized segments, for example:
//   Windows: []string{"f:", "ODDS&ENDS", "filesystem"}
//   Unix:    []string{"etc", "cron.d"}
//
// A node marked with isNode represents a registered root and stores the
// namespace supplied by RootManager. The namespace is not derived from the path.
// Once a root exists, no child root may be inserted below it. If a parent root
// is inserted above existing child roots, the child roots are removed and
// returned to the caller.
type RootTree struct {
	subTree   map[string]*RootTree
	isNode    bool
	namespace string
}

// NewRootTree creates an empty root tree.
//
// If admin is true, the root tree itself is considered a root and therefore
// matches every path. Normal workspace trees should pass false.
func NewRootTree(admin bool) *RootTree {
	return &RootTree{
		subTree: make(map[string]*RootTree),
		isNode:  admin,
	}
}

// Insert registers a root path with its external namespace.
//
// It returns:
//   - ok=false when namespace or segments is empty, the same path already
//     exists, or an existing parent root already covers this path.
//   - ok=true and removed="" when a new independent root is inserted.
//   - ok=true and removed="a,b,c" when this new parent root replaces existing
//     child roots. The removed value is a comma-joined list of removed
//     namespaces.
func (rt *RootTree) Insert(namespace string, segments []string) (ok bool, removed string) {
	if namespace == "" || len(segments) == 0 {
		return false, ""
	}
	if rt.subTree == nil {
		rt.subTree = make(map[string]*RootTree)
	}

	node := rt
	for _, seg := range segments {
		child, exists := node.subTree[seg]
		if !exists {
			child = &RootTree{subTree: make(map[string]*RootTree)}
			node.subTree[seg] = child
		}
		if child.isNode {
			return false, ""
		}
		node = child
	}

	if node.isNode || node.namespace == namespace {
		return false, ""
	}

	if len(node.subTree) > 0 {
		removed = node.collectNamespaces()
		node.subTree = make(map[string]*RootTree)
	}
	node.isNode = true
	node.namespace = namespace
	return true, removed
}

// Delete removes an existing root path.
//
// Delete clears only the requested root and prunes now-empty branches while
// keeping sibling roots intact.
func (rt *RootTree) Delete(namespace string, segments []string) bool {
	if namespace == "" || len(segments) == 0 {
		return false
	}

	type step struct {
		parent *RootTree
		seg    string
	}

	node := rt
	steps := make([]step, 0, len(segments))
	for _, seg := range segments {
		child, exists := node.subTree[seg]
		if !exists {
			return false
		}
		steps = append(steps, step{parent: node, seg: seg})
		node = child
	}

	if !node.isNode || node.namespace != namespace {
		return false
	}

	node.isNode = false
	node.namespace = ""

	for i := len(steps) - 1; i >= 0; i-- {
		parent := steps[i].parent
		seg := steps[i].seg
		child := parent.subTree[seg]
		if child.isNode || len(child.subTree) > 0 {
			break
		}
		delete(parent.subTree, seg)
	}
	return true
}

// Match reports whether the given path is covered by a registered root.
//
// If a match exists, namespace is the namespace stored at the matching root
// node. A parent root matches all of its descendants.
func (rt *RootTree) Match(segments []string) (matched bool, namespace string) {
	if rt.isNode {
		return true, rt.namespace
	}

	node := rt
	for _, seg := range segments {
		child, exists := node.subTree[seg]
		if !exists {
			return false, ""
		}
		if child.isNode {
			return true, child.namespace
		}
		node = child
	}
	return false, ""
}

func (rt *RootTree) collectNamespaces() string {
	var result []string
	rt.collect(&result)
	return strings.Join(result, ",")
}

func (rt *RootTree) collect(result *[]string) {
	if rt.isNode {
		*result = append(*result, rt.namespace)
	}
	for _, child := range rt.subTree {
		child.collect(result)
	}
}
