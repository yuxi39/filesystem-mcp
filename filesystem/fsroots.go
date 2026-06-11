package filesystem

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yuxi39/filesystem-mcp/internal/bypass"
	"github.com/yuxi39/filesystem-mcp/internal/root"
)

var ToolPathList = &mcp.Tool{
	Name:        "path/list",
	Description: "List registered filesystem roots and bypass rules. Use this before resolving namespace paths.",
}

var ToolRootsAdd = &mcp.Tool{
	Name:        "path/roots/add",
	Description: "Register a concrete absolute filesystem path as a named root. Shell expressions such as ~, $HOME, %USERPROFILE%, and globs are not supported.",
}

var ToolRootsDel = &mcp.Tool{
	Name:        "path/roots/del",
	Description: "Remove a registered root namespace and its related bypass rules.",
}

var ToolBypassAdd = &mcp.Tool{
	Name:        "path/bypass/add",
	Description: "Block access to a namespace path under an existing root. The path must look like namespace:relative/path.",
}

var ToolBypassDel = &mcp.Tool{
	Name:        "path/bypass/del",
	Description: "Remove a bypass rule by index. Get indexes from path/list first.",
}

type PathListInput struct{}

type PathListOutput struct {
	Roots    []*root.Root        `json:"roots" jsonschema:"All registered filesystem roots. Use namespace:path when addressing files under these roots."`
	Bypasses []bypass.BypassRule `json:"bypasses" jsonschema:"Active bypass rules that block access to sensitive sub-paths."`
	Notes    []string            `json:"notes" jsonschema:"Path syntax reminders for the AI before calling filesystem tools."`
}

func HandlerPathList(ctx context.Context, req *mcp.CallToolRequest, in PathListInput) (*mcp.CallToolResult, PathListOutput, error) {
	return nil, PathListOutput{
		Roots:    root.Global.ListAll(),
		Bypasses: bypass.Global.List(),
		Notes: []string{
			"filesystem-mcp accepts concrete filesystem paths, not shell expressions.",
			"Unsupported examples: ~, $HOME, $env:USERPROFILE, %USERPROFILE%, %PATH%, $PATH, $(...), *.go.",
			"Use namespace paths such as odds:README.md after registering a root.",
		},
	}, nil
}

type RootsAddInput struct {
	Name string `json:"name" jsonschema:"Unique namespace for this root. If the namespace already exists, the new root replaces the old one."`
	Path string `json:"path" jsonschema:"Concrete absolute filesystem path or file:// URI. Do not pass shell expressions such as ~, $HOME, %USERPROFILE%, or globs."`
}

type RootsAddOutput struct {
	OK      string     `json:"ok" jsonschema:"Result status. The value is 'added' when registration succeeds."`
	Root    *root.Root `json:"root" jsonschema:"Registered root with normalized internal path and native absolute path."`
	Removed []string   `json:"removed" jsonschema:"Namespaces removed because the new root path covers their paths."`
}

func HandlerRootsAdd(ctx context.Context, req *mcp.CallToolRequest, in RootsAddInput) (*mcp.CallToolResult, RootsAddOutput, error) {
	rt, removed, err := root.Global.Add(in.Name, in.Path)
	if err != nil {
		return nil, RootsAddOutput{}, err
	}
	for _, namespace := range removed {
		bypass.Global.RemoveRoot(namespace)
	}
	return nil, RootsAddOutput{
		OK:      "added",
		Root:    rt,
		Removed: removed,
	}, nil
}

type RootsDelInput struct {
	Name string `json:"name" jsonschema:"Root namespace to delete. This also removes bypass rules owned by the namespace."`
}

type RootsDelOutput struct {
	OK   string     `json:"ok" jsonschema:"Result status. The value is 'deleted' when deletion succeeds."`
	Root *root.Root `json:"root" jsonschema:"Deleted root record."`
}

func HandlerRootsDel(ctx context.Context, req *mcp.CallToolRequest, in RootsDelInput) (*mcp.CallToolResult, RootsDelOutput, error) {
	rt, err := root.Global.Delete(in.Name)
	if err != nil {
		return nil, RootsDelOutput{}, err
	}
	bypass.Global.RemoveRoot(in.Name)
	return nil, RootsDelOutput{
		OK:   "deleted",
		Root: rt,
	}, nil
}

type BypassAddInput struct {
	Path   string `json:"path" jsonschema:"Namespace path to block, for example odds:secret or odds:.env. The root namespace must already exist."`
	Reason string `json:"reason" jsonschema:"Human-readable reason shown to the AI when access is denied."`
}

type BypassAddOutput struct {
	OK     string            `json:"ok" jsonschema:"Result status. The value is 'added' when the bypass rule is registered."`
	Bypass bypass.BypassRule `json:"bypass" jsonschema:"Registered bypass rule with normalized internal path and native absolute path."`
}

func HandlerBypassAdd(ctx context.Context, req *mcp.CallToolRequest, in BypassAddInput) (*mcp.CallToolResult, BypassAddOutput, error) {
	rule, err := bypass.Global.Add(in.Path, in.Reason, root.Global)
	if err != nil {
		return nil, BypassAddOutput{}, err
	}
	return nil, BypassAddOutput{
		OK:     "added",
		Bypass: rule,
	}, nil
}

type BypassDelInput struct {
	Index int `json:"index" jsonschema:"Index of the bypass rule to remove. Read indexes from path/list."`
}

type BypassDelOutput struct {
	OK     string            `json:"ok" jsonschema:"Result status. The value is 'deleted' when removal succeeds."`
	Bypass bypass.BypassRule `json:"bypass" jsonschema:"Removed bypass rule."`
}

func HandlerBypassDel(ctx context.Context, req *mcp.CallToolRequest, in BypassDelInput) (*mcp.CallToolResult, BypassDelOutput, error) {
	rule, err := bypass.Global.Delete(in.Index)
	if err != nil {
		return nil, BypassDelOutput{}, err
	}
	return nil, BypassDelOutput{
		OK:     "deleted",
		Bypass: rule,
	}, nil
}
