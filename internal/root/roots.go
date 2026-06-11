package root

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/yuxi39/filesystem-mcp/internal/innerpath"
)

var (
	ErrRootNotFound      = errors.New("root not found")
	ErrInvalidRootPath   = errors.New("invalid root path")
	ErrRootPathCovered   = errors.New("root path is already covered by an existing root")
	ErrRootPathNotNative = errors.New("root path cannot be represented on this platform")
)

type RootManager struct {
	mu   sync.RWMutex
	rts  map[string]*Root
	tree *RootTree
}

func NewRootManager() *RootManager {
	return &RootManager{
		rts:  make(map[string]*Root),
		tree: NewRootTree(false),
	}
}

var Global = NewRootManager()

func (rmng *RootManager) List() []string {
	rmng.mu.RLock()
	defer rmng.mu.RUnlock()

	result := make([]string, 0, len(rmng.rts))
	for namespace := range rmng.rts {
		result = append(result, namespace)
	}
	return result
}

func (rmng *RootManager) ListAll() []*Root {
	rmng.mu.RLock()
	defer rmng.mu.RUnlock()

	result := make([]*Root, 0, len(rmng.rts))
	for _, rt := range rmng.rts {
		cp := *rt
		cp.PathSegments = append([]string(nil), rt.PathSegments...)
		result = append(result, &cp)
	}
	return result
}

func (rmng *RootManager) ListTarget(namespace string) *Root {
	rmng.mu.RLock()
	defer rmng.mu.RUnlock()

	rt, ok := rmng.rts[namespace]
	if !ok {
		return nil
	}
	cp := *rt
	cp.PathSegments = append([]string(nil), rt.PathSegments...)
	return &cp
}

func (rmng *RootManager) Add(namespace, rawPath string) (*Root, []string, error) {
	if namespace == "" {
		return nil, nil, fmt.Errorf("%w: namespace is required", ErrInvalidRootPath)
	}

	parsed, err := innerpath.Parse(rawPath)
	if err != nil {
		return nil, nil, err
	}
	if parsed.Kind == innerpath.PathNamespace || parsed.Kind == innerpath.PathRelative || parsed.Kind == innerpath.PathUnsupported {
		return nil, nil, fmt.Errorf("%w: roots require a concrete absolute path, got %q", ErrInvalidRootPath, rawPath)
	}

	native, err := innerpath.ToNativePath(parsed)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrRootPathNotNative, err)
	}

	rt := &Root{
		PathType:     parsed.Kind,
		NameSpace:    namespace,
		InnerPath:    strings.Join(parsed.Segments, "/"),
		AbsolutePath: native,
		PathSegments: append([]string(nil), parsed.Segments...),
		Status:       RootStatusUnknown,
	}
	rt.RefreshStatus()

	rmng.mu.Lock()
	defer rmng.mu.Unlock()

	var old *Root
	if existing, exists := rmng.rts[namespace]; exists {
		old = existing
		rmng.tree.Delete(namespace, old.PathSegments)
		delete(rmng.rts, namespace)
	}

	ok, removedJoined := rmng.tree.Insert(namespace, rt.PathSegments)
	if !ok {
		if old != nil {
			rmng.tree.Insert(namespace, old.PathSegments)
			rmng.rts[namespace] = old
		}
		return nil, nil, fmt.Errorf("%w: %q", ErrRootPathCovered, rawPath)
	}

	var removed []string
	if removedJoined != "" {
		removed = strings.Split(removedJoined, ",")
		for _, ns := range removed {
			delete(rmng.rts, ns)
		}
	}

	rmng.rts[namespace] = rt
	cp := *rt
	cp.PathSegments = append([]string(nil), rt.PathSegments...)
	return &cp, removed, nil
}

func (rmng *RootManager) Delete(namespace string) (*Root, error) {
	rmng.mu.Lock()
	defer rmng.mu.Unlock()

	rt, ok := rmng.rts[namespace]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrRootNotFound, namespace)
	}
	rmng.tree.Delete(namespace, rt.PathSegments)
	delete(rmng.rts, namespace)

	cp := *rt
	cp.PathSegments = append([]string(nil), rt.PathSegments...)
	return &cp, nil
}

func (rmng *RootManager) Match(segments []string) (bool, string) {
	rmng.mu.RLock()
	defer rmng.mu.RUnlock()
	return rmng.tree.Match(segments)
}

func (r *Root) RefreshStatus() {
	info, err := os.Stat(r.AbsolutePath)
	r.LastChecked = time.Now()
	r.LastError = ""
	if err == nil {
		if info.IsDir() {
			r.Status = RootStatusReachable
			return
		}
		r.Status = RootStatusMissing
		r.LastError = "root path is not a directory"
		return
	}
	if os.IsNotExist(err) {
		r.Status = RootStatusMissing
		r.LastError = err.Error()
		return
	}
	if os.IsPermission(err) {
		r.Status = RootStatusDenied
		r.LastError = err.Error()
		return
	}
	r.Status = RootStatusUnknown
	r.LastError = err.Error()
}
