package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// EverythingConfig 配置 Everything HTTP API 的地址
type EverythingConfig struct {
	BaseURL  string
	Port     int
	Username string
	Password string
	Timeout  time.Duration
}

// DefaultConfig 返回默认配置
func DefaultConfig() *EverythingConfig {
	return &EverythingConfig{
		BaseURL: "http://192.168.7.187",
		Port:    51780,
		Timeout: 10 * time.Second,
	}
}

// EverythingSearcher 定义搜索接口，便于测试
type EverythingSearcher interface {
	Search(ctx context.Context, query string, maxResults int) ([]SearchResult, error)
}

// EverythingClient Everything HTTP API 客户端
type EverythingClient struct {
	config *EverythingConfig
	client *http.Client
}

// NewEverythingClient 创建新的 Everything 客户端
func NewEverythingClient(config *EverythingConfig) *EverythingClient {
	if config == nil {
		config = DefaultConfig()
	}
	return &EverythingClient{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// SearchResult Everything 搜索结果项
type SearchResult struct {
	Path     string `json:"path"`
	Size     int64  `json:"size,omitempty"`
	Date     string `json:"date,omitempty"`
	Type     string `json:"type,omitempty"`
	FullPath string `json:"full_path,omitempty"`
}

// Search 执行文件搜索
func (c *EverythingClient) Search(ctx context.Context, query string, maxResults int) ([]SearchResult, error) {
	var baseURL string
	// 如果 BaseURL 已经包含协议（http:// 或 https://），直接使用
	if strings.HasPrefix(c.config.BaseURL, "http://") || strings.HasPrefix(c.config.BaseURL, "https://") {
		baseURL = c.config.BaseURL
		// 检查 URL 中是否已经包含端口号（在协议之后）
		// 例如: http://192.168.7.187:51780 vs http://192.168.7.187
		urlWithoutProtocol := strings.TrimPrefix(strings.TrimPrefix(c.config.BaseURL, "https://"), "http://")
		if c.config.Port != 0 && !strings.Contains(urlWithoutProtocol, ":") {
			// URL 中没有端口，但配置了端口，需要添加
			baseURL = fmt.Sprintf("%s:%d", c.config.BaseURL, c.config.Port)
		}
	} else {
		baseURL = fmt.Sprintf("%s:%d", c.config.BaseURL, c.config.Port)
	}

	// Everything HTTP API 使用 /?search= 参数
	params := url.Values{}
	params.Add("search", query)
	params.Add("json", "1") // 请求 JSON 格式输出
	if maxResults > 0 {
		params.Add("count", fmt.Sprintf("%d", maxResults)) // Everything 使用 count 参数限制结果数量
	}

	searchURL := fmt.Sprintf("%s/?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 如果配置了用户名和密码，添加 HTTP Basic Auth
	// 只有当用户名和密码都不为空时才添加认证头
	if c.config.Username != "" && c.config.Password != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)

		// 调试：输出认证信息（仅在调试模式下）
		if os.Getenv("EVERYTHING_DEBUG") == "true" {
			authHeader := req.Header.Get("Authorization")
			fmt.Fprintf(os.Stderr, "[DEBUG] 请求 URL: %s\n", searchURL)
			fmt.Fprintf(os.Stderr, "[DEBUG] 用户名: %s\n", c.config.Username)
			fmt.Fprintf(os.Stderr, "[DEBUG] 密码长度: %d\n", len(c.config.Password))
			fmt.Fprintf(os.Stderr, "[DEBUG] Authorization 头: %s\n", authHeader)
		}
	} else {
		// 如果用户名或密码为空，但仍然收到 401 错误，说明服务器需要认证
		// 这种情况下，我们应该返回一个更明确的错误信息
		if os.Getenv("EVERYTHING_DEBUG") == "true" {
			fmt.Fprintf(os.Stderr, "[DEBUG] 警告: 未设置认证信息\n")
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// 如果是 401 错误，提供更详细的错误信息
		if resp.StatusCode == http.StatusUnauthorized {
			hasAuth := c.config.Username != "" && c.config.Password != ""
			if !hasAuth {
				return nil, fmt.Errorf("HTTP 错误 401: 服务器需要认证，但未提供用户名和密码。请设置 EVERYTHING_USERNAME 和 EVERYTHING_PASSWORD 环境变量")
			}
			return nil, fmt.Errorf("HTTP 错误 401: 认证失败。请检查用户名和密码是否正确（当前用户名: %s）", c.config.Username)
		}
		return nil, fmt.Errorf("HTTP 错误 %d: %s", resp.StatusCode, string(body))
	}

	// Everything HTTP API 返回 JSON 格式（因为我们添加了 json=1 参数）
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 尝试解析为 JSON 格式
	var jsonResponse struct {
		TotalResults int `json:"totalResults"`
		Results      []struct {
			Type string `json:"type"`
			Name string `json:"name"`
			Path string `json:"path"`
			Size int64  `json:"size,omitempty"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		// 如果 JSON 解析失败，尝试作为文本格式处理（向后兼容）
		lines := strings.Split(strings.TrimSpace(string(body)), "\n")
		results := make([]SearchResult, 0, len(lines))
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				results = append(results, SearchResult{
					Path:     line,
					FullPath: line,
				})
			}
		}
		return results, nil
	}

	// 解析 JSON 结果
	results := make([]SearchResult, 0, len(jsonResponse.Results))
	for _, item := range jsonResponse.Results {
		// 构建完整路径
		fullPath := item.Path
		if item.Path != "" && item.Name != "" {
			fullPath = item.Path + "\\" + item.Name
		} else if item.Name != "" {
			fullPath = item.Name
		}

		results = append(results, SearchResult{
			Path:     fullPath,
			Type:     item.Type,
			Size:     item.Size,
			FullPath: fullPath,
		})
	}

	return results, nil
}

// MCPEverythingServer MCP 服务器
type MCPEverythingServer struct {
	server *server.DefaultServer
	client EverythingSearcher
	config *EverythingConfig
}

// NewMCPEverythingServer 创建新的 MCP Everything 服务器
func NewMCPEverythingServer(config *EverythingConfig) *MCPEverythingServer {
	if config == nil {
		config = DefaultConfig()
	}

	mcpServer := server.NewDefaultServer("everything-mcp", "1.0.0")

	everythingClient := NewEverythingClient(config)

	s := &MCPEverythingServer{
		server: mcpServer,
		client: everythingClient,
		config: config,
	}

	// 注册工具处理器
	mcpServer.HandleListTools(s.handleListTools)
	mcpServer.HandleCallTool(s.handleCallTool)

	// 注册自定义初始化处理器，声明 tools capability
	mcpServer.HandleInitialize(s.handleInitialize)

	return s
}

// handleInitialize 处理初始化请求，声明 tools capability
func (s *MCPEverythingServer) handleInitialize(
	ctx context.Context,
	capabilities mcp.ClientCapabilities,
	clientInfo mcp.Implementation,
	protocolVersion string,
) (*mcp.InitializeResult, error) {
	// 使用客户端请求的协议版本，或者默认使用 "2024-11-05"
	// Cursor 需要特定的协议版本格式
	resultVersion := protocolVersion
	if resultVersion == "" || resultVersion == "1.0" {
		resultVersion = "2024-11-05"
	}

	return &mcp.InitializeResult{
		ServerInfo: mcp.Implementation{
			Name:    "everything-mcp",
			Version: "1.0.0",
		},
		ProtocolVersion: resultVersion,
		Capabilities: mcp.ServerCapabilities{
			Tools: &struct {
				ListChanged bool `json:"listChanged"`
			}{
				ListChanged: true,
			},
		},
	}, nil
}

// handleListTools 处理工具列表请求
func (s *MCPEverythingServer) handleListTools(
	ctx context.Context,
	cursor *string,
) (*mcp.ListToolsResult, error) {
	return &mcp.ListToolsResult{
		Tools: []mcp.Tool{
			{
				Name:        "search_files",
				Description: "搜索文件和文件夹。支持文件名、路径、扩展名等多种搜索方式。",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "搜索关键词，支持文件名、路径、扩展名等",
						},
						"max_results": map[string]interface{}{
							"type":        "integer",
							"description": "最大返回结果数量，默认 100",
							"default":     100,
						},
					},
				},
			},
			{
				Name:        "search_by_extension",
				Description: "按文件扩展名搜索文件。例如搜索所有 .txt 或 .pdf 文件。",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"extension": map[string]interface{}{
							"type":        "string",
							"description": "文件扩展名，例如: txt, pdf, jpg (不需要点号)",
						},
						"max_results": map[string]interface{}{
							"type":        "integer",
							"description": "最大返回结果数量，默认 100",
							"default":     100,
						},
					},
				},
			},
			{
				Name:        "search_by_path",
				Description: "在指定路径中搜索文件。可以结合关键词进行更精确的搜索。",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "搜索路径，例如: C:\\Users\\Documents",
						},
						"query": map[string]interface{}{
							"type":        "string",
							"description": "搜索关键词（可选）",
						},
						"max_results": map[string]interface{}{
							"type":        "integer",
							"description": "最大返回结果数量，默认 100",
							"default":     100,
						},
					},
				},
			},
			{
				Name:        "search_by_size",
				Description: "按文件大小搜索文件。可以搜索大于、小于或在特定范围内的文件。",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"size_min": map[string]interface{}{
							"type":        "string",
							"description": "最小文件大小，例如: 1MB, 100KB, 1GB",
						},
						"size_max": map[string]interface{}{
							"type":        "string",
							"description": "最大文件大小，例如: 10MB, 1GB",
						},
						"query": map[string]interface{}{
							"type":        "string",
							"description": "附加搜索关键词（可选）",
						},
						"max_results": map[string]interface{}{
							"type":        "integer",
							"description": "最大返回结果数量，默认 100",
							"default":     100,
						},
					},
				},
			},
			{
				Name:        "search_by_date",
				Description: "按日期搜索文件。可以搜索特定日期范围内修改或创建的文件。",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"date_type": map[string]interface{}{
							"type":        "string",
							"description": "日期类型: modified (修改日期) 或 created (创建日期)",
							"enum":        []string{"modified", "created"},
							"default":     "modified",
						},
						"date_from": map[string]interface{}{
							"type":        "string",
							"description": "开始日期，格式: YYYY-MM-DD，例如: 2024-01-01",
						},
						"date_to": map[string]interface{}{
							"type":        "string",
							"description": "结束日期，格式: YYYY-MM-DD，例如: 2024-12-31",
						},
						"query": map[string]interface{}{
							"type":        "string",
							"description": "附加搜索关键词（可选）",
						},
						"max_results": map[string]interface{}{
							"type":        "integer",
							"description": "最大返回结果数量，默认 100",
							"default":     100,
						},
					},
				},
			},
			{
				Name:        "search_recent_files",
				Description: "搜索最近修改的文件。快速查找最近工作的文件。",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"days": map[string]interface{}{
							"type":        "integer",
							"description": "最近多少天内修改的文件，默认 7 天",
							"default":     7,
						},
						"query": map[string]interface{}{
							"type":        "string",
							"description": "附加搜索关键词（可选）",
						},
						"max_results": map[string]interface{}{
							"type":        "integer",
							"description": "最大返回结果数量，默认 100",
							"default":     100,
						},
					},
				},
			},
			{
				Name:        "search_large_files",
				Description: "搜索大文件。快速找出占用空间较大的文件。",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"min_size": map[string]interface{}{
							"type":        "string",
							"description": "最小文件大小，默认 100MB。例如: 100MB, 1GB",
							"default":     "100MB",
						},
						"path": map[string]interface{}{
							"type":        "string",
							"description": "搜索路径（可选）",
						},
						"max_results": map[string]interface{}{
							"type":        "integer",
							"description": "最大返回结果数量，默认 100",
							"default":     100,
						},
					},
				},
			},
			{
				Name:        "search_empty_files",
				Description: "搜索空文件或空文件夹。帮助清理无用的文件。",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"type": map[string]interface{}{
							"type":        "string",
							"description": "搜索类型: file (空文件) 或 folder (空文件夹)",
							"enum":        []string{"file", "folder"},
							"default":     "file",
						},
						"path": map[string]interface{}{
							"type":        "string",
							"description": "搜索路径（可选）",
						},
						"max_results": map[string]interface{}{
							"type":        "integer",
							"description": "最大返回结果数量，默认 100",
							"default":     100,
						},
					},
				},
			},
			{
				Name:        "search_by_content_type",
				Description: "按内容类型搜索文件。例如：图片、视频、音频、文档、压缩包等。",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"content_type": map[string]interface{}{
							"type":        "string",
							"description": "内容类型: image, video, audio, document, archive, executable",
							"enum":        []string{"image", "video", "audio", "document", "archive", "executable"},
						},
						"query": map[string]interface{}{
							"type":        "string",
							"description": "附加搜索关键词（可选）",
						},
						"max_results": map[string]interface{}{
							"type":        "integer",
							"description": "最大返回结果数量，默认 100",
							"default":     100,
						},
					},
				},
			},
			{
				Name:        "search_with_regex",
				Description: "使用正则表达式搜索文件。适合复杂的文件名模式匹配。",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"regex": map[string]interface{}{
							"type":        "string",
							"description": "正则表达式模式，例如: .*\\.log$",
						},
						"path": map[string]interface{}{
							"type":        "string",
							"description": "搜索路径（可选）",
						},
						"max_results": map[string]interface{}{
							"type":        "integer",
							"description": "最大返回结果数量，默认 100",
							"default":     100,
						},
					},
				},
			},
			{
				Name:        "search_duplicate_names",
				Description: "搜索具有相同文件名的文件。帮助找出重复或同名文件。",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"filename": map[string]interface{}{
							"type":        "string",
							"description": "要搜索的文件名，例如: config.txt",
						},
						"max_results": map[string]interface{}{
							"type":        "integer",
							"description": "最大返回结果数量，默认 100",
							"default":     100,
						},
					},
				},
			},
			{
				Name:        "list_drives",
				Description: "列出所有驱动器（C:, D:, E: 等）。类似于查看此电脑中的所有驱动器。",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
			{
				Name:        "list_directory",
				Description: "列出指定目录的内容（文件和文件夹）。可以一步步浏览文件系统。",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "要浏览的目录路径，例如: C:\\, C:\\Users, D:\\Projects",
						},
						"max_results": map[string]interface{}{
							"type":        "integer",
							"description": "最大返回结果数量，默认 100",
							"default":     100,
						},
					},
				},
			},
			{
				Name:        "get_file_info",
				Description: "获取文件或文件夹的详细信息（大小、日期、类型等）。",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "文件或文件夹的完整路径",
						},
					},
				},
			},
		},
	}, nil
}

// handleCallTool 处理工具调用请求
func (s *MCPEverythingServer) handleCallTool(
	ctx context.Context,
	name string,
	args map[string]interface{},
) (*mcp.CallToolResult, error) {
	switch name {
	case "search_files":
		return s.handleSearchFiles(ctx, args)
	case "search_by_extension":
		return s.handleSearchByExtension(ctx, args)
	case "search_by_path":
		return s.handleSearchByPath(ctx, args)
	case "search_by_size":
		return s.handleSearchBySize(ctx, args)
	case "search_by_date":
		return s.handleSearchByDate(ctx, args)
	case "search_recent_files":
		return s.handleSearchRecentFiles(ctx, args)
	case "search_large_files":
		return s.handleSearchLargeFiles(ctx, args)
	case "search_empty_files":
		return s.handleSearchEmptyFiles(ctx, args)
	case "search_by_content_type":
		return s.handleSearchByContentType(ctx, args)
	case "search_with_regex":
		return s.handleSearchWithRegex(ctx, args)
	case "search_duplicate_names":
		return s.handleSearchDuplicateNames(ctx, args)
	case "list_drives":
		return s.handleListDrives(ctx, args)
	case "list_directory":
		return s.handleListDirectory(ctx, args)
	case "get_file_info":
		return s.handleGetFileInfo(ctx, args)
	default:
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("未知的工具: %s", name),
				},
			},
		}, nil
	}
}

// handleSearchFiles 处理文件搜索请求
func (s *MCPEverythingServer) handleSearchFiles(
	ctx context.Context,
	args map[string]interface{},
) (*mcp.CallToolResult, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "query 参数是必需的且必须是非空字符串",
				},
			},
		}, nil
	}

	maxResults := 100
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	results, err := s.client.Search(ctx, query, maxResults)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("搜索失败: %v", err),
				},
			},
		}, nil
	}

	// 格式化结果
	resultText := fmt.Sprintf("搜索查询: %s\n找到 %d 个结果:\n\n", query, len(results))
	for i, result := range results {
		if i >= maxResults {
			break
		}
		resultText += fmt.Sprintf("%d. %s\n", i+1, result.Path)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// handleSearchByExtension 处理按扩展名搜索请求
func (s *MCPEverythingServer) handleSearchByExtension(
	ctx context.Context,
	args map[string]interface{},
) (*mcp.CallToolResult, error) {
	extension, ok := args["extension"].(string)
	if !ok || extension == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "extension 参数是必需的且必须是非空字符串",
				},
			},
		}, nil
	}

	// 移除可能的点号
	extension = strings.TrimPrefix(extension, ".")

	maxResults := 100
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// Everything 支持 ext: 语法
	query := fmt.Sprintf("ext:%s", extension)
	results, err := s.client.Search(ctx, query, maxResults)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("搜索失败: %v", err),
				},
			},
		}, nil
	}

	// 格式化结果
	resultText := fmt.Sprintf("扩展名搜索: .%s\n找到 %d 个结果:\n\n", extension, len(results))
	for i, result := range results {
		if i >= maxResults {
			break
		}
		resultText += fmt.Sprintf("%d. %s\n", i+1, result.Path)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// handleSearchByPath 处理按路径搜索请求
func (s *MCPEverythingServer) handleSearchByPath(
	ctx context.Context,
	args map[string]interface{},
) (*mcp.CallToolResult, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "path 参数是必需的且必须是非空字符串",
				},
			},
		}, nil
	}

	maxResults := 100
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// 构建查询：路径 + 可选的关键词
	query := path
	if q, ok := args["query"].(string); ok && q != "" {
		query = fmt.Sprintf("%s %s", path, q)
	}

	results, err := s.client.Search(ctx, query, maxResults)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("搜索失败: %v", err),
				},
			},
		}, nil
	}

	// 格式化结果
	resultText := fmt.Sprintf("路径搜索: %s\n找到 %d 个结果:\n\n", path, len(results))
	for i, result := range results {
		if i >= maxResults {
			break
		}
		resultText += fmt.Sprintf("%d. %s\n", i+1, result.Path)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// Serve 启动 MCP 服务器
func (s *MCPEverythingServer) Serve() error {
	return serveStdioWithNotificationSupport(s.server)
}

// serveStdioWithNotificationSupport 自定义 stdio 服务器，正确处理通知
func serveStdioWithNotificationSupport(mcpServer *server.DefaultServer) error {
	// 复制 mcp-go 的 ServeStdio 实现，但添加通知支持
	reader := bufio.NewReader(os.Stdin)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	done := make(chan struct{})
	go func() {
		<-sigChan
		close(done)
		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-done:
			return nil
		default:
			// 读取一行
			readChan := make(chan string, 1)
			errChan := make(chan error, 1)

			go func() {
				line, err := reader.ReadString('\n')
				if err != nil {
					errChan <- err
					return
				}
				readChan <- line
			}()

			select {
			case <-ctx.Done():
				return nil
			case err := <-errChan:
				if err == io.EOF {
					return nil
				}
				return err
			case line := <-readChan:
				if err := handleMessageWithNotifications(ctx, mcpServer, line); err != nil {
					if err == io.EOF {
						return nil
					}
					// 对于通知错误，继续处理
					if strings.Contains(err.Error(), "notifications/initialized") {
						continue
					}
				}
			}
		}
	}
}

// handleMessageWithNotifications 处理消息，正确识别通知
func handleMessageWithNotifications(ctx context.Context, mcpServer *server.DefaultServer, line string) error {
	// 解析 JSON-RPC 消息
	var msg map[string]interface{}
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// 检查是否是通知（没有 id 字段或 id 为 null）
	_, hasID := msg["id"]
	idValue, idExists := msg["id"]
	isNotification := !hasID || (idExists && idValue == nil)

	method, ok := msg["method"].(string)
	if !ok {
		return fmt.Errorf("missing method field")
	}

	// 如果是通知，静默处理，不发送响应
	if isNotification {
		// 对于 notifications/initialized，直接忽略
		if method == "notifications/initialized" {
			return nil
		}
		// 其他通知也忽略
		return nil
	}

	// 对于请求，使用正常的处理流程
	// 将消息转换为 JSON-RPC 请求格式
	params, _ := msg["params"].(map[string]interface{})
	paramsBytes, _ := json.Marshal(params)
	if paramsBytes == nil {
		paramsBytes = []byte("{}")
	}

	// 调用服务器处理请求
	result, err := mcpServer.Request(ctx, method, json.RawMessage(paramsBytes))
	if err != nil {
		// 发送错误响应
		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      msg["id"],
			"error": map[string]interface{}{
				"code":    -32603,
				"message": err.Error(),
			},
		}
		responseBytes, _ := json.Marshal(response)
		fmt.Println(string(responseBytes))
		return err
	}

	// 发送成功响应
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      msg["id"],
		"result":  result,
	}
	responseBytes, _ := json.Marshal(response)
	fmt.Println(string(responseBytes))

	return nil
}

// handleSearchBySize 处理按大小搜索请求
func (s *MCPEverythingServer) handleSearchBySize(
	ctx context.Context,
	args map[string]interface{},
) (*mcp.CallToolResult, error) {
	sizeMin, _ := args["size_min"].(string)
	sizeMax, _ := args["size_max"].(string)
	query, _ := args["query"].(string)

	maxResults := 100
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// 构建 Everything 搜索语法
	searchQuery := ""
	if sizeMin != "" {
		searchQuery += fmt.Sprintf("size:>%s ", sizeMin)
	}
	if sizeMax != "" {
		searchQuery += fmt.Sprintf("size:<%s ", sizeMax)
	}
	if query != "" {
		searchQuery += query
	}

	if searchQuery == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "至少需要提供 size_min 或 size_max 参数",
				},
			},
		}, nil
	}

	results, err := s.client.Search(ctx, strings.TrimSpace(searchQuery), maxResults)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("搜索失败: %v", err),
				},
			},
		}, nil
	}

	resultText := fmt.Sprintf("大小搜索: %s\n找到 %d 个结果:\n\n", searchQuery, len(results))
	for i, result := range results {
		if i >= maxResults {
			break
		}
		sizeStr := ""
		if result.Size > 0 {
			sizeStr = fmt.Sprintf(" (%s)", formatFileSize(result.Size))
		}
		resultText += fmt.Sprintf("%d. %s%s\n", i+1, result.Path, sizeStr)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// handleSearchByDate 处理按日期搜索请求
func (s *MCPEverythingServer) handleSearchByDate(
	ctx context.Context,
	args map[string]interface{},
) (*mcp.CallToolResult, error) {
	dateType, _ := args["date_type"].(string)
	if dateType == "" {
		dateType = "modified"
	}
	dateFrom, _ := args["date_from"].(string)
	dateTo, _ := args["date_to"].(string)
	query, _ := args["query"].(string)

	maxResults := 100
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// 构建 Everything 搜索语法
	searchQuery := ""
	prefix := "dm:"
	if dateType == "created" {
		prefix = "dc:"
	}

	if dateFrom != "" && dateTo != "" {
		searchQuery = fmt.Sprintf("%s%s..%s", prefix, dateFrom, dateTo)
	} else if dateFrom != "" {
		searchQuery = fmt.Sprintf("%s>%s", prefix, dateFrom)
	} else if dateTo != "" {
		searchQuery = fmt.Sprintf("%s<%s", prefix, dateTo)
	}

	if searchQuery == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "至少需要提供 date_from 或 date_to 参数",
				},
			},
		}, nil
	}

	if query != "" {
		searchQuery += " " + query
	}

	results, err := s.client.Search(ctx, searchQuery, maxResults)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("搜索失败: %v", err),
				},
			},
		}, nil
	}

	resultText := fmt.Sprintf("日期搜索 (%s): %s\n找到 %d 个结果:\n\n", dateType, searchQuery, len(results))
	for i, result := range results {
		if i >= maxResults {
			break
		}
		resultText += fmt.Sprintf("%d. %s\n", i+1, result.Path)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// handleSearchRecentFiles 处理最近文件搜索请求
func (s *MCPEverythingServer) handleSearchRecentFiles(
	ctx context.Context,
	args map[string]interface{},
) (*mcp.CallToolResult, error) {
	days := 7
	if d, ok := args["days"].(float64); ok {
		days = int(d)
	}
	query, _ := args["query"].(string)

	maxResults := 100
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// 构建 Everything 搜索语法：最近N天修改的文件
	searchQuery := fmt.Sprintf("dm:last%ddays", days)
	if query != "" {
		searchQuery += " " + query
	}

	results, err := s.client.Search(ctx, searchQuery, maxResults)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("搜索失败: %v", err),
				},
			},
		}, nil
	}

	resultText := fmt.Sprintf("最近 %d 天修改的文件\n找到 %d 个结果:\n\n", days, len(results))
	for i, result := range results {
		if i >= maxResults {
			break
		}
		resultText += fmt.Sprintf("%d. %s\n", i+1, result.Path)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// handleSearchLargeFiles 处理大文件搜索请求
func (s *MCPEverythingServer) handleSearchLargeFiles(
	ctx context.Context,
	args map[string]interface{},
) (*mcp.CallToolResult, error) {
	minSize, _ := args["min_size"].(string)
	if minSize == "" {
		minSize = "100MB"
	}
	path, _ := args["path"].(string)

	maxResults := 100
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// 构建 Everything 搜索语法
	searchQuery := fmt.Sprintf("size:>%s", minSize)
	if path != "" {
		searchQuery += fmt.Sprintf(" path:\"%s\"", path)
	}

	results, err := s.client.Search(ctx, searchQuery, maxResults)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("搜索失败: %v", err),
				},
			},
		}, nil
	}

	resultText := fmt.Sprintf("大文件搜索 (>%s)\n找到 %d 个结果:\n\n", minSize, len(results))
	for i, result := range results {
		if i >= maxResults {
			break
		}
		sizeStr := ""
		if result.Size > 0 {
			sizeStr = fmt.Sprintf(" (%s)", formatFileSize(result.Size))
		}
		resultText += fmt.Sprintf("%d. %s%s\n", i+1, result.Path, sizeStr)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// handleSearchEmptyFiles 处理空文件搜索请求
func (s *MCPEverythingServer) handleSearchEmptyFiles(
	ctx context.Context,
	args map[string]interface{},
) (*mcp.CallToolResult, error) {
	fileType, _ := args["type"].(string)
	if fileType == "" {
		fileType = "file"
	}
	path, _ := args["path"].(string)

	maxResults := 100
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// 构建 Everything 搜索语法
	searchQuery := ""
	if fileType == "folder" {
		searchQuery = "folder: empty:"
	} else {
		searchQuery = "file: size:0"
	}

	if path != "" {
		searchQuery += fmt.Sprintf(" path:\"%s\"", path)
	}

	results, err := s.client.Search(ctx, searchQuery, maxResults)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("搜索失败: %v", err),
				},
			},
		}, nil
	}

	typeStr := "空文件"
	if fileType == "folder" {
		typeStr = "空文件夹"
	}
	resultText := fmt.Sprintf("%s搜索\n找到 %d 个结果:\n\n", typeStr, len(results))
	for i, result := range results {
		if i >= maxResults {
			break
		}
		resultText += fmt.Sprintf("%d. %s\n", i+1, result.Path)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// handleSearchByContentType 处理按内容类型搜索请求
func (s *MCPEverythingServer) handleSearchByContentType(
	ctx context.Context,
	args map[string]interface{},
) (*mcp.CallToolResult, error) {
	contentType, ok := args["content_type"].(string)
	if !ok || contentType == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "content_type 参数是必需的",
				},
			},
		}, nil
	}
	query, _ := args["query"].(string)

	maxResults := 100
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// 定义内容类型对应的扩展名
	extMap := map[string]string{
		"image":      "ext:jpg;jpeg;png;gif;bmp;webp;svg;ico",
		"video":      "ext:mp4;avi;mkv;mov;wmv;flv;webm;m4v",
		"audio":      "ext:mp3;wav;flac;aac;ogg;wma;m4a",
		"document":   "ext:doc;docx;pdf;txt;rtf;odt;xls;xlsx;ppt;pptx",
		"archive":    "ext:zip;rar;7z;tar;gz;bz2;xz",
		"executable": "ext:exe;msi;bat;cmd;sh;app;dmg",
	}

	extQuery, exists := extMap[contentType]
	if !exists {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("不支持的内容类型: %s", contentType),
				},
			},
		}, nil
	}

	searchQuery := extQuery
	if query != "" {
		searchQuery += " " + query
	}

	results, err := s.client.Search(ctx, searchQuery, maxResults)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("搜索失败: %v", err),
				},
			},
		}, nil
	}

	resultText := fmt.Sprintf("内容类型搜索: %s\n找到 %d 个结果:\n\n", contentType, len(results))
	for i, result := range results {
		if i >= maxResults {
			break
		}
		resultText += fmt.Sprintf("%d. %s\n", i+1, result.Path)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// handleSearchWithRegex 处理正则表达式搜索请求
func (s *MCPEverythingServer) handleSearchWithRegex(
	ctx context.Context,
	args map[string]interface{},
) (*mcp.CallToolResult, error) {
	regex, ok := args["regex"].(string)
	if !ok || regex == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "regex 参数是必需的",
				},
			},
		}, nil
	}
	path, _ := args["path"].(string)

	maxResults := 100
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// 构建 Everything 搜索语法
	searchQuery := fmt.Sprintf("regex:%s", regex)
	if path != "" {
		searchQuery += fmt.Sprintf(" path:\"%s\"", path)
	}

	results, err := s.client.Search(ctx, searchQuery, maxResults)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("搜索失败: %v", err),
				},
			},
		}, nil
	}

	resultText := fmt.Sprintf("正则表达式搜索: %s\n找到 %d 个结果:\n\n", regex, len(results))
	for i, result := range results {
		if i >= maxResults {
			break
		}
		resultText += fmt.Sprintf("%d. %s\n", i+1, result.Path)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// handleSearchDuplicateNames 处理重复文件 名搜索请求
func (s *MCPEverythingServer) handleSearchDuplicateNames(
	ctx context.Context,
	args map[string]interface{},
) (*mcp.CallToolResult, error) {
	filename, ok := args["filename"].(string)
	if !ok || filename == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "filename 参数是必需的",
				},
			},
		}, nil
	}

	maxResults := 100
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// 搜索精确文件名
	searchQuery := fmt.Sprintf("file:%s", filename)

	results, err := s.client.Search(ctx, searchQuery, maxResults)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("搜索失败: %v", err),
				},
			},
		}, nil
	}

	resultText := fmt.Sprintf("重复文件名搜索: %s\n找到 %d 个结果:\n\n", filename, len(results))
	for i, result := range results {
		if i >= maxResults {
			break
		}
		resultText += fmt.Sprintf("%d. %s\n", i+1, result.Path)
	}

	if len(results) > 1 {
		resultText += fmt.Sprintf("\n发现 %d 个同名文件！\n", len(results))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// handleListDrives 处理列出驱动器请求
func (s *MCPEverythingServer) handleListDrives(
	ctx context.Context,
	args map[string]interface{},
) (*mcp.CallToolResult, error) {
	// 搜索所有根目录（驱动器）
	// Everything 语法: root: 表示搜索所有驱动器根目录
	searchQuery := "root:"

	results, err := s.client.Search(ctx, searchQuery, 100)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("获取驱动器列表失败: %v", err),
				},
			},
		}, nil
	}

	// 过滤出驱动器（通常是单个字母后跟冒号）
	drives := []SearchResult{}
	for _, result := range results {
		// 驱动器格式通常是 "C:", "D:" 等
		if len(result.Path) <= 3 && strings.HasSuffix(result.Path, ":") {
			drives = append(drives, result)
		}
	}

	resultText := fmt.Sprintf("系统驱动器列表\n找到 %d 个驱动器:\n\n", len(drives))
	for i, drive := range drives {
		resultText += fmt.Sprintf("%d. %s\\\n", i+1, drive.Path)
	}

	if len(drives) == 0 {
		resultText += "提示: 使用 list_directory 工具浏览特定驱动器，例如: C:\\, D:\\\n"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// handleListDirectory 处理列出目录内容请求
func (s *MCPEverythingServer) handleListDirectory(
	ctx context.Context,
	args map[string]interface{},
) (*mcp.CallToolResult, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "path 参数是必需的",
				},
			},
		}, nil
	}

	maxResults := 100
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// 规范化路径
	path = strings.TrimSpace(path)
	if !strings.HasSuffix(path, "\\") && !strings.HasSuffix(path, "/") {
		path += "\\"
	}

	// 构建搜索查询：查找指定路径下的直接子项
	// parent: 语法可以查找指定目录的直接子项
	searchQuery := fmt.Sprintf("parent:\"%s\"", strings.TrimSuffix(path, "\\"))

	results, err := s.client.Search(ctx, searchQuery, maxResults)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("浏览目录失败: %v", err),
				},
			},
		}, nil
	}

	// 分类为文件夹和文件
	folders := []SearchResult{}
	files := []SearchResult{}
	for _, result := range results {
		if result.Type == "folder" {
			folders = append(folders, result)
		} else {
			files = append(files, result)
		}
	}

	resultText := fmt.Sprintf("目录浏览: %s\n", path)
	resultText += fmt.Sprintf("找到 %d 个文件夹, %d 个文件\n\n", len(folders), len(files))

	// 显示文件夹
	if len(folders) > 0 {
		resultText += "📁 文件夹:\n"
		for i, folder := range folders {
			if i >= maxResults/2 {
				resultText += fmt.Sprintf("... 还有 %d 个文件夹\n", len(folders)-i)
				break
			}
			// 只显示文件夹名称，不显示完整路径
			name := strings.TrimPrefix(folder.Path, path)
			if name == "" {
				name = folder.Path
			}
			resultText += fmt.Sprintf("%d. 📁 %s\n", i+1, name)
		}
		resultText += "\n"
	}

	// 显示文件
	if len(files) > 0 {
		resultText += "📄 文件:\n"
		count := 0
		for i, file := range files {
			if count >= maxResults/2 {
				resultText += fmt.Sprintf("... 还有 %d 个文件\n", len(files)-i)
				break
			}
			name := strings.TrimPrefix(file.Path, path)
			if name == "" {
				name = file.Path
			}
			sizeStr := ""
			if file.Size > 0 {
				sizeStr = fmt.Sprintf(" (%s)", formatFileSize(file.Size))
			}
			resultText += fmt.Sprintf("%d. 📄 %s%s\n", i+1, name, sizeStr)
			count++
		}
	}

	if len(folders) == 0 && len(files) == 0 {
		resultText += "该目录为空或不存在\n"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// handleGetFileInfo 处理获取文件信息请求
func (s *MCPEverythingServer) handleGetFileInfo(
	ctx context.Context,
	args map[string]interface{},
) (*mcp.CallToolResult, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "path 参数是必需的",
				},
			},
		}, nil
	}

	// 使用精确路径搜索
	searchQuery := fmt.Sprintf("\"%s\"", path)

	results, err := s.client.Search(ctx, searchQuery, 1)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("获取文件信息失败: %v", err),
				},
			},
		}, nil
	}

	if len(results) == 0 {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("文件或文件夹不存在: %s", path),
				},
			},
		}, nil
	}

	result := results[0]
	resultText := fmt.Sprintf("文件信息: %s\n\n", result.Path)
	resultText += fmt.Sprintf("类型: %s\n", result.Type)
	if result.Size > 0 {
		resultText += fmt.Sprintf("大小: %s (%d 字节)\n", formatFileSize(result.Size), result.Size)
	} else if result.Type == "file" {
		resultText += "大小: 0 字节 (空文件)\n"
	}
	if result.Date != "" {
		resultText += fmt.Sprintf("修改日期: %s\n", result.Date)
	}
	resultText += fmt.Sprintf("完整路径: %s\n", result.FullPath)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func main() {
	// 从环境变量读取配置
	baseURL := os.Getenv("EVERYTHING_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost"
	}

	port := 80
	if portStr := os.Getenv("EVERYTHING_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	username := os.Getenv("EVERYTHING_USERNAME")
	password := os.Getenv("EVERYTHING_PASSWORD")

	config := &EverythingConfig{
		BaseURL:  baseURL,
		Port:     port,
		Username: username,
		Password: password,
		Timeout:  10 * time.Second,
	}

	// 创建并启动服务器
	server := NewMCPEverythingServer(config)

	// 注意：不要输出到 stderr，因为 MCP 协议使用 stdio 进行 JSON-RPC 通信
	// 输出到 stderr 可能会干扰通信
	// 如果需要调试，可以通过环境变量控制
	if os.Getenv("EVERYTHING_DEBUG") == "true" {
		fmt.Fprintf(os.Stderr, "Everything MCP Server 启动中...\n")
		fmt.Fprintf(os.Stderr, "Everything HTTP API: %s:%d\n", config.BaseURL, config.Port)
		if username != "" {
			fmt.Fprintf(os.Stderr, "用户名已配置: %s\n", username)
		} else {
			fmt.Fprintf(os.Stderr, "警告: 未配置用户名\n")
		}
		if password != "" {
			fmt.Fprintf(os.Stderr, "密码已配置: %s\n", strings.Repeat("*", len(password)))
		} else {
			fmt.Fprintf(os.Stderr, "警告: 未配置密码\n")
		}
	}

	if err := server.Serve(); err != nil {
		// 错误信息输出到 stderr 是安全的
		log.Printf("服务器错误: %v\n", err)
		os.Exit(1)
	}
}
