package main

import (
	"context"
	_ "embed"
	"encoding/base64"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	icons := mcpIcons()
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "filesystem",
		Title:   "Hello Sekai Filesystem",
		Version: "v0.0.1",
		Icons:   icons,
	}, &mcp.ServerOptions{
		Instructions: "A small filesystem MCP server for Hello Sekai.",
	})

	regist(server)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

//go:embed mcp.png
var mcpIconData []byte

func mcpIcons() []mcp.Icon {
	return []mcp.Icon{
		{
			Source:   "data:image/png;base64," + base64.StdEncoding.EncodeToString(mcpIconData),
			MIMEType: "image/png",
			Sizes:    []string{"48x48"},
			Theme:    mcp.IconThemeLight,
		},
	}
}
