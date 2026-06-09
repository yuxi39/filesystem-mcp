package roots

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// BypassRule 表示一条路径排除规则。
type BypassRule struct {
	Path   string `json:"path"   jsonschema:"Blocked absolute path"`
	Root   string `json:"root"   jsonschema:"Owner root name"`
	Reason string `json:"reason" jsonschema:"Why this path is blocked"`
}

// BypassManager 管理 bypass 规则，独立于 roots。
type BypassManager struct {
	mu    sync.RWMutex
	rules []BypassRule
}

// BypassGlobal 是全局 bypass 管理器。
var BypassGlobal = &BypassManager{}

// Add 添加一条 bypass 规则。
// path 是相对于 root 的命名空间路径，如 "odds:secret"。
func (bm *BypassManager) Add(namespacePath, reason string) error {
	name, rel, ok := strings.Cut(namespacePath, ":")
	if !ok {
		return fmt.Errorf("invalid bypass path %q: missing root name", namespacePath)
	}

	root, ok := Global.Get(name)
	if !ok {
		return fmt.Errorf("root %q not found", name)
	}

	rel = strings.TrimLeft(rel, "/\\")
	abs := filepath.Join(root.Path, filepath.FromSlash(rel))

	if !isWithinRoot(root.Path, abs) {
		return fmt.Errorf("bypass path %q escapes root %q", namespacePath, root.Path)
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.rules = append(bm.rules, BypassRule{
		Path:   abs,
		Root:   name,
		Reason: reason,
	})
	return nil
}

// Del 按索引删除一条 bypass 规则（从 All 返回的列表中的位置）。
func (bm *BypassManager) Del(index int) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	if index < 0 || index >= len(bm.rules) {
		return fmt.Errorf("bypass index %d out of range", index)
	}
	bm.rules = append(bm.rules[:index], bm.rules[index+1:]...)
	return nil
}

// All 返回所有 bypass 规则的副本。
func (bm *BypassManager) All() []BypassRule {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	out := make([]BypassRule, len(bm.rules))
	copy(out, bm.rules)
	return out
}

// IsBlocked 检查 rootPath 下的 rel 路径是否被 bypass 命中。
func (bm *BypassManager) IsBlocked(rootPath, rel string) bool {
	_, ok := bm.match(rootPath, rel)
	return ok
}

// Match 返回命中的第一条 bypass 规则。
func (bm *BypassManager) Match(rootPath, rel string) BypassRule {
	r, _ := bm.match(rootPath, rel)
	return r
}

func (bm *BypassManager) match(rootPath, rel string) (BypassRule, bool) {
	target := filepath.Join(rootPath, filepath.FromSlash(rel))

	bm.mu.RLock()
	defer bm.mu.RUnlock()
	for _, rule := range bm.rules {
		if isWithinRoot(rule.Path, target) || target == rule.Path {
			return rule, true
		}
	}
	return BypassRule{}, false
}
