package roots

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yuxi39/filesystem-mcp/internal/path"
)

// --- roots/list ---

type RootsListInput struct{}

type RootsListOutput struct {
	Roots    []Root       `json:"roots"    jsonschema:"All registered workspace roots (manual + session)"`
	Bypasses []BypassRule `json:"bypasses" jsonschema:"All active bypass rules that block access to sub-paths"`
}

func ToolRootsList(ctx context.Context, req *mcp.CallToolRequest, _ RootsListInput) (*mcp.CallToolResult, RootsListOutput, error) {
	// Collect manually added roots.
	all := Global.All()

	// Collect session roots from the MCP client (e.g. VS Code workspace folders).
	if req.Session != nil {
		sres, err := req.Session.ListRoots(ctx, nil)
		if err == nil {
			for _, sr := range sres.Roots {
				// Convert file:// URI to native path.
				nativePath, err := path.URIToPath(sr.URI)
				if err != nil {
					nativePath = sr.URI
				}

				// Check bypass — blocked session roots are still listed
				// but the agent will get an error if it tries to access them.
				blocked := false
				for _, b := range BypassGlobal.All() {
					if isWithinRoot(b.Path, nativePath) || b.Path == nativePath {
						blocked = true
						break
					}
				}
				if blocked {
					continue
				}

				// Register as initial root if not already present.
				found := false
				for _, r := range all {
					if r.Name == sr.Name {
						found = true
						break
					}
				}
				if !found {
					Global.Add(sr.Name, nativePath) // add to Global for Resolve etc.
					all = append(all, Root{Name: sr.Name, Path: nativePath})
				}
			}
		}
	}

	return nil, RootsListOutput{
		Roots:    all,
		Bypasses: BypassGlobal.All(),
	}, nil
}

// --- roots/add ---

type RootsAddInput struct {
	Name string `json:"name" jsonschema:"Root name for namespace prefix, e.g. 'odds'"`
	Path string `json:"path" jsonschema:"Absolute filesystem path, e.g. D:\\ODDS&ENDS"`
}

type RootsAddOutput struct {
	OK   string `json:"ok"   jsonschema:"Result status: 'added' or error"`
	Name string `json:"name" jsonschema:"Registered root name"`
	Path string `json:"path" jsonschema:"Registered absolute path"`
}

func ToolRootsAdd(_ context.Context, _ *mcp.CallToolRequest, in RootsAddInput) (*mcp.CallToolResult, RootsAddOutput, error) {
	if in.Name == "" {
		return nil, RootsAddOutput{}, fmt.Errorf("name is required")
	}
	if !filepath.IsAbs(in.Path) {
		return nil, RootsAddOutput{}, fmt.Errorf("path must be absolute, got %q", in.Path)
	}

	// Check bypass — if this root's path itself is blocked, reject.
	for _, b := range BypassGlobal.All() {
		if isWithinRoot(b.Path, in.Path) || b.Path == in.Path {
			return nil, RootsAddOutput{}, fmt.Errorf("path %q is blocked by bypass rule for %q (reason: %s)",
				in.Path, b.Path, b.Reason)
		}
	}

	if err := Global.Add(in.Name, in.Path); err != nil {
		return nil, RootsAddOutput{}, err
	}
	return nil, RootsAddOutput{OK: "added", Name: in.Name, Path: in.Path}, nil
}

// --- roots/del ---

type RootsDelInput struct {
	Name string `json:"name" jsonschema:"Root name to remove"`
}

type RootsDelOutput struct {
	OK   string `json:"ok"   jsonschema:"Result status: 'deleted' or error"`
	Name string `json:"name" jsonschema:"Removed root name"`
}

func ToolRootsDel(_ context.Context, _ *mcp.CallToolRequest, in RootsDelInput) (*mcp.CallToolResult, RootsDelOutput, error) {
	if err := Global.Del(in.Name); err != nil {
		return nil, RootsDelOutput{}, err
	}
	return nil, RootsDelOutput{OK: "deleted", Name: in.Name}, nil
}

// --- bypass/add ---

type BypassAddInput struct {
	Path   string `json:"path"   jsonschema:"Namespace path to block, e.g. 'odds:secret'"`
	Reason string `json:"reason" jsonschema:"Why this path is blocked, shown to agent on access"`
}

type BypassAddOutput struct {
	OK     string `json:"ok"     jsonschema:"Result status: 'added' or error"`
	Path   string `json:"path"   jsonschema:"Blocked namespace path"`
	Reason string `json:"reason" jsonschema:"Block reason"`
}

func ToolBypassAdd(_ context.Context, _ *mcp.CallToolRequest, in BypassAddInput) (*mcp.CallToolResult, BypassAddOutput, error) {
	if err := BypassGlobal.Add(in.Path, in.Reason); err != nil {
		return nil, BypassAddOutput{}, err
	}
	return nil, BypassAddOutput{OK: "added", Path: in.Path, Reason: in.Reason}, nil
}

// --- bypass/del ---

type BypassDelInput struct {
	Index int `json:"index" jsonschema:"Index from the bypasses list returned by roots/list (0-based)"`
}

type BypassDelOutput struct {
	OK    string `json:"ok"    jsonschema:"Result status: 'deleted' or error"`
	Index int    `json:"index" jsonschema:"Removed bypass index"`
}

func ToolBypassDel(_ context.Context, _ *mcp.CallToolRequest, in BypassDelInput) (*mcp.CallToolResult, BypassDelOutput, error) {
	if err := BypassGlobal.Del(in.Index); err != nil {
		return nil, BypassDelOutput{}, err
	}
	return nil, BypassDelOutput{OK: "deleted", Index: in.Index}, nil
}
