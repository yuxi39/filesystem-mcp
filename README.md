# hellosekai filesystem MCP

**EN:** A lightweight, safe, and context-efficient filesystem MCP server for AI coding agents.

**ZH:** 一个轻量、安全、上下文高效的 filesystem MCP 服务器，专为 AI 编程助手设计。

---

## Quick Start

### Install

```bash
go install github.com/yuxi39/filesystem-mcp@latest
```

Or build from source:

```bash
git clone https://github.com/yuxi39/filesystem-mcp.git
cd filesystem-mcp
go build
```

### Configure in VS Code

```json
{
  "mcp": {
    "servers": {
      "hello-sekai-fs": {
        "command": "filesystem",
        "args": []
      }
    }
  }
}
```

---

## Current Status — v0.0.1

Roots management foundation.

### Implemented Tools

| Tool | Description |
|---|---|
| `roots/list` | List all workspace roots and bypass rules |
| `roots/add` | Register a new workspace root |
| `roots/del` | Remove a workspace root |
| `bypass/add` | Block access to a sub-path within a root |
| `bypass/del` | Remove a bypass rule |

### Path Format

```
<rootName>:<relative/path>
```

Example: `odds:hello/README.md` resolves to `F:\ODDS&ENDS\hello\README.md`.

### Session Integration

On startup, `roots/list` automatically discovers VS Code workspace folders and merges them as initial roots. No manual registration needed for workspace roots.

---

## Roadmap

### v0.1.0 — File Reading

| Tool | Description |
|---|---|
| `fs.stat` | File or directory metadata with optional hash |
| `fs.list` | List directory children (sorted, filtered) |
| `fs.read` | Read file by line range with byte limit |
| `fs.tree` | Recursive directory tree with depth limit |
| `fs.search_names` | Substring search over file/directory names |

### v0.2.0 — File Writing

| Tool | Description |
|---|---|
| `fs.diff` | Unified diff preview |
| `fs.patch` | Safe range replacement with SHA-256 conflict check |
| Dry-run mode | Preview before writing |
| Conflict detection | Reject writes when file changed since last read |

### v0.3.0 — Memory & Recovery

| Tool | Description |
|---|---|
| `memory.append` | Append structured notes |
| `memory.read` | Read stored notes |
| Session persistence | Survive server restart |
| Error recovery | Revert to last known good state |

---

## Design Principles

1. **Safe** — Respect workspace boundaries, reject path traversal and subdirectory conflicts.
2. **Structured** — Return JSON with clear schemas, not prose.
3. **Concise** — Support line ranges, byte limits, truncation notices.
4. **Auditable** — Support dry-run, diffs, and hash verification.
5. **Recoverable** — Conflict detection prevents overwriting user edits.

---

## License

MIT

## Tool: `fs.stat`

Return metadata for one file or directory.

### Input

```json
{
  "path": "odds:hello/README.md",
  "hash": true
}
```

### Output

```json
{
  "path": "odds:hello/README.md",
  "absolutePath": "F:\\ODDS&ENDS\\hello\\README.md",
  "kind": "file",
  "sizeBytes": 9123,
  "modifiedAt": "2026-06-08T12:00:00+08:00",
  "sha256": "..."
}
```

### Why Codex Needs It

Hashes allow conflict-safe writes.

## Tool: `fs.read`

Read a file, optionally by line range.

### Input

```json
{
  "path": "odds:hello/README.md",
  "startLine": 1,
  "endLine": 80,
  "maxBytes": 40000
}
```

### Output

```json
{
  "path": "odds:hello/README.md",
  "absolutePath": "F:\\ODDS&ENDS\\hello\\README.md",
  "encoding": "utf-8",
  "lineStart": 1,
  "lineEnd": 80,
  "totalLines": 240,
  "sizeBytes": 9123,
  "sha256": "...",
  "content": "# hello sekai\n\n...",
  "truncated": false
}
```

### Behavior

- Default to UTF-8.
- Detect UTF-8 BOM.
- Reject binary files unless `binaryMode` is explicitly requested later.
- Include file hash.
- Include total lines.
- If truncated, say where and why.

### Why Codex Needs It

The agent often needs only a slice of a file, not the whole thing.

## Tool: `fs.search_names`

Search file and directory names.

### Input

```json
{
  "root": "odds:hello",
  "query": "memory",
  "kind": "any",
  "limit": 50
}
```

### Output

```json
{
  "root": "odds:hello",
  "query": "memory",
  "matches": [
    {
      "path": "odds:hello/memory",
      "kind": "directory"
    },
    {
      "path": "odds:hello/src/memory/memoryStore.ts",
      "kind": "file"
    }
  ],
  "truncated": false
}
```

### Behavior

- Case-insensitive by default on Windows.
- Respect ignore rules.
- Prefer substring search first.
- Regex can be added later.

## Tool: `fs.search_text`

Search text contents.

### Input

```json
{
  "root": "odds:hello",
  "query": "module registry",
  "mode": "literal",
  "caseSensitive": false,
  "contextLines": 2,
  "limit": 50
}
```

### Output

```json
{
  "root": "odds:hello",
  "query": "module registry",
  "matches": [
    {
      "path": "odds:hello/README.md",
      "line": 83,
      "column": 5,
      "preview": "The registry reads module manifests and answers:",
      "before": [
        "### Step 4: Build The Registry"
      ],
      "after": [
        "",
        "- What modules exist?"
      ]
    }
  ],
  "truncated": false
}
```

### Behavior

- Use Go implementation first.
- Optionally use `rg` if available later.
- Skip binary files.
- Respect ignore rules.
- Limit per-file matches to avoid flooding.

### Why Codex Needs It

This is one of the highest-value tools. It lets the agent locate definitions, TODOs, duplicated logic, and related files quickly.

## Tool: `fs.diff`

Preview a proposed full-file replacement or patch.

### Input: Full Replacement

```json
{
  "path": "odds:hello/README.md",
  "expectedSha256": "...",
  "newContent": "# new content\n"
}
```

### Input: Range Replacement

```json
{
  "path": "odds:hello/README.md",
  "expectedSha256": "...",
  "replace": {
    "startLine": 10,
    "endLine": 20,
    "content": "replacement\n"
  }
}
```

### Output

```json
{
  "path": "odds:hello/README.md",
  "ok": true,
  "changed": true,
  "oldSha256": "...",
  "newSha256": "...",
  "diff": "--- README.md\n+++ README.md\n..."
}
```

### Why Codex Needs It

The agent can show or inspect the change before applying it.

## Tool: `fs.patch`

Apply a scoped edit.

### Input: Create File

```json
{
  "op": "create",
  "path": "odds:hello/modules/hello.sekai/module.json",
  "content": "{\n  \"name\": \"hello.sekai\"\n}\n",
  "createParents": true,
  "overwrite": false
}
```

### Input: Replace Range

```json
{
  "op": "replace_range",
  "path": "odds:hello/README.md",
  "expectedSha256": "...",
  "startLine": 10,
  "endLine": 20,
  "content": "replacement\n"
}
```

### Input: Apply Unified Diff

```json
{
  "op": "unified_diff",
  "patch": "*** Begin Patch\n*** Update File: hello/README.md\n@@\n-old\n+new\n*** End Patch\n"
}
```

### Output

```json
{
  "ok": true,
  "changedFiles": [
    {
      "path": "odds:hello/README.md",
      "oldSha256": "...",
      "newSha256": "...",
      "diff": "--- README.md\n+++ README.md\n..."
    }
  ]
}
```

### Required Safety

- Reject writes outside roots.
- Reject overwrite if `overwrite` is false.
- Reject update if `expectedSha256` does not match.
- Return a conflict with current hash and current modified time.
- Preserve line endings if possible.
- Use atomic write: write temp file, fsync if practical, then rename.

### Why Codex Needs It

This is the main editing tool. The most important property is not raw power. The most important property is that failed edits are understandable and recoverable.

## Tool: `memory.append`

Append a structured memory note to a known memory file.

This is separate from generic file editing because project memory should be easy and safe to update.

### Input

```json
{
  "root": "odds:hello",
  "kind": "build-log",
  "title": "Created filesystem MCP design draft",
  "body": "Wrote the first design for a Go MCP server focused on safe filesystem collaboration.",
  "time": "2026-06-08T12:00:00+08:00"
}
```

### Output

```json
{
  "ok": true,
  "path": "odds:hello/memory/build-log.md",
  "appendedBytes": 164,
  "newSha256": "..."
}
```

### Supported Kinds

```txt
vision
architecture
decision
question
build-log
module-intent
module-interface
module-changelog
module-run
```

### Behavior

For project-level memory:

```txt
memory/vision.md
memory/architecture.md
memory/decisions.md
memory/questions.md
memory/build-log.md
```

For module-level memory, require `moduleName`.

```txt
modules/<moduleName>/memory/intent.md
modules/<moduleName>/memory/interface.md
modules/<moduleName>/memory/changelog.md
modules/<moduleName>/memory/runs.jsonl
```

### Why Codex Needs It

If memory writing is easy, the agent can keep the software's design history alive while coding.

## Error Model

All tools should return errors in a structured format.

```json
{
  "ok": false,
  "error": {
    "code": "PATH_OUTSIDE_ROOT",
    "message": "Path resolves outside configured workspace root.",
    "path": "odds:../secret.txt",
    "recoverable": true
  }
}
```

Recommended error codes:

```txt
PATH_OUTSIDE_ROOT
PATH_NOT_FOUND
PATH_IS_DIRECTORY
PATH_IS_FILE
PERMISSION_DENIED
FILE_TOO_LARGE
BINARY_FILE
INVALID_ENCODING
HASH_MISMATCH
PATCH_CONFLICT
INVALID_PATCH
LIMIT_EXCEEDED
IGNORED_PATH
UNKNOWN_ROOT
INVALID_ARGUMENT
INTERNAL_ERROR
```

## Ignore Rules

Default ignored names:

```txt
.git
node_modules
dist
build
.next
.turbo
target
bin
obj
coverage
.venv
__pycache__
```

The MCP should also read `.gitignore` later, but this is not required for Milestone 1.

## Security Rules

Minimum safety rules:

- Resolve all paths with `filepath.Clean`.
- Convert relative paths to absolute paths under a known root.
- Resolve symlinks before allowing writes.
- Reject writes that escape the root after symlink resolution.
- Do not expose arbitrary environment variables.
- Do not run shell commands in this MCP.
- Do not delete recursively in Milestone 1.
- Do not implement move or rename in Milestone 1.

This MCP is filesystem-only. Shell execution should be a separate tool with stricter approval rules.

## Go Package Layout

Suggested layout:

```txt
filesystem/
  README.md
  go.mod
  cmd/
    hellosekai-fs-mcp/
      main.go
  internal/
    config/
      config.go
    mcp/
      server.go
      tools.go
    workspace/
      roots.go
      path.go
      ignore.go
    fsops/
      list.go
      tree.go
      stat.go
      read.go
      search.go
      diff.go
      patch.go
    memory/
      append.go
    text/
      encoding.go
      lines.go
      hash.go
      unified_diff.go
```

## Implementation Phases

### Phase 1: Read-Only Core

Implement:

- `fs.roots`
- `fs.list`
- `fs.tree`
- `fs.stat`
- `fs.read`
- `fs.search_names`
- `fs.search_text`

This phase is already useful.

### Phase 2: Safe Writes

Implement:

- `fs.diff`
- `fs.patch` with `create`
- `fs.patch` with `replace_range`

Require `expectedSha256` for modifying existing files.

### Phase 3: Memory Convenience

Implement:

- `memory.append`

This lets the agent update project memory without hand-editing Markdown every time.

### Phase 4: Better Editing

Add:

- unified diff patching
- multi-file patch transaction
- format-preserving line endings
- `.gitignore` support
- file summaries

## What Not To Build Yet

Do not build these in the first version:

- recursive delete
- arbitrary shell execution
- package manager execution
- Git commands
- binary file editing
- file watching
- background indexing
- vector search
- web search

These can become separate MCP servers later.

## Best First Implementation Target

Build this first:

```txt
fs.roots
fs.list
fs.read
fs.search_text
fs.patch(create)
fs.patch(replace_range)
```

With only those six capabilities, Codex can already:

- inspect a project
- read the relevant files
- find references
- create new files
- make conflict-safe edits
- explain what changed

That is the smallest filesystem MCP worth having.

