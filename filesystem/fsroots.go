package filesystem

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yuxi39/filesystem-mcp/internal/roots"
)

// Tool descriptors for roots management.
// Handlers and types are defined in internal/roots.

var ToolRootsList = &mcp.Tool{
	Name:        "roots/list",
	Description: "List all configured workspace roots and active bypass rules",
}

var ToolRootsAdd = &mcp.Tool{
	Name:        "roots/add",
	Description: "Register a new workspace root with a name prefix. Path must be absolute. Rejects subdirectory conflicts.",
}

var ToolRootsDel = &mcp.Tool{
	Name:        "roots/del",
	Description: "Remove a registered root by name. Future accesses using this namespace will fail.",
}

var ToolBypassAdd = &mcp.Tool{
	Name:        "bypass/add",
	Description: "Block access to a sub-path within a root. Agent will be blocked with the reason message.",
}

var ToolBypassDel = &mcp.Tool{
	Name:        "bypass/del",
	Description: "Remove a bypass rule by index (from roots/list bypasses list).",
}

// Exported handler references so register.go can use typed AddTool.
var (
	HandlerRootsList = roots.ToolRootsList
	HandlerRootsAdd  = roots.ToolRootsAdd
	HandlerRootsDel  = roots.ToolRootsDel
	HandlerBypassAdd = roots.ToolBypassAdd
	HandlerBypassDel = roots.ToolBypassDel
)
