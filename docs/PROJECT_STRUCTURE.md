# 项目结构说明

## 目录结构

```
everything-mcp/
├── cmd/                           # 可执行程序源码
│   ├── everything-mcp/           # 主程序
│   │   └── main.go               # MCP 服务器实现
│   └── test-client/              # 测试客户端
│       └── main.go               # 测试客户端实现
│
├── docs/                          # 项目文档
│   ├── QUICK_START.md            # 快速开始指南
│   ├── USAGE.md                  # 详细使用说明
│   └── PROJECT_STRUCTURE.md      # 本文档
│
├── examples/                      # 示例配置文件
│   └── mcp-config-example.json   # MCP 客户端配置示例
│
├── scripts/                       # 脚本工具
│   ├── start.sh                  # 服务器启动脚本
│   └── test-mcp.sh               # MCP 测试脚本
│
├── go.mod                         # Go 模块定义
├── go.sum                         # Go 依赖校验和
├── Makefile                       # 构建和任务自动化
├── README.md                      # 项目主文档
└── .gitignore                     # Git 忽略规则
```

## 构建产物

运行 `make build` 或 `make build-all` 后会生成：

```
everything-mcp/
├── everything-mcp                 # 主程序可执行文件
└── test-client                    # 测试客户端可执行文件
```

## 文件说明

### 源码文件

#### `cmd/everything-mcp/main.go`
主程序源码，包含：
- Everything HTTP API 客户端实现
- MCP 服务器实现
- 三个搜索工具的处理逻辑
- HTTP Basic 认证支持
- JSON-RPC 2.0 通信处理

#### `cmd/test-client/main.go`
测试客户端源码，用于：
- 测试 MCP 服务器功能
- 验证工具调用
- 端到端集成测试
- 调试和诊断

### 文档文件

#### `README.md`
项目主文档，包含：
- 项目概述和功能特性
- 安装和配置说明
- 使用方法和示例
- API 文档
- 故障排除指南

#### `docs/QUICK_START.md`
快速开始指南，提供：
- 5 分钟快速上手步骤
- 基本配置示例
- 常见问题快速解决

#### `docs/USAGE.md`
详细使用说明，包含：
- 完整的配置选项
- 高级使用场景
- 搜索语法说明
- 最佳实践

#### `docs/PROJECT_STRUCTURE.md`
项目结构说明（本文档）

### 配置文件

#### `examples/mcp-config-example.json`
MCP 客户端配置示例，用于：
- Cursor IDE 配置
- Claude Desktop 配置
- 其他 MCP 客户端配置

### 脚本文件

#### `scripts/start.sh`
服务器启动脚本，功能：
- 设置环境变量
- 自动编译程序
- 显示配置信息
- 启动服务器

#### `scripts/test-mcp.sh`
MCP 测试脚本，用于：
- 诊断连接问题
- 测试工具调用
- 验证配置

### 构建文件

#### `Makefile`
构建和任务自动化，提供命令：
- `make build` - 编译主程序
- `make build-all` - 编译所有程序
- `make test` - 运行测试
- `make test-coverage` - 生成覆盖率报告
- `make run` - 运行主程序
- `make run-test` - 运行测试客户端
- `make clean` - 清理构建产物
- `make help` - 显示帮助信息

#### `go.mod` 和 `go.sum`
Go 模块管理文件：
- 定义项目依赖
- 锁定依赖版本
- 确保可重现构建

#### `.gitignore`
Git 忽略规则，排除：
- 构建产物（可执行文件）
- 测试输出
- IDE 配置
- 临时文件

## 设计原则

### 1. 标准 Go 项目布局

遵循 Go 社区的标准项目布局：
- `cmd/` - 可执行程序入口
- `docs/` - 项目文档
- `examples/` - 示例配置
- `scripts/` - 辅助脚本

### 2. 单一职责

每个目录和文件都有明确的职责：
- 源码与文档分离
- 可执行程序与库代码分离
- 配置与代码分离

### 3. 易于构建和测试

- 提供 Makefile 简化常用操作
- 测试客户端独立可执行
- 脚本自动化常见任务

### 4. 文档完善

- README 提供全面概览
- QUICK_START 快速上手
- USAGE 详细说明
- PROJECT_STRUCTURE 结构说明

## 开发工作流

### 1. 克隆项目

```bash
git clone https://github.com/yourusername/everything-mcp.git
cd everything-mcp
```

### 2. 构建

```bash
make build-all
```

### 3. 测试

```bash
# 运行单元测试
make test

# 运行集成测试
make run-test
```

### 4. 开发

修改 `cmd/everything-mcp/main.go` 后：

```bash
# 重新编译
make build

# 运行测试
make test

# 测试实际功能
make run-test
```

### 5. 清理

```bash
make clean
```

## 添加新功能

### 添加新工具

1. 在 `cmd/everything-mcp/main.go` 中添加工具定义
2. 实现工具处理函数
3. 在 `handleListTools` 中注册工具
4. 在 `handleCallTool` 中添加路由
5. 更新文档

### 添加新文档

1. 在 `docs/` 目录创建新文档
2. 在 `README.md` 中添加链接
3. 更新 `PROJECT_STRUCTURE.md`

### 添加新脚本

1. 在 `scripts/` 目录创建脚本
2. 添加执行权限：`chmod +x scripts/your-script.sh`
3. 在 `Makefile` 中添加相应目标（可选）
4. 更新文档

## 发布流程

### 1. 更新版本

在 `README.md` 的更新日志中添加新版本信息

### 2. 构建

```bash
make clean
make build-all
make test
```

### 3. 标签

```bash
git tag -a v1.0.1 -m "Release v1.0.1"
git push origin v1.0.1
```

### 4. 发布

创建 GitHub Release，附上编译好的二进制文件

## 维护指南

### 定期任务

- 更新依赖：`go get -u ./...`
- 运行测试：`make test`
- 检查覆盖率：`make test-coverage`
- 更新文档

### 代码质量

- 保持测试覆盖率 > 70%
- 遵循 Go 代码规范
- 添加必要的注释
- 保持文档同步

## 常见问题

### Q: 为什么使用 cmd/ 目录？

A: 这是 Go 项目的标准布局，用于存放可执行程序的入口点。每个子目录代表一个可执行程序。

### Q: 为什么测试客户端是独立程序？

A: 测试客户端需要独立运行来测试服务器，将其作为独立程序可以更灵活地使用和分发。

### Q: 可以添加更多子包吗？

A: 可以。如果代码变得复杂，可以创建 `internal/` 或 `pkg/` 目录来组织共享代码。

### Q: Makefile 是必需的吗？

A: 不是必需的，但强烈推荐。它简化了常用操作，特别是对不熟悉 Go 的用户。
