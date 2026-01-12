# Everything MCP Server - 快速开始指南

## 5 分钟快速上手

### 1. 安装 Everything

1. 下载并安装 [Everything](https://www.voidtools.com/)
2. 启动 Everything，等待索引完成

### 2. 配置 Everything HTTP 服务器

1. 打开 Everything → **工具** → **选项**
2. 选择 **HTTP 服务器** 页面
3. 勾选 **启用 HTTP 服务器**
4. 设置端口（例如：51780）
5. （可选）启用认证：
   - 勾选 **需要用户名和密码**
   - 设置用户名和密码
6. 点击 **确定**

### 3. 编译 MCP 服务器

```bash
cd /path/to/everything-mcp
go build -o everything-mcp main.go
```

### 4. 配置环境变量

编辑 `start.sh`（Linux/macOS）或创建你自己的配置：

```bash
export EVERYTHING_BASE_URL="http://192.168.7.187"
export EVERYTHING_PORT="51780"
export EVERYTHING_USERNAME="your_username"
export EVERYTHING_PASSWORD="your_password"
```

### 5. 测试连接

使用 curl 测试 Everything HTTP API：

```bash
curl -u username:password "http://host:port/?search=test&json=1&count=5"
```

如果返回 JSON 格式的结果，说明配置正确。

### 6. 测试 MCP 服务器

使用测试客户端：

```bash
# 编译测试客户端
go build -o test_client test_client.go

# 运行测试
./test_client mcp-config-example.json
```

如果看到 ✅ 标记和搜索结果，说明一切正常！

### 7. 配置 MCP 客户端

#### Cursor IDE

编辑配置文件（通常在 `~/.cursor/mcp.json`）：

```json
{
  "mcpServers": {
    "everything": {
      "command": "/absolute/path/to/everything-mcp",
      "args": [],
      "env": {
        "EVERYTHING_BASE_URL": "http://192.168.7.187",
        "EVERYTHING_PORT": "51780",
        "EVERYTHING_USERNAME": "your_username",
        "EVERYTHING_PASSWORD": "your_password"
      }
    }
  }
}
```

#### Claude Desktop

编辑配置文件：
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`

使用相同的 JSON 格式。

### 8. 开始使用

在 MCP 客户端中尝试：

- "帮我找所有 PDF 文件"
- "搜索包含 'report' 的文件"
- "在 Documents 文件夹中查找 txt 文件"

## 常见问题快速解决

### ❌ HTTP 401 错误

```bash
# 检查认证
curl -v -u username:password "http://host:port/?search=test&json=1"

# 启用调试
export EVERYTHING_DEBUG="true"
./everything-mcp
```

### ❌ 连接超时

1. 确认 Everything 正在运行
2. 检查端口号是否正确
3. 测试网络连接：`ping host`

### ❌ 搜索无结果

1. 在 Everything 界面中直接搜索测试
2. 确认索引已完成
3. 检查搜索语法

### ❌ MCP 客户端无法连接

1. 使用绝对路径指定 `command`
2. 检查环境变量是否正确
3. 使用测试客户端验证服务器

## 下一步

- 📖 阅读完整文档：[README.md](README.md)
- 🔧 故障排除：[TROUBLESHOOTING.md](TROUBLESHOOTING.md)
- 🧪 测试客户端：[TEST_CLIENT.md](TEST_CLIENT.md)
- 🔍 Everything 搜索语法：https://www.voidtools.com/support/everything/searching/

## 需要帮助？

1. 检查日志输出（启用 `EVERYTHING_DEBUG=true`）
2. 使用测试客户端诊断问题
3. 查看故障排除文档
4. 提交 Issue 到 GitHub

祝您使用愉快！🎉
