.PHONY: all build test clean install run help

# 变量定义
BINARY_NAME=everything-mcp
TEST_CLIENT=test-client
GO=go
GOFLAGS=-v
LDFLAGS=-s -w

# 默认目标
all: clean build

# 构建主程序
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/everything-mcp

# 构建测试客户端
build-test-client:
	@echo "Building $(TEST_CLIENT)..."
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(TEST_CLIENT) ./cmd/test-client

# 构建所有程序
build-all: build build-test-client

# 运行测试
test:
	@echo "Running tests..."
	$(GO) test -v -cover ./...

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 清理构建产物
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME) $(TEST_CLIENT)
	rm -f coverage.out coverage.html
	rm -f *.test

# 安装到 GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install ./cmd/everything-mcp

# 运行主程序
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# 运行测试客户端
run-test: build-all
	@echo "Running test client..."
	./$(TEST_CLIENT) config.json

# 格式化代码
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# 代码检查
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# 显示帮助信息
help:
	@echo "Available targets:"
	@echo "  all              - Clean and build (default)"
	@echo "  build            - Build the main program"
	@echo "  build-test-client - Build the test client"
	@echo "  build-all        - Build all programs"
	@echo "  test             - Run tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  clean            - Remove build artifacts"
	@echo "  install          - Install to GOPATH/bin"
	@echo "  run              - Build and run the main program"
	@echo "  run-test         - Build and run the test client"
	@echo "  fmt              - Format code"
	@echo "  lint             - Run linter"
	@echo "  help             - Show this help message"
