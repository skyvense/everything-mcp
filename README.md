# Everything MCP Server

一个用 Go 语言实现的 MCP (Model Context Protocol) 服务器，用于调用 Everything 的 HTTP API，为 LLM agent 提供自然语言的文件搜索功能。

> 🚀 **快速开始**: 如果您想立即开始使用，请查看 [快速开始指南](docs/QUICK_START.md)

## 功能特性

- 🔍 **11 种搜索工具**: 提供丰富的文件搜索功能
  - 基本搜索、扩展名搜索、路径搜索
  - 按大小、日期、内容类型搜索
  - 最近文件、大文件、空文件搜索
  - 正则表达式、重复文件名搜索
- 🚀 **高性能**: 利用 Everything 的快速索引能力
- 💬 **自然语言**: LLM agent 可以通过自然语言描述来查找文件
- 🔐 **认证支持**: 支持 HTTP Basic 认证
- 📊 **JSON 格式**: 返回结构化的 JSON 数据
- 🎯 **精确匹配**: 支持 Everything 的完整搜索语法

## 前置要求

1. **Everything 软件**: 需要安装并运行 [Everything](https://www.voidtools.com/)
2. **启用 HTTP 服务器**: 在 Everything 中启用 HTTP 服务器功能
   - 打开 Everything → 工具 → 选项
   - 选择 "HTTP 服务器" 页面
   - 启用 HTTP 服务器
   - 设置端口（默认 80）
   - （可选）启用认证

## 安装

### 从源码编译

```bash
git clone https://github.com/skyvense/everything-mcp.git
cd everything-mcp

# 使用 Makefile 编译（推荐）
make build

# 或者直接使用 go build
go build -o everything-mcp ./cmd/everything-mcp
```

### 使用 Go 安装

```bash
go install github.com/skyvense/everything-mcp/cmd/everything-mcp@latest
```

## 配置

### 环境变量

- `EVERYTHING_BASE_URL`: Everything HTTP API 的基础 URL（默认: `http://localhost`）
- `EVERYTHING_PORT`: Everything HTTP API 的端口（默认: `80`）
- `EVERYTHING_USERNAME`: Everything HTTP API 的用户名（可选，如果 Everything 启用了认证）
- `EVERYTHING_PASSWORD`: Everything HTTP API 的密码（可选，如果 Everything 启用了认证）
- `EVERYTHING_DEBUG`: 启用调试日志（设置为 `true` 可查看详细的请求信息）

### 示例配置

```bash
# 基本配置（无认证）
export EVERYTHING_BASE_URL="http://localhost"
export EVERYTHING_PORT="80"

# 带认证的配置
export EVERYTHING_BASE_URL="http://192.168.7.187"
export EVERYTHING_PORT="51780"
export EVERYTHING_USERNAME="your_username"
export EVERYTHING_PASSWORD="your_password"

# 启用调试模式
export EVERYTHING_DEBUG="true"
```

## 使用方法

### 使用启动脚本（推荐）

项目提供了便捷的启动脚本，已预配置连接信息：

```bash
# Linux/macOS
./scripts/start.sh
```

启动脚本会自动：
- 设置环境变量（URL、端口、用户名、密码）
- 检查并编译程序（如果需要）
- 显示配置信息
- 启动服务器

### 使用 Makefile

```bash
# 编译主程序
make build

# 编译所有程序（包括测试客户端）
make build-all

# 运行主程序
make run

# 运行测试
make test

# 查看所有可用命令
make help
```

### 直接运行

如果需要自定义配置，可以直接设置环境变量后运行：

```bash
export EVERYTHING_BASE_URL="http://192.168.7.187"
export EVERYTHING_PORT="51780"
export EVERYTHING_USERNAME="nate"
export EVERYTHING_PASSWORD="your_password"

./everything-mcp
```

服务器将通过 stdio 与 MCP 客户端通信。

### 在 MCP 客户端中配置

#### Cursor IDE

在 Cursor 的 MCP 配置文件中添加（通常位于 `~/.cursor/mcp.json` 或通过设置界面配置）：

```json
{
  "mcpServers": {
    "everything": {
      "command": "/path/to/everything-mcp",
      "args": [],
      "env": {
        "EVERYTHING_BASE_URL": "http://localhost",
        "EVERYTHING_PORT": "80",
        "EVERYTHING_USERNAME": "your_username",
        "EVERYTHING_PASSWORD": "your_password"
      }
    }
  }
}
```

#### Claude Desktop

在 Claude Desktop 的配置文件中添加：
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`

使用相同的 JSON 格式。

**注意**: 
- 如果 Everything HTTP 服务器没有启用认证，可以省略 `EVERYTHING_USERNAME` 和 `EVERYTHING_PASSWORD`
- 确保 `command` 路径指向实际的可执行文件位置
- 如需调试，可添加 `"EVERYTHING_DEBUG": "true"` 到 `env` 中

## 可用工具

Everything MCP Server 提供 **11 个强大的搜索工具**：

### 基础搜索工具

1. **search_files** - 基本文件搜索
2. **search_by_extension** - 按扩展名搜索
3. **search_by_path** - 按路径搜索

### 高级搜索工具

4. **search_by_size** - 按文件大小搜索
5. **search_by_date** - 按日期搜索
6. **search_recent_files** - 搜索最近修改的文件
7. **search_large_files** - 搜索大文件
8. **search_empty_files** - 搜索空文件/文件夹

### 专业搜索工具

9. **search_by_content_type** - 按内容类型搜索（图片、视频、音频、文档等）
10. **search_with_regex** - 正则表达式搜索
11. **search_duplicate_names** - 搜索重复文件名

### 快速示例

```json
// 搜索最近 7 天的 PDF 文件
{
  "name": "search_recent_files",
  "arguments": {
    "days": 7,
    "query": "ext:pdf"
  }
}

// 搜索大于 100MB 的视频文件
{
  "name": "search_by_content_type",
  "arguments": {
    "content_type": "video",
    "query": "size:>100MB"
  }
}

// 搜索重复的配置文件
{
  "name": "search_duplicate_names",
  "arguments": {
    "filename": "config.json"
  }
}
```

**详细文档**: 查看 [TOOLS.md](docs/TOOLS.md) 了解所有工具的完整说明和使用示例。

## 使用示例

### 通过 LLM Agent 使用

LLM agent 可以通过自然语言来调用这些工具：

**基础搜索**:
- "帮我找所有 PDF 文件"
- "在 Documents 文件夹中搜索包含 'report' 的文件"
- "查找所有 .txt 文件"

**高级搜索**:
- "找出最近 3 天修改的文件"
- "搜索大于 100MB 的文件"
- "查找所有空文件夹"
- "找出 2024 年创建的所有文档"

**专业搜索**:
- "搜索所有图片文件"
- "找出名为 config.json 的所有文件"
- "使用正则表达式搜索所有 .log 文件"
- "查找占用空间最大的 20 个文件"

## 技术细节

### Everything HTTP API

Everything 的 HTTP API 使用简单的 GET 请求：

```
GET http://localhost:80/?search=<查询字符串>&json=1&count=<最大结果数>
```

**重要参数：**
- `search`: 搜索查询字符串
- `json=1`: 请求 JSON 格式响应（推荐）
- `count`: 限制返回结果数量
- `path`: 指定搜索路径

**JSON 响应格式：**
```json
{
  "totalResults": 123,
  "results": [
    {
      "type": "file",
      "name": "example.txt",
      "path": "C:\\Users\\Documents",
      "size": 1024
    }
  ]
}
```

### MCP 协议

本服务器实现了 MCP (Model Context Protocol) 标准：
- **通信方式**: 通过 stdio 与客户端通信
- **协议**: JSON-RPC 2.0
- **协议版本**: 2024-11-05
- **支持的功能**: Tools（工具调用）

## 开发

### 项目结构

```
everything-mcp/
├── cmd/                           # 可执行程序
│   ├── everything-mcp/           # 主程序
│   │   └── main.go
│   └── test-client/              # 测试客户端
│       └── main.go
├── docs/                          # 文档
│   ├── QUICK_START.md            # 快速开始指南
│   ├── USAGE.md                  # 详细使用说明
│   ├── TOOLS.md                  # 工具列表和使用说明
│   └── PROJECT_STRUCTURE.md      # 项目结构说明
├── examples/                      # 示例配置
│   └── mcp-config-example.json   # MCP 配置示例
├── scripts/                       # 脚本
│   ├── start.sh                  # 启动脚本
│   └── test-mcp.sh               # 测试脚本
├── go.mod                         # Go 模块定义
├── go.sum                         # Go 依赖校验
├── Makefile                       # 构建脚本
├── README.md                      # 本文档
└── .gitignore                     # Git 忽略文件
```

### 依赖

- `github.com/mark3labs/mcp-go`: MCP 协议 Go 实现

### 构建

```bash
# 使用 Makefile（推荐）
make build

# 或者使用 go build
go build -o everything-mcp ./cmd/everything-mcp

# 构建测试客户端
make build-test-client
# 或者
go build -o test-client ./cmd/test-client
```

### 测试

#### 单元测试

运行所有单元测试：

```bash
# 使用 Makefile
make test

# 或者使用 go test
go test -v ./...
```

查看测试覆盖率：

```bash
# 生成 HTML 覆盖率报告
make test-coverage

# 或者使用 go test
go test -cover ./...
```

运行特定测试：

```bash
go test -v -run TestEverythingClient_Search ./cmd/everything-mcp
```

当前测试覆盖率达到 **79%**，包括：
- EverythingClient 的搜索功能测试
- MCP 服务器的工具列表和处理测试
- 所有三个搜索工具的完整测试
- 错误处理和边界情况测试
- HTTP 认证测试

#### 集成测试

使用测试客户端进行端到端测试：

```bash
# 使用 Makefile（推荐）
make run-test

# 或者手动编译和运行
make build-test-client
./test-client examples/mcp-config-example.json
```

测试客户端会自动：
1. 启动 MCP 服务器
2. 执行完整的 MCP 协议握手
3. 测试所有可用工具
4. 显示详细的测试结果

详见 [docs/USAGE.md](docs/USAGE.md) 了解更多信息。

### 运行服务器

确保 Everything 的 HTTP 服务器正在运行，然后：

```bash
# 使用 Makefile
make run

# 或者直接运行
./everything-mcp
```

## 故障排除

### HTTP 401 认证错误

如果遇到 "HTTP 错误 401: 认证失败" 错误：

1. **检查用户名和密码**
   ```bash
   # 使用 curl 测试认证
   curl -u username:password "http://host:port/?search=test&json=1"
   ```

2. **确认 Everything HTTP 服务器配置**
   - 打开 Everything → 工具 → 选项 → HTTP 服务器
   - 检查"需要用户名和密码"选项是否启用
   - 确认用户名和密码设置

3. **检查端口配置**
   - 确保 `EVERYTHING_PORT` 与 Everything 中配置的端口一致
   - URL 应该是 `http://host:port` 格式（端口号必须正确）

4. **启用调试模式**
   ```bash
   export EVERYTHING_DEBUG="true"
   ./everything-mcp
   ```
   查看详细的请求信息，包括 URL、认证头等

### 连接错误

如果遇到连接错误，请检查：

1. Everything 是否正在运行
2. HTTP 服务器是否已启用
3. 端口配置是否正确（包括在 URL 中）
4. 防火墙是否阻止了连接
5. 如果是远程服务器，检查网络连接和服务器可访问性

**验证连接：**
```bash
# 测试基本连接（无认证）
curl "http://localhost:80/?search=test&json=1"

# 测试带认证的连接
curl -u username:password "http://host:port/?search=test&json=1"
```

### 搜索无结果

- 确保 Everything 已经索引了您的文件系统
- 检查搜索查询是否正确
- 尝试在 Everything 界面中直接搜索以验证
- 使用 `json=1` 参数确保返回 JSON 格式

### MCP 客户端连接问题

如果 MCP 客户端无法连接到服务器：

1. **检查可执行文件路径**
   - 确保配置文件中的 `command` 路径正确
   - 使用绝对路径而不是相对路径

2. **检查环境变量**
   - 确认所有必需的环境变量都已设置
   - 特别是 `EVERYTHING_BASE_URL` 和 `EVERYTHING_PORT`

3. **查看日志**
   - 在环境变量中添加 `EVERYTHING_DEBUG=true`
   - 检查客户端的日志输出

4. **测试服务器**
   - 使用测试客户端验证服务器功能：
     ```bash
     make run-test
     ```

### 常见问题

**Q: 为什么搜索返回 HTML 而不是文件列表？**

A: 需要在请求中添加 `json=1` 参数。本服务器已自动处理，如果仍有问题，请检查 Everything 版本是否支持 JSON 输出。

**Q: 如何限制搜索结果数量？**

A: 使用 `max_results` 参数，服务器会自动转换为 Everything API 的 `count` 参数。

**Q: 支持哪些搜索语法？**

A: 支持 Everything 的完整搜索语法，包括：
- 通配符：`*.txt`
- 路径搜索：`C:\Users\Documents\`
- 扩展名：`ext:pdf`
- 正则表达式：`regex:.*\.log$`
- 更多语法见 [Everything 搜索语法](https://www.voidtools.com/support/everything/searching/)

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 更新日志

### v1.1.0 (2026-01-12)

- ✨ **新增 8 个搜索工具**，总计 11 个工具
  - `search_by_size` - 按文件大小搜索
  - `search_by_date` - 按日期搜索
  - `search_recent_files` - 搜索最近修改的文件
  - `search_large_files` - 搜索大文件
  - `search_empty_files` - 搜索空文件/文件夹
  - `search_by_content_type` - 按内容类型搜索
  - `search_with_regex` - 正则表达式搜索
  - `search_duplicate_names` - 搜索重复文件名
- ✨ 添加文件大小格式化显示
- 📝 新增完整的工具文档 (TOOLS.md)

### v1.0.1 (2026-01-12)

- 🐛 修复 URL 端口号未正确添加的问题
- 🐛 修复 Everything HTTP API 返回 HTML 而不是 JSON 的问题
- ✨ 添加 JSON 格式支持（`json=1` 参数）
- ✨ 添加调试模式（`EVERYTHING_DEBUG` 环境变量）
- ✨ 改进错误消息，特别是 401 认证错误
- 📝 添加测试客户端 (`test-client`)
- 📝 完善文档和故障排除指南
- 🏗️ 重构项目结构，采用标准 Go 项目布局
- 🔧 添加 Makefile 简化构建和测试

### v1.0.0 (2026-01-11)

- 🎉 初始版本发布
- ✨ 实现三个搜索工具：`search_files`、`search_by_extension`、`search_by_path`
- ✨ 支持 HTTP Basic 认证
- ✨ 支持 MCP 协议 2024-11-05
- 📝 完整的单元测试覆盖

## 相关链接

- [Everything 官网](https://www.voidtools.com/)
- [Everything HTTP API 文档](https://www.voidtools.com/support/everything/http/)
- [Everything 搜索语法](https://www.voidtools.com/support/everything/searching/)
- [MCP 协议规范](https://modelcontextprotocol.io/)
- [mcp-go 库](https://github.com/mark3labs/mcp-go)

## 致谢

- [Everything](https://www.voidtools.com/) - 快速文件搜索工具
- [mcp-go](https://github.com/mark3labs/mcp-go) - Go 语言的 MCP 协议实现
- [Anthropic](https://www.anthropic.com/) - MCP 协议的创建者
