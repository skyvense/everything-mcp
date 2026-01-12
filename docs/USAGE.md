# Everything MCP Server 使用指南

## 快速开始

MCP server 通过 **stdio**（标准输入输出）与客户端通信，这意味着它不需要单独启动，而是由 MCP 客户端自动启动和管理。

## 工作原理

```
MCP 客户端 (Claude Desktop/Cursor) 
    ↓ (启动进程)
Everything MCP Server (./everything-mcp)
    ↓ (HTTP 请求)
Everything HTTP API (192.168.7.187:51780)
    ↓ (返回结果)
Everything MCP Server
    ↓ (JSON-RPC 响应)
MCP 客户端
    ↓ (显示给用户)
LLM Agent / 用户
```

## 配置 MCP 客户端

### 1. Claude Desktop (Anthropic)

配置文件位置：
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

配置内容：

```json
{
  "mcpServers": {
    "everything": {
      "command": "/Users/nate/Documents/git/everything-mcp/everything-mcp",
      "env": {
        "EVERYTHING_BASE_URL": "http://192.168.7.187",
        "EVERYTHING_PORT": "51780",
        "EVERYTHING_USERNAME": "nate",
        "EVERYTHING_PASSWORD": "checkout888"
      }
    }
  }
}
```

**重要提示**：
- 使用 `everything-mcp` 的**绝对路径**
- 确保文件有执行权限：`chmod +x everything-mcp`
- 配置后重启 Claude Desktop

### 2. Cursor IDE

配置文件位置：
- `~/.cursor/mcp.json` 或项目根目录的 `.cursor/mcp.json`

配置内容：

```json
{
  "mcpServers": {
    "everything": {
      "command": "/Users/nate/Documents/git/everything-mcp/everything-mcp",
      "env": {
        "EVERYTHING_BASE_URL": "http://192.168.7.187",
        "EVERYTHING_PORT": "51780",
        "EVERYTHING_USERNAME": "nate",
        "EVERYTHING_PASSWORD": "checkout888"
      }
    }
  }
}
```

### 3. 其他 MCP 客户端

任何支持 MCP 协议的客户端都可以使用，配置格式类似：

```json
{
  "mcpServers": {
    "everything": {
      "command": "/绝对路径/to/everything-mcp",
      "env": {
        "EVERYTHING_BASE_URL": "http://192.168.7.187",
        "EVERYTHING_PORT": "51780",
        "EVERYTHING_USERNAME": "nate",
        "EVERYTHING_PASSWORD": "checkout888"
      }
    }
  }
}
```

## 使用方法

### 方法 1: 通过 LLM Agent 使用（推荐）

配置完成后，在支持 MCP 的 LLM 客户端中，你可以直接用自然语言提问：

#### 示例对话

**用户**: "帮我找所有 PDF 文件"

**LLM Agent** (自动调用 `search_by_extension` 工具):
```
找到了以下 PDF 文件：
1. C:\Users\Documents\report.pdf
2. C:\Users\Documents\invoice.pdf
...
```

**用户**: "在 Documents 文件夹中搜索包含 'report' 的文件"

**LLM Agent** (自动调用 `search_by_path` 工具):
```
在 C:\Users\Documents 中找到以下文件：
1. C:\Users\Documents\report_2024.pdf
2. C:\Users\Documents\report_summary.txt
...
```

**用户**: "查找所有 .txt 文件，最多显示 10 个"

**LLM Agent** (自动调用 `search_by_extension` 工具):
```
找到了以下 .txt 文件（显示前 10 个）：
1. C:\Users\Documents\notes.txt
2. C:\Users\Documents\readme.txt
...
```

### 方法 2: 直接测试 MCP 连接

如果你想测试 MCP server 是否正常工作，可以手动发送 JSON-RPC 请求：

```bash
# 测试工具列表
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./everything-mcp

# 测试搜索（需要先初始化）
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1.0","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./everything-mcp
```

## 可用工具说明

### 1. search_files - 通用文件搜索

**功能**: 搜索文件和文件夹

**自然语言示例**:
- "搜索名为 'invoice' 的文件"
- "查找包含 'report' 的所有文件"
- "找一下 'document' 相关的文件"

**技术参数**:
- `query`: 搜索关键词
- `max_results`: 最大结果数（默认 100）

### 2. search_by_extension - 按扩展名搜索

**功能**: 查找特定类型的文件

**自然语言示例**:
- "帮我找所有 PDF 文件"
- "查找所有 .txt 文件"
- "搜索所有图片文件（jpg, png）"

**技术参数**:
- `extension`: 文件扩展名（不需要点号）
- `max_results`: 最大结果数（默认 100）

### 3. search_by_path - 按路径搜索

**功能**: 在指定目录中搜索

**自然语言示例**:
- "在 Documents 文件夹中搜索文件"
- "在 C:\Users\Documents 中查找包含 'report' 的文件"
- "搜索 D:\Projects 目录下的所有文件"

**技术参数**:
- `path`: 搜索路径
- `query`: 可选的关键词
- `max_results`: 最大结果数（默认 100）

## 故障排除

### 问题 1: MCP 客户端无法连接

**检查项**:
1. 确认 `everything-mcp` 文件存在且有执行权限
2. 确认路径是绝对路径
3. 检查配置文件 JSON 格式是否正确
4. 查看客户端日志（通常在客户端设置中）

**解决方案**:
```bash
# 检查文件权限
ls -l everything-mcp

# 如果没有执行权限，添加权限
chmod +x everything-mcp

# 测试程序是否能运行
./everything-mcp --help  # 或者直接运行看是否有错误
```

### 问题 2: 搜索返回空结果

**可能原因**:
1. Everything HTTP API 连接失败
2. 认证信息错误
3. Everything 未索引文件系统

**检查方法**:
```bash
# 测试 Everything API 连接
curl -u nate:checkout888 "http://192.168.7.187:51780/?search=test"

# 如果返回 401，说明认证信息错误
# 如果返回 200，说明连接正常
```

### 问题 3: 工具列表为空

**可能原因**:
1. MCP server 未正确启动
2. 初始化失败

**解决方案**:
1. 检查环境变量是否正确设置
2. 查看客户端错误日志
3. 尝试手动运行 server 查看错误信息

## 高级用法

### 自定义配置

如果需要为不同的环境使用不同的配置，可以创建多个配置：

```json
{
  "mcpServers": {
    "everything-local": {
      "command": "/path/to/everything-mcp",
      "env": {
        "EVERYTHING_BASE_URL": "http://localhost",
        "EVERYTHING_PORT": "80"
      }
    },
    "everything-remote": {
      "command": "/path/to/everything-mcp",
      "env": {
        "EVERYTHING_BASE_URL": "http://192.168.7.187",
        "EVERYTHING_PORT": "51780",
        "EVERYTHING_USERNAME": "nate",
        "EVERYTHING_PASSWORD": "checkout888"
      }
    }
  }
}
```

### 调试模式

如果需要查看详细的通信日志，可以修改代码添加日志输出，或者使用支持调试的 MCP 客户端。

## 下一步

1. ✅ 配置 MCP 客户端
2. ✅ 重启客户端
3. ✅ 在对话中尝试自然语言搜索
4. ✅ 享受快速的文件搜索体验！

## 常见问题

**Q: 为什么不能直接运行 `./start.sh` 来使用？**

A: MCP server 需要通过 stdio 与客户端通信，不能独立运行。`start.sh` 主要用于测试和开发。实际使用时，MCP 客户端会自动启动 server。

**Q: 可以在多个客户端中同时使用吗？**

A: 可以！每个客户端会启动自己的 server 实例，它们互不干扰。

**Q: 搜索速度如何？**

A: 搜索速度取决于 Everything 的索引速度，通常非常快（毫秒级），因为 Everything 使用文件系统索引而不是实时扫描。

**Q: 支持哪些操作系统？**

A: Everything 主要支持 Windows，但如果你在 macOS/Linux 上运行 Everything 的 HTTP API，这个 MCP server 也可以工作。
