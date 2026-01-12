#!/bin/bash

# MCP Server 测试脚本
# 用于诊断 MCP 连接问题

echo "=== Everything MCP Server 诊断测试 ==="
echo ""

# 1. 检查程序是否存在
echo "1. 检查程序文件..."
if [ ! -f "./everything-mcp" ]; then
    echo "   ❌ 程序不存在，正在编译..."
    go build -o everything-mcp main.go
    if [ $? -ne 0 ]; then
        echo "   ❌ 编译失败"
        exit 1
    fi
    echo "   ✅ 编译成功"
else
    echo "   ✅ 程序存在"
fi

# 2. 检查执行权限
echo ""
echo "2. 检查执行权限..."
if [ -x "./everything-mcp" ]; then
    echo "   ✅ 有执行权限"
else
    echo "   ⚠️  没有执行权限，正在添加..."
    chmod +x ./everything-mcp
    echo "   ✅ 已添加执行权限"
fi

# 3. 测试初始化请求
echo ""
echo "3. 测试 MCP 初始化请求..."
echo "   发送 initialize 请求..."

init_request='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"1.0","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}'

response=$(echo "$init_request" | timeout 3 ./everything-mcp 2>&1)

if echo "$response" | grep -q "result"; then
    echo "   ✅ 初始化成功"
    echo "$response" | head -3
else
    echo "   ⚠️  初始化响应异常"
    echo "   响应: $response"
fi

# 4. 测试工具列表
echo ""
echo "4. 测试工具列表请求..."
echo "   发送 tools/list 请求..."

tools_request='{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'

# 先初始化，再请求工具列表
(echo "$init_request"; sleep 0.1; echo "$tools_request") | timeout 3 ./everything-mcp 2>&1 | grep -E "(result|error|tools)" | head -5

echo ""
echo "=== 测试完成 ==="
echo ""
echo "如果测试失败，请检查："
echo "1. Everything HTTP API 是否可访问"
echo "2. 环境变量是否正确设置"
echo "3. 查看完整日志了解详细错误"
