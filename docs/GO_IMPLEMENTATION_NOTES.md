# Go Implementation Notes

These notes translate the MCP design into practical Go pieces.

## Recommended First Dependency Choice

Use the official or widely used Go MCP SDK if you already have one selected.

If not, keep the core filesystem logic independent from MCP transport:

```txt
MCP JSON/tool layer
  -> internal service methods
  -> workspace/path safety
  -> fs operations
```

This keeps the filesystem logic testable even if the MCP library changes.

## Core Types

```go
type Root struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Writable bool   `json:"writable"`
}

type Config struct {
	Roots            []Root   `json:"roots"`
	DefaultIgnore    []string `json:"defaultIgnore"`
	MaxReadBytes     int64    `json:"maxReadBytes"`
	MaxSearchResults int      `json:"maxSearchResults"`
}

type ResolvedPath struct {
	RootName string
	RelPath  string
	AbsPath  string
	Writable bool
}
```

## Path Resolution

Path resolution is the most important part of the server.

Suggested function:

```go
func ResolvePath(cfg Config, input string, forWrite bool) (ResolvedPath, error)
```

Behavior:

1. Accept `rootName:relative/path`.
2. Optionally accept absolute paths if they are inside a configured root.
3. Clean the path with `filepath.Clean`.
4. Join relative paths to the selected root.
5. Convert to absolute path.
6. For existing paths, resolve symlinks with `filepath.EvalSymlinks`.
7. Ensure the final path is inside the configured root.
8. If `forWrite` is true, ensure the root is writable.

Important helper:

```go
func IsInsideRoot(rootAbs string, candidateAbs string) bool
```

Do not use simple string prefix checks without path boundary checks.

## Standard Tool Result Shape

For Go internals, return normal Go values and errors.

At the MCP boundary, convert errors to this:

```go
type ToolError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Path        string `json:"path,omitempty"`
	Recoverable bool   `json:"recoverable"`
}

type ErrorResult struct {
	OK    bool      `json:"ok"`
	Error ToolError `json:"error"`
}
```

Successful tools do not need an `ok` field unless the MCP client convention prefers it.

## `fs.read` Types

```go
type ReadInput struct {
	Path      string `json:"path"`
	StartLine *int   `json:"startLine,omitempty"`
	EndLine   *int   `json:"endLine,omitempty"`
	MaxBytes  *int64 `json:"maxBytes,omitempty"`
}

type ReadOutput struct {
	Path         string `json:"path"`
	AbsolutePath string `json:"absolutePath"`
	Encoding     string `json:"encoding"`
	LineStart    int    `json:"lineStart"`
	LineEnd      int    `json:"lineEnd"`
	TotalLines   int    `json:"totalLines"`
	SizeBytes    int64  `json:"sizeBytes"`
	SHA256       string `json:"sha256"`
	Content      string `json:"content"`
	Truncated    bool   `json:"truncated"`
}
```

Line numbers should be 1-based.

If `startLine` and `endLine` are absent, return from the beginning up to `maxBytes`.

## Hashing

Use SHA-256 for conflict checks.

```go
func FileSHA256(path string) (string, error)
func BytesSHA256(data []byte) string
```

Every read should return the current file hash.

Every modifying write to an existing file should optionally accept `expectedSha256`; for early safety, make it required.

## `fs.patch` Types

```go
type PatchInput struct {
	Op             string `json:"op"`
	Path           string `json:"path,omitempty"`
	Content        string `json:"content,omitempty"`
	CreateParents  bool   `json:"createParents,omitempty"`
	Overwrite      bool   `json:"overwrite,omitempty"`
	ExpectedSHA256 string `json:"expectedSha256,omitempty"`
	StartLine      int    `json:"startLine,omitempty"`
	EndLine        int    `json:"endLine,omitempty"`
	Patch          string `json:"patch,omitempty"`
}

type ChangedFile struct {
	Path       string `json:"path"`
	OldSHA256 string `json:"oldSha256,omitempty"`
	NewSHA256 string `json:"newSha256"`
	Diff       string `json:"diff"`
}

type PatchOutput struct {
	OK           bool          `json:"ok"`
	ChangedFiles []ChangedFile `json:"changedFiles"`
}
```

Milestone 1 operations:

```txt
create
replace_range
```

Leave `unified_diff` for later.

## Range Replacement

Suggested function:

```go
func ReplaceLineRange(original string, startLine int, endLine int, replacement string) (string, error)
```

Rules:

- Lines are 1-based.
- `startLine` must be >= 1.
- `endLine` must be >= `startLine`.
- Replacement should end with a newline if replacing full lines.
- Preserve original line ending style when practical.

Implementation idea:

1. Detect line ending: prefer `\r\n` if the file mostly uses CRLF.
2. Normalize to `\n` internally.
3. Split into lines.
4. Replace `[startLine-1:endLine]`.
5. Join with original line ending.

## Atomic Writes

Suggested function:

```go
func AtomicWriteFile(path string, data []byte, perm fs.FileMode) error
```

Behavior:

1. Create a temp file in the same directory.
2. Write data.
3. Sync file if practical.
4. Close file.
5. Rename temp file over target.

Same-directory rename makes the operation more reliable.

## Diff Generation

For the first version, use a small Go diff library instead of writing your own.

Candidate:

```txt
github.com/sergi/go-diff/diffmatchpatch
```

If you want fewer dependencies, return a simple line-based diff first.

The exact diff format matters less than consistently showing:

- old lines removed
- new lines added
- file path

## Search Text

Suggested input:

```go
type SearchTextInput struct {
	Root          string `json:"root"`
	Query         string `json:"query"`
	Mode          string `json:"mode"` // "literal" first, "regex" later
	CaseSensitive bool   `json:"caseSensitive"`
	ContextLines   int    `json:"contextLines"`
	Limit          int    `json:"limit"`
}
```

Suggested output:

```go
type SearchMatch struct {
	Path    string   `json:"path"`
	Line    int      `json:"line"`
	Column  int      `json:"column"`
	Preview string   `json:"preview"`
	Before  []string `json:"before,omitempty"`
	After   []string `json:"after,omitempty"`
}

type SearchTextOutput struct {
	Root      string        `json:"root"`
	Query     string        `json:"query"`
	Matches   []SearchMatch `json:"matches"`
	Truncated bool          `json:"truncated"`
}
```

Implementation notes:

- Walk directories with `filepath.WalkDir`.
- Skip ignored directories early.
- Skip files above a configurable max size.
- Skip binary-looking files.
- Use case-folded strings for case-insensitive search.
- Stop when total result limit is reached.

## Binary Detection

Simple first version:

```go
func LooksBinary(sample []byte) bool {
	return bytes.IndexByte(sample, 0) >= 0
}
```

Read the first 8KB and reject if it contains NUL bytes.

## Memory Append

Suggested input:

```go
type MemoryAppendInput struct {
	Root       string `json:"root"`
	Kind       string `json:"kind"`
	ModuleName string `json:"moduleName,omitempty"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Time       string `json:"time,omitempty"`
}
```

Project-level kind mapping:

```go
var projectMemoryFiles = map[string]string{
	"vision":       "memory/vision.md",
	"architecture": "memory/architecture.md",
	"decision":     "memory/decisions.md",
	"question":     "memory/questions.md",
	"build-log":    "memory/build-log.md",
}
```

Append format:

```md
## 2026-06-08 - Title

Body text...
```

For `module-run`, append JSONL instead of Markdown.

## Tests To Write First

Start with path safety tests.

```txt
ResolvePath accepts root-relative path
ResolvePath accepts absolute path inside root
ResolvePath rejects ../ escape
ResolvePath rejects unknown root
ResolvePath rejects write to readonly root
ResolvePath rejects symlink escape
```

Then edit safety tests:

```txt
Patch create refuses overwrite
Patch create creates parent directories when requested
Patch replace_range requires matching hash
Patch replace_range rejects stale hash
Patch replace_range preserves unrelated lines
```

Then read/search tests:

```txt
Read returns hash and line metadata
Read range is 1-based
Search text respects ignore rules
Search text returns context lines
Search text truncates at limit
```

## Best Development Order

1. Config loading.
2. Root/path resolution.
3. `fs.roots`.
4. `fs.list`.
5. `fs.read`.
6. `fs.search_text`.
7. `fs.patch(create)`.
8. `fs.patch(replace_range)`.
9. `memory.append`.

Do not start with MCP transport complexity. First make the internal package pass tests.

