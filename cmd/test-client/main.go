package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

// MCPConfig MCP 配置文件结构
type MCPConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// MCPServerConfig 单个 MCP 服务器配置
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

// JSONRPCRequest JSON-RPC 请求
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse JSON-RPC 响应
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError JSON-RPC 错误
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// toFloat64 将 interface{} 转换为 float64（用于比较 JSON 数字）
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}

// MCPClient MCP 客户端
type MCPClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	scanner *bufio.Scanner
	requestID int
}

// NewMCPClient 创建新的 MCP 客户端
func NewMCPClient(config MCPServerConfig) (*MCPClient, error) {
	// 构建命令
	cmd := exec.Command(config.Command, config.Args...)
	
	// 设置环境变量
	env := os.Environ()
	for k, v := range config.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env
	
	// 设置 stdio
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("创建 stdin pipe 失败: %w", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("创建 stdout pipe 失败: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("创建 stderr pipe 失败: %w", err)
	}
	
	// 启动进程
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动进程失败: %w", err)
	}
	
	// 启动 stderr 读取 goroutine
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("[SERVER STDERR] %s", scanner.Text())
		}
	}()
	
	return &MCPClient{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		scanner: bufio.NewScanner(stdout),
		requestID: 1,
	}, nil
}

// SendRequest 发送 JSON-RPC 请求并等待响应
func (c *MCPClient) SendRequest(ctx context.Context, method string, params interface{}) (*JSONRPCResponse, error) {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      c.requestID,
		Method:  method,
		Params:  params,
	}
	c.requestID++
	
	// 序列化请求
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}
	
	// 发送请求
	log.Printf("[CLIENT] 发送请求: %s", string(requestBytes))
	if _, err := fmt.Fprintf(c.stdin, "%s\n", requestBytes); err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	
	// 等待响应
	response, err := c.ReadResponse(ctx)
	if err != nil {
		return nil, err
	}
	
	// 验证响应 ID（JSON 数字可能被解析为 float64，需要比较数值）
	requestIDFloat, reqIsFloat := toFloat64(request.ID)
	responseIDFloat, respIsFloat := toFloat64(response.ID)
	
	if reqIsFloat && respIsFloat {
		if requestIDFloat != responseIDFloat {
			return nil, fmt.Errorf("响应 ID 不匹配: 期望 %v, 得到 %v", request.ID, response.ID)
		}
	} else if response.ID != request.ID {
		return nil, fmt.Errorf("响应 ID 不匹配: 期望 %v, 得到 %v", request.ID, response.ID)
	}
	
	return response, nil
}

// SendNotification 发送 JSON-RPC 通知（不需要响应）
func (c *MCPClient) SendNotification(method string, params interface{}) error {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("序列化通知失败: %w", err)
	}
	
	log.Printf("[CLIENT] 发送通知: %s", string(requestBytes))
	if _, err := fmt.Fprintf(c.stdin, "%s\n", requestBytes); err != nil {
		return fmt.Errorf("发送通知失败: %w", err)
	}
	
	return nil
}

// ReadResponse 读取 JSON-RPC 响应
func (c *MCPClient) ReadResponse(ctx context.Context) (*JSONRPCResponse, error) {
	// 设置超时
	done := make(chan error, 1)
	var response *JSONRPCResponse
	var err error
	
	go func() {
		if !c.scanner.Scan() {
			err = fmt.Errorf("读取响应失败: %v", c.scanner.Err())
			done <- err
			return
		}
		
		line := c.scanner.Text()
		log.Printf("[CLIENT] 收到响应: %s", line)
		
		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			done <- fmt.Errorf("解析响应失败: %w", err)
			return
		}
		
		response = &resp
		done <- nil
	}()
	
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-done:
		if err != nil {
			return nil, err
		}
		return response, nil
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("读取响应超时")
	}
}

// Close 关闭客户端
func (c *MCPClient) Close() error {
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.stdout != nil {
		c.stdout.Close()
	}
	if c.stderr != nil {
		c.stderr.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

func main() {
	// 读取配置文件
	configFile := "mcp-config-example.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	
	configData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}
	
	var mcpConfig MCPConfig
	if err := json.Unmarshal(configData, &mcpConfig); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}
	
	// 获取 everything 服务器配置
	serverConfig, ok := mcpConfig.MCPServers["everything"]
	if !ok {
		log.Fatalf("配置文件中未找到 'everything' 服务器配置")
	}
	
	fmt.Println("=== MCP Client 测试 ===")
	fmt.Printf("配置文件: %s\n", configFile)
	fmt.Printf("服务器命令: %s\n", serverConfig.Command)
	fmt.Printf("参数: %v\n", serverConfig.Args)
	fmt.Printf("环境变量: %v\n", serverConfig.Env)
	fmt.Println()
	
	// 创建客户端
	client, err := NewMCPClient(serverConfig)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()
	
	ctx := context.Background()
	
	// 1. 发送 initialize 请求
	fmt.Println("1. 发送 initialize 请求...")
	initParams := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}
	
	initResponse, err := client.SendRequest(ctx, "initialize", initParams)
	if err != nil {
		log.Fatalf("initialize 请求失败: %v", err)
	}
	
	if initResponse.Error != nil {
		log.Fatalf("initialize 错误: %v", initResponse.Error)
	}
	
	fmt.Printf("✅ initialize 成功\n")
	var initResult map[string]interface{}
	if err := json.Unmarshal(initResponse.Result, &initResult); err == nil {
		if serverInfo, ok := initResult["serverInfo"].(map[string]interface{}); ok {
			fmt.Printf("   服务器: %v %v\n", serverInfo["name"], serverInfo["version"])
		}
		if protocolVersion, ok := initResult["protocolVersion"].(string); ok {
			fmt.Printf("   协议版本: %s\n", protocolVersion)
		}
	}
	fmt.Println()
	
	// 2. 发送 initialized 通知
	fmt.Println("2. 发送 initialized 通知...")
	if err := client.SendNotification("notifications/initialized", map[string]interface{}{}); err != nil {
		log.Fatalf("initialized 通知失败: %v", err)
	}
	fmt.Println("✅ initialized 通知已发送")
	fmt.Println()
	
	// 等待一下，让服务器处理通知
	time.Sleep(100 * time.Millisecond)
	
	// 3. 列出工具
	fmt.Println("3. 列出可用工具...")
	toolsResponse, err := client.SendRequest(ctx, "tools/list", map[string]interface{}{})
	if err != nil {
		log.Fatalf("tools/list 请求失败: %v", err)
	}
	
	if toolsResponse.Error != nil {
		log.Fatalf("tools/list 错误: %v", toolsResponse.Error)
	}
	
	fmt.Println("✅ tools/list 成功")
	var toolsResult map[string]interface{}
	if err := json.Unmarshal(toolsResponse.Result, &toolsResult); err == nil {
		if tools, ok := toolsResult["tools"].([]interface{}); ok {
			fmt.Printf("   找到 %d 个工具:\n", len(tools))
			for i, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					if name, ok := toolMap["name"].(string); ok {
						if desc, ok := toolMap["description"].(string); ok {
							fmt.Printf("   %d. %s: %s\n", i+1, name, desc)
						} else {
							fmt.Printf("   %d. %s\n", i+1, name)
						}
					}
				}
			}
		}
	}
	fmt.Println()
	
	// 4. 测试搜索文件
	fmt.Println("4. 测试搜索文件 (search_files: oray)...")
	searchParams := map[string]interface{}{
		"name": "search_files",
		"arguments": map[string]interface{}{
			"query":       "oray",
			"max_results": 10,
		},
	}
	
	searchResponse, err := client.SendRequest(ctx, "tools/call", searchParams)
	if err != nil {
		log.Fatalf("tools/call 请求失败: %v", err)
	}
	
	if searchResponse.Error != nil {
		log.Fatalf("tools/call 错误: %v", searchResponse.Error)
	}
	
	fmt.Println("✅ search_files 调用成功")
	var callResult map[string]interface{}
	if err := json.Unmarshal(searchResponse.Result, &callResult); err == nil {
		if content, ok := callResult["content"].([]interface{}); ok {
			for _, item := range content {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if text, ok := itemMap["text"].(string); ok {
						fmt.Printf("   结果:\n%s\n", text)
					}
				}
			}
		}
		if isError, ok := callResult["isError"].(bool); ok && isError {
			fmt.Println("   ⚠️  工具返回错误")
		}
	}
	fmt.Println()
	
	// 5. 测试按扩展名搜索
	fmt.Println("5. 测试按扩展名搜索 (search_by_extension: txt)...")
	extParams := map[string]interface{}{
		"name": "search_by_extension",
		"arguments": map[string]interface{}{
			"extension":   "txt",
			"max_results": 5,
		},
	}
	
	extResponse, err := client.SendRequest(ctx, "tools/call", extParams)
	if err != nil {
		log.Fatalf("tools/call 请求失败: %v", err)
	}
	
	if extResponse.Error != nil {
		log.Fatalf("tools/call 错误: %v", extResponse.Error)
	}
	
	fmt.Println("✅ search_by_extension 调用成功")
	var extResult map[string]interface{}
	if err := json.Unmarshal(extResponse.Result, &extResult); err == nil {
		if content, ok := extResult["content"].([]interface{}); ok {
			for _, item := range content {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if text, ok := itemMap["text"].(string); ok {
						// 只显示前几行
						lines := []rune(text)
						if len(lines) > 200 {
							fmt.Printf("   结果 (前200字符):\n%s...\n", string(lines[:200]))
						} else {
							fmt.Printf("   结果:\n%s\n", text)
						}
					}
				}
			}
		}
	}
	fmt.Println()
	
	fmt.Println("=== 测试完成 ===")
}
