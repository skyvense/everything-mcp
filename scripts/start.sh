#!/bin/bash

# Everything MCP Server 启动脚本
# 配置信息:
#   - URL: http://192.168.7.187
#   - 端口: 51780
#   - 用户名: nate
#   - 密码: checkout

# 设置环境变量
export EVERYTHING_BASE_URL="http://192.168.7.187"
export EVERYTHING_PORT="51780"
export EVERYTHING_USERNAME="nate"
export EVERYTHING_PASSWORD="checkout888"

# 检查程序是否存在
if [ ! -f "./everything-mcp" ]; then
    echo "错误: 找不到 everything-mcp 程序"
    echo "正在编译..."
    go build -o everything-mcp ./cmd/everything-mcp
    if [ $? -ne 0 ]; then
        echo "编译失败，请检查 Go 环境"
        exit 1
    fi
    echo "编译成功"
fi

# 显示配置信息
echo "=== Everything MCP Server 启动配置 ==="
echo "Everything API: ${EVERYTHING_BASE_URL}:${EVERYTHING_PORT}"
echo "用户名: ${EVERYTHING_USERNAME}"
echo "密码: ${EVERYTHING_PASSWORD:0:1}***"  # 只显示密码的第一个字符
echo ""
echo "启动服务器..."
echo ""

# 启动服务器
./everything-mcp
