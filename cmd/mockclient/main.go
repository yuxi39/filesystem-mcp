package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

// This Go MCP SDK (v1.6.1) uses NDJSON (newline-delimited JSON), NOT Content-Length headers.
// Each JSON-RPC message is a single line terminated by '\n'.

func main() {
	serverPath := "bin/fs.exe"
	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		serverPath = "../../bin/fs.exe"
	}

	cmd := exec.Command(serverPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("stdin pipe failed: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("stdout pipe failed: %v", err)
	}

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
	defer cmd.Process.Kill()

	fmt.Println("=== Server started ===")
	fmt.Println()

	// --- 1. Initialize ---
	sendMsg(stdin, 1, "initialize", map[string]any{
		"protocolVersion": "2025-11-25",
		"clientInfo": map[string]string{
			"name":    "mock-client",
			"version": "0.0.1",
		},
	})
	recvMsg(stdout, stdin, "initialize response")

	// --- 2. Initialized notification ---
	sendNotify(stdin, "initialized", nil)
	fmt.Println(">>> SENT: initialized notification")
	fmt.Println()

	// --- 3. List tools ---
	sendMsg(stdin, 2, "tools/list", map[string]any{})
	recvMsg(stdout, stdin, "tools/list response")

	// --- 4. roots/add: register a root ---
	sendMsg(stdin, 3, "tools/call", map[string]any{
		"name": "roots/add",
		"arguments": map[string]any{
			"name": "odds",
			"path": `F:\ODDS&ENDS`,
		},
	})
	recvMsg(stdout, stdin, "roots/add response")

	// --- 5. roots/list: see all roots (should show 'odds') ---
	sendMsg(stdin, 4, "tools/call", map[string]any{
		"name":      "roots/list",
		"arguments": map[string]any{},
	})
	recvMsg(stdout, stdin, "roots/list (after add) response")

	// --- 6. bypass/add: block a sub-path ---
	sendMsg(stdin, 5, "tools/call", map[string]any{
		"name": "bypass/add",
		"arguments": map[string]any{
			"path":   "odds:secret",
			"reason": "Sensitive directory",
		},
	})
	recvMsg(stdout, stdin, "bypass/add response")

	// --- 7. roots/list: see roots + bypasses ---
	sendMsg(stdin, 6, "tools/call", map[string]any{
		"name":      "roots/list",
		"arguments": map[string]any{},
	})
	recvMsg(stdout, stdin, "roots/list (with bypass) response")

	// --- 8. roots/del: remove the root ---
	sendMsg(stdin, 7, "tools/call", map[string]any{
		"name": "roots/del",
		"arguments": map[string]any{
			"name": "odds",
		},
	})
	recvMsg(stdout, stdin, "roots/del response")

	// --- 9. roots/list: confirm empty ---
	sendMsg(stdin, 8, "tools/call", map[string]any{
		"name":      "roots/list",
		"arguments": map[string]any{},
	})
	recvMsg(stdout, stdin, "roots/list (after del) response")

	fmt.Println()
	fmt.Println("=== Done ===")
}

// sendMsg writes a JSON-RPC request as a single NDJSON line.
func sendMsg(w io.Writer, id int, method string, params any) {
	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}
	data, _ := json.Marshal(msg)
	fmt.Printf(">>> SEND (id=%d, method=%s):\n%s\n\n", id, method, string(data))

	if _, err := w.Write(append(data, '\n')); err != nil {
		log.Fatalf("write failed: %v", err)
	}
}

// sendNotify writes a JSON-RPC notification (no id) as a single NDJSON line.
func sendNotify(w io.Writer, method string, params any) {
	msg := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if params != nil {
		msg["params"] = params
	}
	data, _ := json.Marshal(msg)
	if _, err := w.Write(append(data, '\n')); err != nil {
		log.Fatalf("write failed: %v", err)
	}
}

// rawReply sends a raw JSON-RPC response back to the server.
func rawReply(w io.Writer, id any, result any) {
	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	data, _ := json.Marshal(msg)
	fmt.Printf(">>> REPLY (id=%v):\n%s\n\n", id, string(data))
	if _, err := w.Write(append(data, '\n')); err != nil {
		log.Fatalf("write failed: %v", err)
	}
}

// recvMsg reads NDJSON messages from the server until we get a response
// (a message with "id" but no "method") matching the expected label.
// If the server sends a request (has "method" + "id"), it automatically replies
// with a canned response to keep the conversation going.
func recvMsg(r io.Reader, w io.Writer, label string) {
	br := bufio.NewReader(r)

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			log.Fatalf("read failed (%s): %v", label, err)
		}
		line = line[:len(line)-1] // trim trailing newline

		// Pretty-print the raw message
		var raw any
		if err := json.Unmarshal([]byte(line), &raw); err == nil {
			formatted, _ := json.MarshalIndent(raw, "", "  ")
			fmt.Printf("<<< %s:\n%s\n\n", label, string(formatted))
		} else {
			fmt.Printf("<<< %s (raw):\n%s\n\n", label, line)
		}

		// Check if this is a request from the server (has both "method" and "id")
		var msgMap map[string]any
		if err := json.Unmarshal([]byte(line), &msgMap); err != nil {
			return
		}
		_, hasMethod := msgMap["method"]
		msgID, hasID := msgMap["id"]

		if hasMethod && hasID {
			// Server sent us a request — auto-reply
			switch msgMap["method"] {
			case "roots/list":
				rawReply(w, msgID, map[string]any{
					"roots": []map[string]any{
						{
							"name": "filesystem",
							"uri":  "file:///f%3A/ODDS%26ENDS/filesystem",
						},
					},
				})
			case "ping":
				rawReply(w, msgID, map[string]any{})
			default:
				rawReply(w, msgID, map[string]any{})
			}
			// Continue reading — the next message should be the actual response
			continue
		}

		// It's a response (has "id" but no "method") — we're done
		return
	}
}
