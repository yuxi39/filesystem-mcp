package bypass

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/yuxi39/filesystem-mcp/internal/innerpath"
	"github.com/yuxi39/filesystem-mcp/internal/root"
)

var (
	ErrBypassNotFound     = errors.New("bypass rule not found")
	ErrInvalidBypassPath  = errors.New("invalid bypass path")
	ErrBypassRootNotFound = errors.New("bypass root not found")
)

type BypassRule struct {
	Index        int      `json:"index" jsonschema:"Index used to delete this bypass rule with path/bypass/del."`
	Root         string   `json:"root" jsonschema:"Root namespace that owns this bypass rule."`
	Path         string   `json:"path" jsonschema:"Namespace path that is blocked, for example odds:secret."`
	InnerPath    string   `json:"innerPath" jsonschema:"Platform-independent slash-joined path segments for matching."`
	AbsolutePath string   `json:"absolutePath" jsonschema:"Native absolute path that this bypass rule blocks."`
	Segments     []string `json:"segments" jsonschema:"Normalized platform-independent path segments used for prefix matching."`
	Reason       string   `json:"reason" jsonschema:"Human-readable reason shown to the AI when this path is blocked."`
	Readable     bool     `json:"readable" jsonschema:"Whether the path is readable. false means blocked."`
}

type Manager struct {
	mu    sync.RWMutex
	rules []BypassRule
}

var Global = &Manager{}

func (m *Manager) List() []BypassRule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]BypassRule, len(m.rules))
	copy(out, m.rules)
	for i := range out {
		out[i].Index = i
		out[i].Segments = append([]string(nil), out[i].Segments...)
	}
	return out
}

func (m *Manager) Add(namespacePath, reason string, roots *root.RootManager) (BypassRule, error) {
	parsed, err := innerpath.Parse(namespacePath)
	if err != nil {
		return BypassRule{}, err
	}
	if parsed.Kind != innerpath.PathNamespace || parsed.Namespace == "" {
		return BypassRule{}, fmt.Errorf("%w: bypass path must use namespace form such as odds:secret", ErrInvalidBypassPath)
	}

	rt := roots.ListTarget(parsed.Namespace)
	if rt == nil {
		return BypassRule{}, fmt.Errorf("%w: %s", ErrBypassRootNotFound, parsed.Namespace)
	}

	segments := append([]string(nil), rt.PathSegments...)
	segments = append(segments, parsed.Segments...)
	absolute := rt.AbsolutePath
	if len(parsed.Segments) > 0 {
		absolute = joinNative(rt.AbsolutePath, parsed.Segments)
	}

	rule := BypassRule{
		Root:         parsed.Namespace,
		Path:         namespacePath,
		InnerPath:    strings.Join(segments, "/"),
		AbsolutePath: absolute,
		Segments:     segments,
		Reason:       reason,
		Readable:     false,
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	rule.Index = len(m.rules)
	m.rules = append(m.rules, rule)
	return rule, nil
}

func (m *Manager) Delete(index int) (BypassRule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if index < 0 || index >= len(m.rules) {
		return BypassRule{}, fmt.Errorf("%w: %d", ErrBypassNotFound, index)
	}
	removed := m.rules[index]
	m.rules = append(m.rules[:index], m.rules[index+1:]...)
	removed.Index = index
	return removed, nil
}

func (m *Manager) RemoveRoot(namespace string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	kept := m.rules[:0]
	for _, rule := range m.rules {
		if rule.Root != namespace {
			kept = append(kept, rule)
		}
	}
	m.rules = kept
}
