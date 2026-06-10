# filesystem-mcp

一个为 AI 编程助手设计的轻量级 filesystem MCP server。

当前项目还处在早期阶段。它已经跑通了 MCP server、工具注册、workspace root 管理和 bypass 规则的最小闭环；现在正在重构路径系统，为后续的 `fs.list`、`fs.read`、`fs.stat` 和安全写入打地基。

> 现阶段重点不是“尽快读写文件”，而是先把路径解析、root 边界和前缀匹配做稳。

## 安装

```bash
go install github.com/yuxi39/filesystem-mcp@latest
```

或者从源码构建：

```bash
git clone https://github.com/yuxi39/filesystem-mcp.git
cd filesystem-mcp
go build
```

## VS Code 配置

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

项目级配置可以放在 `.vscode/mcp.json`。

配置完成后，重启 VS Code 或当前 MCP session，然后检查 MCP server 是否处于 running 状态。

## 当前状态

### 已完成

- MCP server 可以启动并被 VS Code / Codex 调用。
- server metadata 已配置名称、版本和 icon。
- 已实现 roots 管理工具。
- 已实现 bypass 管理工具。
- 已上传 GitHub。
- 已打 tag。
- 已验证 `go install` 安装链路。
- `internal/path` 中开始重构路径系统。
- 新增了目录前缀树 `prefixTree`，用于 root / bypass / ignore 的前缀匹配。
- 已为 `prefixTree` 补充测试，覆盖插入、匹配、父子覆盖、删除和兄弟分支保留。

### 当前正在重构

路径系统正在从旧的 roots 逻辑中拆出来，目标是形成独立的 `PathManager`。

预期职责：

- 管理 roots。
- 管理 bypass。
- 管理 ignore。
- 接收 MCP 输入路径。
- 解析 namespace path，例如 `odds:hello/README.md`。
- 解析绝对路径。
- 解析 file URI。
- 在 Windows 和 Linux/macOS 上统一转换为内部路径段。
- 判断路径是否落在允许的 root 内。
- 判断路径是否命中 bypass。
- 为后续 `fs.read` / `fs.write` 提供安全边界。

### 尚未完成

- Windows 路径解析还在重构中。
- Linux/macOS 路径解析还在重构中。
- `PathManager.Resolve` 尚未完成。
- `fs.stat` 尚未实现。
- `fs.list` 尚未实现。
- `fs.read` 尚未实现。
- `fs.search_names` 尚未实现。
- `fs.search_text` 尚未实现。
- `fs.diff` / `fs.patch` 尚未实现。
- memory 工具尚未实现。

## 已实现工具

### `roots/list`

列出当前注册的 workspace roots 和 bypass 规则。

### `roots/add`

注册新的 workspace root。

输入示例：

```json
{
  "name": "odds",
  "path": "F:\\ODDS&ENDS"
}
```

### `roots/del`

删除一个已注册的 workspace root。

输入示例：

```json
{
  "name": "odds"
}
```

### `bypass/add`

阻止 agent 访问某个 root 下的子路径。

输入示例：

```json
{
  "path": "odds:secret",
  "reason": "Contains sensitive credentials"
}
```

### `bypass/del`

按 index 删除 bypass 规则。index 来自 `roots/list` 返回的 bypasses 数组。

输入示例：

```json
{
  "index": 0
}
```

## 路径模型草案

用户或 MCP client 输入的路径可能有几种形式：

- namespace path: `odds:hello/README.md`
- Windows absolute path: `F:\ODDS&ENDS\hello\README.md`
- Unix absolute path: `/etc/cron.d`
- file URI: `file:///f%3A/ODDS%26ENDS/filesystem`

内部路径系统会把路径转换成规范化后的 segment 列表。

Windows 示例：

```go
[]string{"f:", "ODDS&ENDS", "filesystem"}
```

Linux/macOS 示例：

```go
[]string{"etc", "cron.d"}
```

`prefixTree` 只处理这种已经规范化后的 segment 列表。它不负责：

- 清理 `.` / `..`
- 大小写规范化
- URI decode
- 路径分隔符转换
- symlink 解析

这些应该由上游路径解析层完成。

## `prefixTree` 当前语义

`prefixTree` 用于判断一个路径是否被某个已注册前缀覆盖。

支持：

- 插入 root / bypass / ignore 前缀。
- 查询某个路径是否命中已注册前缀。
- 插入父路径时替换已有子路径。
- 已有父路径时拒绝插入子路径。
- 删除某个前缀，并清理无用分支。

示例：

```go
tree.InsertTree("odds", []string{"f:", "ODDS&ENDS"})
tree.Match([]string{"f:", "ODDS&ENDS", "hello", "README.md"}) // true
```

## 近期路线

### 1. 完成路径解析

先完成 `internal/path`：

- Windows absolute path -> segments
- Unix absolute path -> segments
- file URI -> native path -> segments
- namespace path -> root + relative segments
- root boundary check
- bypass check

### 2. 把 roots / bypass 迁入 PathManager

当前 roots 和 bypass 仍在旧结构中。下一步要让它们统一经过 `PathManager`。

### 3. 实现只读文件工具

优先级：

```txt
fs.stat
fs.list
fs.read
```

这三个完成后，Codex 就可以用这个 MCP 稳定地查看项目，而不是依赖 shell。

### 4. 再实现搜索

```txt
fs.search_names
fs.search_text
```

搜索工具会让 agent 更快找到相关代码。

### 5. 最后进入安全写入

```txt
fs.diff
fs.patch
```

写入必须建立在可靠的路径系统、hash 冲突检测和 diff 预览之上。

## 设计原则

- 安全优先：任何路径都必须先通过 root 边界检查。
- 先只读，后写入。
- 路径模型先稳定，再做文件操作。
- 工具返回结构化 JSON，不返回大段自然语言。
- 大文件读取必须支持限制和截断。
- 写入必须支持 hash 校验和可审计 diff。
- 不在 filesystem MCP 里执行 shell 命令。
- 不在早期版本实现递归删除。

## 当前反思

这次重构看起来让功能进度变慢了，但它是在修正地基。

filesystem MCP 最危险的部分不是 “文件读不出来”，而是：

- 把不该暴露的路径暴露出来。
- 误判 root 边界。
- 被 `..` 或 symlink 绕过。
- bypass 规则没有真正生效。
- 写入时覆盖用户未读到的新改动。

因此，当前阶段把 path system 拆出来是合理的。

## 测试

运行全部测试：

```bash
go test ./...
```

当前重点测试：

```bash
go test ./internal/path
```

`prefixTree` 已有测试覆盖：

- 插入并匹配。
- 父前缀匹配子路径。
- 已有父前缀时拒绝插入子前缀。
- 插入父前缀时替换子前缀。
- 删除唯一节点。
- 删除一个分支时保留兄弟分支。

## License

MIT
