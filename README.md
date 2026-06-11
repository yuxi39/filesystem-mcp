# filesystem-mcp

一个为 AI 编程助手设计的 filesystem MCP server。

项目目前处于验证阶段。目标不是马上做出一个功能很多的文件管理器，而是先把路径模型、root 边界、bypass 规则和 MCP 工具契约跑通。现在很多代码还会继续改版，这是正常状态：越实践，边界会越清楚，属于人的设计判断也会越来越多。

## 当前结论

filesystem-mcp 接收的是**明确的文件系统路径**，不是 shell 表达式。

支持方向：

- namespace path: `odds:README.md`
- Windows absolute path: `F:\ODDS&ENDS\hello`
- Windows slash path: `F:/ODDS&ENDS/hello`
- Windows UNC path: `\\wsl$\Ubuntu\home\user`
- Unix absolute path: `/etc/cron.d`
- file URI: `file:///f%3A/ODDS%26ENDS/hello%20world`
- explicit relative path: `./README.md` 或 `.\README.md`

不支持方向：

- `~`
- `$HOME`
- `$env:USERPROFILE`
- `%USERPROFILE%`
- `$PATH` / `%PATH%`
- `$(...)`
- glob，比如 `*.go`

这些表达式应该由 agent 或用户先解析成明确路径，再传给 filesystem-mcp。

## 安装

```bash
go install github.com/yuxi39/filesystem-mcp@latest
```

从源码运行：

```bash
git clone https://github.com/yuxi39/filesystem-mcp.git
cd filesystem-mcp
go run .
```

## VS Code MCP 配置

用户级配置文件：

- Windows: `%APPDATA%\Code\User\mcp.json`
- macOS: `~/Library/Application Support/Code/User/mcp.json`
- Linux: `~/.config/Code/User/mcp.json`

示例：

```json
{
  "mcp": {
    "servers": {
      "filesystem": {
        "command": "filesystem-mcp",
        "args": []
      }
    }
  }
}
```

开发时也可以直接指向源码：

```json
{
  "mcp": {
    "servers": {
      "filesystem": {
        "command": "go",
        "args": ["run", "."],
        "cwd": "F:\\ODDS&ENDS\\filesystem"
      }
    }
  }
}
```

## 已实现工具

### `path/list`

列出当前注册的 roots 和 bypass rules。

输出里会包含给 agent 的路径规范提醒，例如不要把 `$HOME`、`%USERPROFILE%`、`*.go` 直接传给 filesystem-mcp。

### `path/roots/add`

注册一个 root。

输入示例：

```json
{
  "name": "odds",
  "path": "F:\\ODDS&ENDS"
}
```

规则：

- `name` 是 namespace。
- namespace 冲突时，新 root 覆盖旧 root。
- root path 必须是明确的绝对路径或 file URI。
- 不接受 shell 表达式。
- 如果新 root 覆盖已有子 root，会移除被覆盖的 namespace。

### `path/roots/del`

删除一个 root，并删除它关联的 bypass rules。

输入示例：

```json
{
  "name": "odds"
}
```

### `path/bypass/add`

添加 bypass rule，阻止 agent 访问某个 root 下的子路径。

输入示例：

```json
{
  "path": "odds:secret",
  "reason": "Contains sensitive credentials"
}
```

### `path/bypass/del`

按 index 删除 bypass rule。index 来自 `path/list`。

输入示例：

```json
{
  "index": 0
}
```

## 当前内部结构

### `internal/innerpath`

负责把用户输入路径解析成平台无关的内部路径。

核心输出：

```go
type Path struct {
    Kind      PathKind
    Namespace string
    Segments  []string
}
```

示例：

```go
F:\ODDS&ENDS\hello world\中文\日本語
```

会被解析成：

```go
Path{
    Kind: PathWinDrive,
    Segments: []string{"f:", "ODDS&ENDS", "hello world", "中文", "日本語"},
}
```

### `internal/root`

负责 root 管理和 root tree。

`RootTree` 只关心已经规范化后的 path segments。它维护的约束是：

- root node 不能有子节点。
- 无用分支会被删除。
- 已有父 root 时拒绝插入子 root。
- 插入父 root 时可以替换已有子 root。
- namespace 来自 RootManager，不从路径最后一段推断。

### `internal/bypass`

负责 bypass rule 管理。

bypass path 必须使用 namespace path，例如：

```txt
odds:secret
```

### `internal/list` / `internal/stat`

目前还是占位包。下一步会开始实现只读文件能力。

## 当前验证状态

已经完成：

- MCP server 能启动。
- MCP tools 能注册。
- `innerpath` 路径分类和解析。
- Windows / Unix native path 构建分离。
- root tree 插入、匹配、删除。
- RootManager 基础实现。
- bypass manager 基础实现。
- path roots/bypass MCP handlers。
- mockclient 协议测试入口。

还没有完成：

- `fs.stat`
- `fs.list`
- `fs.read`
- `fs.search_text`
- `fs.patch`
- 持久化 roots/bypass
- symlink 安全处理
- 文件块索引和 block hash

## Mock Client

运行：

```bash
go run ./cmd/mockclient
```

默认情况下 mockclient 会启动当前源码：

```bash
go run .
```

如果要测试某个已编译好的 server，可以设置：

```bash
FILESYSTEM_MCP_SERVER="F:\ODDS&ENDS\filesystem\bin\fs.exe" go run ./cmd/mockclient
```

如果 mockclient 的 `tools/list` 里还显示旧工具名，例如 `roots/add`、`bypass/add`，说明它启动的是旧二进制，不是当前源码。

## 近期路线

1. 完成 `path/list`、roots、bypass 的 mockclient 验证。
2. 实现 `fs.stat`。
3. 实现 `fs.list`。
4. 实现 `fs.read`，并开始做行号缓存。
5. 接入 `rg` 做 `fs.search_text`。
6. 再考虑 block index、block hash 和可恢复编辑。

## 开发原则

- 先验证模型，再堆功能。
- 先只读，再写入。
- 不在 filesystem-mcp 里执行 shell。
- 不解释 shell 特有路径表达式。
- 工具 schema 的 `jsonschema` 描述要写给 AI 看。
- 文件系统支持的合法文件名，不应该被 MCP 路径层弄坏。
- 写入能力必须建立在 hash、diff 和冲突检测之上。

## 测试

运行全部测试：

```bash
go test ./...
```

当前重点：

```bash
go test ./internal/innerpath
go test ./internal/root
```

## License

MIT
