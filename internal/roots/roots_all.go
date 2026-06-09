package roots

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

type Root struct {
	Name string `json:"name" jsonschema:"Root name used as namespace prefix, e.g. 'odds'"`
	Path string `json:"path" jsonschema:"Absolute filesystem path, e.g. D:\\ODDS&ENDS"`
}

type Manager struct {
	mu  sync.RWMutex
	rts map[string]Root
}

var Global = &Manager{
	rts: make(map[string]Root),
}

func (m *Manager) Add(name, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rts[name]; exists {
		return fmt.Errorf("root %q already exists", name)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("root %q path %q: %w", name, path, err)
	}

	// 检查是否在已有 root 之下，或覆盖了已有 root
	for _, existing := range m.rts {
		if isWithinRoot(existing.Path, abs) {
			return fmt.Errorf("path %q is a subdirectory of root %q (%s)",
				abs, existing.Name, existing.Path)
		}
		if isWithinRoot(abs, existing.Path) {
			return fmt.Errorf("root %q (%s) is already covered by new path %q",
				existing.Name, existing.Path, abs)
		}
	}

	m.rts[name] = Root{Name: name, Path: abs}
	return nil
}

func (m *Manager) Del(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.rts[name]; !exists {
		return fmt.Errorf("root %q not found", name)
	}
	delete(m.rts, name)
	return nil
}

func (m *Manager) Get(name string) (Root, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.rts[name]
	return r, ok
}

func (m *Manager) All() []Root {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Root, 0, len(m.rts))
	for _, r := range m.rts {
		out = append(out, r)
	}
	return out
}

func (m *Manager) Resolve(namespacePath string) (string, string, error) {
	name, rel, ok := strings.Cut(namespacePath, ":")
	if !ok {
		return "", "", fmt.Errorf("invalid path %q: missing root name (format: name:rel/path)", namespacePath)
	}

	m.mu.RLock()
	root, exists := m.rts[name]
	m.mu.RUnlock()
	if !exists {
		return "", "", fmt.Errorf("root %q not found", name)
	}

	// 先检查 bypass
	if BypassGlobal.IsBlocked(root.Path, rel) {
		rule := BypassGlobal.Match(root.Path, rel)
		return "", "", fmt.Errorf("path %q is blocked by bypass rule %q (reason: %s)",
			namespacePath, rule.Path, rule.Reason)
	}

	rel = strings.TrimLeft(rel, "/\\")
	if rel == "" {
		return namespacePath, root.Path, nil
	}

	abs := filepath.Join(root.Path, filepath.FromSlash(rel))

	// 安全检查：防止 .. 穿越
	if !isWithinRoot(root.Path, abs) {
		return "", "", fmt.Errorf("path %q escapes root %q", namespacePath, root.Path)
	}

	return namespacePath, abs, nil
}

// isWithinRoot 检查 candidate 是否在 rootPath 之下。两边都必须是绝对路径。
func isWithinRoot(rootPath, candidate string) bool {
	rel, err := filepath.Rel(rootPath, candidate)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}
