package main

import (
	"context"
	_ "embed"
	"encoding/base64"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yuxi39/filesystem-mcp/filesystem"
)

//go:embed filesystem.png
var mcpIconData []byte

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "filesystem",
		Title:   "Hello Sekai Filesystem",
		Version: "v0.0.1",
		Icons: []mcp.Icon{
			{
				Source:   "data:image/png;base64," + base64.StdEncoding.EncodeToString(mcpIconData),
				MIMEType: "image/png",
				Sizes:    []string{"48x48"},
				Theme:    mcp.IconThemeLight,
			},
		},
	}, &mcp.ServerOptions{
		Instructions: "A small filesystem MCP server for Hello Sekai.",
	})

	mcp.AddTool(server, filesystem.ToolRootsList, filesystem.HandlerRootsList)
	mcp.AddTool(server, filesystem.ToolRootsAdd, filesystem.HandlerRootsAdd)
	mcp.AddTool(server, filesystem.ToolRootsDel, filesystem.HandlerRootsDel)
	mcp.AddTool(server, filesystem.ToolBypassAdd, filesystem.HandlerBypassAdd)
	mcp.AddTool(server, filesystem.ToolBypassDel, filesystem.HandlerBypassDel)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
