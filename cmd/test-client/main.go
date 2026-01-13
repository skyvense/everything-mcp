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
	
	// 测试用例列表
	testCases := []struct {
		name        string
		tool        string
		arguments   map[string]interface{}
		description string
	}{
		{
			name: "test_1_search_files",
			tool: "search_files",
			arguments: map[string]interface{}{
				"query":       "txt",
				"max_results": 5,
			},
			description: "基本文件搜索 (搜索包含 txt 的文件)",
		},
		{
			name: "test_2_search_by_extension",
			tool: "search_by_extension",
			arguments: map[string]interface{}{
				"extension":   "txt",
				"max_results": 5,
			},
			description: "按扩展名搜索 (搜索 .txt 文件)",
		},
		{
			name: "test_3_search_by_path",
			tool: "search_by_path",
			arguments: map[string]interface{}{
				"path":        "C:\\",
				"query":       "txt",
				"max_results": 5,
			},
			description: "按路径搜索 (在 C:\\ 中搜索 txt)",
		},
		{
			name: "test_4_search_by_size",
			tool: "search_by_size",
			arguments: map[string]interface{}{
				"size_min":    "1KB",
				"size_max":    "1MB",
				"max_results": 5,
			},
			description: "按大小搜索 (搜索 1KB-1MB 的文件)",
		},
		{
			name: "test_5_search_by_date",
			tool: "search_by_date",
			arguments: map[string]interface{}{
				"date_from":   "2024-01-01",
				"date_to":     "2024-12-31",
				"date_type":   "modified",
				"max_results": 5,
			},
			description: "按日期搜索 (搜索 2024 年修改的文件)",
		},
		{
			name: "test_6_search_recent_files",
			tool: "search_recent_files",
			arguments: map[string]interface{}{
				"days":        7,
				"max_results": 5,
			},
			description: "搜索最近文件 (最近 7 天)",
		},
		{
			name: "test_7_search_large_files",
			tool: "search_large_files",
			arguments: map[string]interface{}{
				"min_size":    "10MB",
				"max_results": 5,
			},
			description: "搜索大文件 (>10MB)",
		},
		{
			name: "test_8_search_empty_files",
			tool: "search_empty_files",
			arguments: map[string]interface{}{
				"type":        "file",
				"max_results": 5,
			},
			description: "搜索空文件",
		},
		{
			name: "test_9_search_by_content_type",
			tool: "search_by_content_type",
			arguments: map[string]interface{}{
				"content_type": "image",
				"max_results":  5,
			},
			description: "按内容类型搜索 (搜索图片)",
		},
		{
			name: "test_10_search_with_regex",
			tool: "search_with_regex",
			arguments: map[string]interface{}{
				"regex":       ".*\\.txt$",
				"max_results": 5,
			},
			description: "正则表达式搜索 (搜索 .txt 结尾的文件)",
		},
		{
			name: "test_11_search_duplicate_names",
			tool: "search_duplicate_names",
			arguments: map[string]interface{}{
				"filename":    "config.txt",
				"max_results": 5,
			},
			description: "搜索重复文件名 (搜索 config.txt)",
		},
		{
			name: "test_12_list_drives",
			tool: "list_drives",
			arguments: map[string]interface{}{},
			description: "列出所有驱动器",
		},
		{
			name: "test_13_list_directory",
			tool: "list_directory",
			arguments: map[string]interface{}{
				"path":        "C:\\",
				"max_results": 10,
			},
			description: "列出目录内容 (C:\\)",
		},
		{
			name: "test_14_get_file_info",
			tool: "get_file_info",
			arguments: map[string]interface{}{
				"path": "C:\\Windows\\System32\\notepad.exe",
			},
			description: "获取文件信息 (notepad.exe)",
		},
	}
	
	// 执行所有测试
	successCount := 0
	failCount := 0
	
	fmt.Println("=== 开始测试所有工具 ===")
	fmt.Println()
	
	for i, tc := range testCases {
		fmt.Printf("%d. %s\n", i+1, tc.description)
		
		params := map[string]interface{}{
			"name":      tc.tool,
			"arguments": tc.arguments,
		}
		
		response, err := client.SendRequest(ctx, "tools/call", params)
		if err != nil {
			fmt.Printf("   ❌ 请求失败: %v\n", err)
			failCount++
			fmt.Println()
			continue
		}
		
		if response.Error != nil {
			fmt.Printf("   ❌ 工具错误: %v\n", response.Error.Message)
			failCount++
			fmt.Println()
			continue
		}
		
		var callResult map[string]interface{}
		if err := json.Unmarshal(response.Result, &callResult); err != nil {
			fmt.Printf("   ❌ 解析结果失败: %v\n", err)
			failCount++
			fmt.Println()
			continue
		}
		
		// 检查是否有错误
		if isError, ok := callResult["isError"].(bool); ok && isError {
			if content, ok := callResult["content"].([]interface{}); ok && len(content) > 0 {
				if itemMap, ok := content[0].(map[string]interface{}); ok {
					if text, ok := itemMap["text"].(string); ok {
						fmt.Printf("   ❌ 工具返回错误: %s\n", text)
					}
				}
			}
			failCount++
			fmt.Println()
			continue
		}
		
		// 显示结果
		fmt.Printf("   ✅ 调用成功\n")
		if content, ok := callResult["content"].([]interface{}); ok && len(content) > 0 {
			if itemMap, ok := content[0].(map[string]interface{}); ok {
				if text, ok := itemMap["text"].(string); ok {
					// 限制输出长度
					lines := []rune(text)
					maxLen := 300
					if len(lines) > maxLen {
						preview := string(lines[:maxLen])
						lineCount := len([]rune(text)) / 50 // 粗略估计行数
						fmt.Printf("   📄 返回数据: %d+ 字符 (约 %d 行)\n", len(lines), lineCount)
						fmt.Printf("   预览:\n")
						// 显示前几行
						previewLines := []string{}
						for _, line := range []rune(preview) {
							if line == '\n' {
								if len(previewLines) >= 3 {
									break
								}
								previewLines = append(previewLines, "")
							}
						}
						fmt.Printf("   %s...\n", preview[:min(200, len(preview))])
					} else {
						fmt.Printf("   📄 返回数据: %d 字符\n", len(lines))
						if len(text) > 0 {
							// 只显示前3行
							allLines := splitLines(text)
							displayLines := allLines
							if len(allLines) > 3 {
								displayLines = allLines[:3]
								fmt.Printf("   预览 (前3行):\n")
								for _, line := range displayLines {
									fmt.Printf("      %s\n", line)
								}
								fmt.Printf("      ... (共 %d 行)\n", len(allLines))
							} else {
								for _, line := range displayLines {
									fmt.Printf("      %s\n", line)
								}
							}
						}
					}
				}
			}
		}
		successCount++
		fmt.Println()
		
		// 稍微等待一下，避免请求过快
		time.Sleep(100 * time.Millisecond)
	}
	
	fmt.Println("=== 测试完成 ===")
	fmt.Printf("✅ 成功: %d/%d\n", successCount, len(testCases))
	fmt.Printf("❌ 失败: %d/%d\n", failCount, len(testCases))
	
	if failCount > 0 {
		os.Exit(1)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func splitLines(text string) []string {
	lines := []string{}
	current := ""
	for _, char := range text {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
