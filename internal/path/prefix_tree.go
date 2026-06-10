package path

import "strings"

type prefixTree struct {
	subPrefix map[string]*prefixTree
	isNode    bool
	namespace string
}

func newPrefixTree() prefixTree {
	return prefixTree{subPrefix: make(map[string]*prefixTree)}
}

// InsertTree inserts a normalized path represented as path segments.
//
// The tree only works on already-normalized segments, for example:
//   Windows: []string{"f:", "hello", "filesystem"}
//   Linux:   []string{"etc", "cron.d"}
//
// Return value:
//   - "" when nothing was replaced.
//   - self when an existing parent prefix already covers this path.
//   - comma-joined child namespaces when this path replaces existing children.
func (pt *prefixTree) InsertTree(self string, segments []string) string {
	node := pt
	if node.subPrefix == nil {
		node.subPrefix = make(map[string]*prefixTree)
	}

	for _, seg := range segments {
		child, ok := node.subPrefix[seg]
		if !ok {
			child = &prefixTree{subPrefix: make(map[string]*prefixTree)}
			node.subPrefix[seg] = child
		}
		if child.isNode {
			return self
		}
		node = child
	}

	if len(node.subPrefix) > 0 {
		replaced := node.collectNamespaces()
		node.subPrefix = make(map[string]*prefixTree)
		node.isNode = true
		node.namespace = self
		return replaced
	}

	node.isNode = true
	node.namespace = self
	return ""
}

func (pt *prefixTree) collectNamespaces() string {
	var result []string
	pt.collect(&result)
	return strings.Join(result, ",")
}

func (pt *prefixTree) collect(result *[]string) {
	if pt.isNode {
		*result = append(*result, pt.namespace)
	}
	for _, child := range pt.subPrefix {
		child.collect(result)
	}
}

func (pt *prefixTree) Delete(self string, segments []string) bool {
	node := pt
	type step struct {
		parent *prefixTree
		seg    string
	}
	steps := make([]step, 0, len(segments))

	for _, seg := range segments {
		child, ok := node.subPrefix[seg]
		if !ok {
			return false
		}
		steps = append(steps, step{parent: node, seg: seg})
		node = child
	}

	if !node.isNode || node.namespace != self {
		return false
	}

	node.isNode = false
	node.namespace = ""

	for i := len(steps) - 1; i >= 0; i-- {
		parent := steps[i].parent
		seg := steps[i].seg
		child := parent.subPrefix[seg]
		if child.isNode || len(child.subPrefix) > 0 {
			break
		}
		delete(parent.subPrefix, seg)
	}
	return true
}

func (pt *prefixTree) Match(segments []string) bool {
	node := pt
	for _, seg := range segments {
		child, ok := node.subPrefix[seg]
		if !ok {
			return false
		}
		if child.isNode {
			return true
		}
		node = child
	}
	return false
}
