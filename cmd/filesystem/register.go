package main

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yuxi39/filesystem-mcp/filesystem"
)

func regist(server *mcp.Server) {
	mcp.AddTool(server, filesystem.ToolRootsList, filesystem.HandlerRootsList)
	mcp.AddTool(server, filesystem.ToolRootsAdd, filesystem.HandlerRootsAdd)
	mcp.AddTool(server, filesystem.ToolRootsDel, filesystem.HandlerRootsDel)
	mcp.AddTool(server, filesystem.ToolBypassAdd, filesystem.HandlerBypassAdd)
	mcp.AddTool(server, filesystem.ToolBypassDel, filesystem.HandlerBypassDel)
}
