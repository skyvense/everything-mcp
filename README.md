# Everything MCP Server

[English](README.md) | [简体中文](README.zh-CN.md)

An MCP (Model Context Protocol) server implementation in Go for calling Everything's HTTP API, enabling LLM agents to perform natural language file search and browsing operations.

> 🚀 **Quick Start**: Check out the [Quick Start Guide](docs/QUICK_START.md) to get started immediately

## Features

- 🔍 **14 Powerful Tools**: Rich file search and browsing capabilities
  - **11 Search Tools**: Basic search, extension, path, size, date, content type, etc.
  - **3 Browse Tools**: Drive list, directory browsing, file information
- 📁 **File System Browsing**: Navigate the file system like Windows Explorer
  - List all drives (C:, D:, E:)
  - Browse directories level by level
  - View detailed file information
- 🚀 **High Performance**: Leverage Everything's lightning-fast indexing
- 💬 **Natural Language**: LLM agents can find files using natural language
- 🔐 **Authentication**: Support for HTTP Basic Authentication
- 📊 **JSON Format**: Returns structured JSON data
- 🎯 **Precise Matching**: Full support for Everything search syntax

## Prerequisites

1. **Everything Software**: Install and run [Everything](https://www.voidtools.com/)
2. **Enable HTTP Server**: Enable HTTP server in Everything
   - Open Everything → Tools → Options
   - Select "HTTP Server" page
   - Enable HTTP Server
   - Set port (default 80)
   - (Optional) Enable authentication

## Installation

### Build from Source

```bash
git clone https://github.com/skyvense/everything-mcp.git
cd everything-mcp

# Build with Makefile (recommended)
make build

# Or build directly with go
go build -o everything-mcp ./cmd/everything-mcp
```

### Install with Go

```bash
go install github.com/skyvense/everything-mcp/cmd/everything-mcp@latest
```

## Configuration

### Environment Variables

- `EVERYTHING_BASE_URL`: Everything HTTP API base URL (default: `http://localhost`)
- `EVERYTHING_PORT`: Everything HTTP API port (default: `80`)
- `EVERYTHING_USERNAME`: Everything HTTP API username (optional, if authentication is enabled)
- `EVERYTHING_PASSWORD`: Everything HTTP API password (optional, if authentication is enabled)
- `EVERYTHING_DEBUG`: Enable debug logs (set to `true` to see detailed request information)

### Example Configuration

```bash
# Basic configuration (no authentication)
export EVERYTHING_BASE_URL="http://localhost"
export EVERYTHING_PORT="80"

# Configuration with authentication
export EVERYTHING_BASE_URL="http://192.168.7.187"
export EVERYTHING_PORT="1780"
export EVERYTHING_USERNAME="your_username"
export EVERYTHING_PASSWORD="your_password"

# Enable debug mode
export EVERYTHING_DEBUG="true"
```

## Usage

### Using Startup Script (Recommended)

The project provides convenient startup scripts with pre-configured settings:

```bash
# Linux/macOS
./scripts/start.sh
```

The startup script will automatically:
- Set environment variables (URL, port, username, password)
- Check and compile the program if needed
- Display configuration information
- Start the server

### Using Makefile

```bash
# Build main program
make build

# Build all programs (including test client)
make build-all

# Run main program
make run

# Run tests
make test

# View all available commands
make help
```

### Direct Run

If you need custom configuration, set environment variables and run directly:

```bash
export EVERYTHING_BASE_URL="http://192.168.7.187"
export EVERYTHING_PORT="1780"
export EVERYTHING_USERNAME="your_username"
export EVERYTHING_PASSWORD="your_password"

./everything-mcp
```

The server will communicate with MCP clients via stdio.

### Configure in MCP Client

#### Cursor IDE

Add to Cursor's MCP configuration file (usually at `~/.cursor/mcp.json` or via settings UI):

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

Add to Claude Desktop's configuration file:
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`

Use the same JSON format.

**Notes**: 
- If Everything HTTP server doesn't have authentication enabled, omit `EVERYTHING_USERNAME` and `EVERYTHING_PASSWORD`
- Ensure `command` points to the actual executable location
- Add `"EVERYTHING_DEBUG": "true"` to `env` for debugging

## Available Tools

Everything MCP Server provides **14 powerful tools**:

### Output Format

All search tools return results with the following information:
- **Path**: Full path of the file or folder
- **Type**: file or folder
- **Size**: File size (folders show `-`)
- **Modified Time**: Last modification date and time

### Search Tools (11)

#### Basic Search
1. **search_files** - Basic file search (returns: path, type, size, date)
2. **search_by_extension** - Search by extension (returns: path, size, date)
3. **search_by_path** - Search by path (returns: path, type, size, date)

#### Advanced Search
4. **search_by_size** - Search by file size (returns: path, size, date)
5. **search_by_date** - Search by date (returns: path, size, date)
6. **search_recent_files** - Search recently modified files (returns: path, size, date)
7. **search_large_files** - Search large files (returns: path, size, date)
8. **search_empty_files** - Search empty files/folders (returns: path, size, date)

#### Professional Search
9. **search_by_content_type** - Search by content type (returns: path, size, date)
10. **search_with_regex** - Regular expression search (returns: path, size, date)
11. **search_duplicate_names** - Search duplicate filenames (returns: path, size, date)

### Browse Tools (3)

12. **list_drives** - List all drives
13. **list_directory** - Browse directory contents (returns: name, type icon, size, date)
14. **get_file_info** - Get detailed file information (returns: type, size, date, full path)

### Quick Examples

**Search Examples**:
```json
// Search PDFs modified in last 7 days
{
  "name": "search_recent_files",
  "arguments": {
    "days": 7,
    "query": "ext:pdf"
  }
}

// Search video files larger than 100MB
{
  "name": "search_by_content_type",
  "arguments": {
    "content_type": "video",
    "query": "size:>100MB"
  }
}
```

**Browse Examples**:
```json
// List all drives
{
  "name": "list_drives",
  "arguments": {}
}

// Browse C drive
{
  "name": "list_directory",
  "arguments": {
    "path": "C:\\"
  }
}

// Get file information
{
  "name": "get_file_info",
  "arguments": {
    "path": "C:\\Users\\Documents\\report.pdf"
  }
}
```

**Full Documentation**: See [TOOLS.md](docs/TOOLS.md) for complete tool descriptions and usage examples.

## Usage Examples

### Using with LLM Agent

LLM agents can call these tools using natural language:

**Basic Search**:
- "Find all PDF files"
- "Search for files containing 'report' in Documents folder"
- "Find all .txt files"

**Advanced Search**:
- "Find files modified in the last 3 days"
- "Search for files larger than 100MB"
- "Find all empty folders"
- "Find all documents created in 2024"

**Professional Search**:
- "Search for all image files"
- "Find all files named config.json"
- "Use regex to search all .log files"
- "Find the 20 largest files"

**File System Browsing**:
- "Show all drives"
- "Browse C drive contents"
- "Enter Documents folder"
- "Show detailed info for this file"

## Technical Details

### Everything HTTP API

Everything's HTTP API uses simple GET requests:

```
GET http://localhost:80/?search=<query>&json=1&count=<max_results>
```

**Important Parameters:**
- `search`: Search query string
- `json=1`: Request JSON format response (recommended)
- `count`: Limit number of results
- `path`: Specify search path

**JSON Response Format:**
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

### MCP Protocol

This server implements the MCP (Model Context Protocol) standard:
- **Communication**: Communicates with clients via stdio
- **Protocol**: JSON-RPC 2.0
- **Protocol Version**: 2024-11-05
- **Supported Capabilities**: Tools (tool calling)

## Development

### Project Structure

```
everything-mcp/
├── cmd/                           # Executable programs
│   ├── everything-mcp/           # Main program
│   │   └── main.go
│   └── test-client/              # Test client
│       └── main.go
├── docs/                          # Documentation
│   ├── QUICK_START.md            # Quick start guide
│   ├── USAGE.md                  # Detailed usage
│   ├── TOOLS.md                  # Tool list and usage
│   └── PROJECT_STRUCTURE.md      # Project structure
├── examples/                      # Example configurations
│   └── mcp-config-example.json   # MCP configuration example
├── scripts/                       # Scripts
│   ├── start.sh                  # Startup script
│   └── test-mcp.sh               # Test script
├── go.mod                         # Go module definition
├── go.sum                         # Go dependency checksums
├── Makefile                       # Build script
├── README.md                      # This document (English)
├── README.zh-CN.md                # Chinese documentation
└── .gitignore                     # Git ignore rules
```

### Dependencies

- `github.com/mark3labs/mcp-go`: MCP protocol Go implementation

### Build

```bash
# Build with Makefile (recommended)
make build

# Or use go build
go build -o everything-mcp ./cmd/everything-mcp

# Build test client
make build-test-client
# Or
go build -o test-client ./cmd/test-client
```

### Testing

#### Unit Tests

Run all unit tests:

```bash
# Using Makefile
make test

# Or using go test
go test -v ./...
```

View test coverage:

```bash
# Generate HTML coverage report
make test-coverage

# Or use go test
go test -cover ./...
```

Run specific tests:

```bash
go test -v -run TestEverythingClient_Search ./cmd/everything-mcp
```

Current test coverage: **79%**, including:
- EverythingClient search functionality tests
- MCP server tool list and handling tests
- Complete tests for all three search tools
- Error handling and boundary case tests
- HTTP authentication tests

#### Integration Tests

Use test client for end-to-end testing:

```bash
# Using Makefile (recommended)
make run-test

# Or manually compile and run
make build-test-client
./test-client examples/mcp-config-example.json
```

Test client will automatically:
1. Start MCP server
2. Execute complete MCP protocol handshake
3. Test all available tools
4. Display detailed test results

See [docs/USAGE.md](docs/USAGE.md) for more information.

### Run Server

Ensure Everything's HTTP server is running, then:

```bash
# Using Makefile
make run

# Or run directly
./everything-mcp
```

## Troubleshooting

### HTTP 401 Authentication Error

If you encounter "HTTP error 401: Authentication failed":

1. **Check username and password**
   ```bash
   # Test authentication with curl
   curl -u username:password "http://host:port/?search=test&json=1"
   ```

2. **Verify Everything HTTP server configuration**
   - Open Everything → Tools → Options → HTTP Server
   - Check if "Require username and password" is enabled
   - Confirm username and password settings

3. **Check port configuration**
   - Ensure `EVERYTHING_PORT` matches the port configured in Everything
   - URL should be in format `http://host:port` (port must be correct)

4. **Enable debug mode**
   ```bash
   export EVERYTHING_DEBUG="true"
   ./everything-mcp
   ```
   View detailed request information including URL, auth headers, etc.

### Connection Error

If you encounter connection errors, check:

1. Is Everything running?
2. Is HTTP server enabled?
3. Is port configuration correct (including in URL)?
4. Is firewall blocking the connection?
5. If remote server, check network connectivity and server accessibility

**Verify connection:**
```bash
# Test basic connection (no auth)
curl "http://localhost:80/?search=test&json=1"

# Test connection with auth
curl -u username:password "http://host:port/?search=test&json=1"
```

### No Search Results

- Ensure Everything has indexed your file system
- Check if search query is correct
- Try searching directly in Everything interface to verify
- Use `json=1` parameter to ensure JSON format is returned

### MCP Client Connection Issues

If MCP client cannot connect to server:

1. **Check executable path**
   - Ensure `command` path in configuration file is correct
   - Use absolute path instead of relative path

2. **Check environment variables**
   - Confirm all required environment variables are set
   - Especially `EVERYTHING_BASE_URL` and `EVERYTHING_PORT`

3. **View logs**
   - Add `EVERYTHING_DEBUG=true` to environment variables
   - Check client log output

4. **Test server**
   - Verify server functionality with test client:
     ```bash
     make run-test
     ```

### Common Questions

**Q: Why does search return HTML instead of file list?**

A: Need to add `json=1` parameter in request. This server handles it automatically. If still issues, check if Everything version supports JSON output.

**Q: How to limit search result count?**

A: Use `max_results` parameter, server will automatically convert to Everything API's `count` parameter.

**Q: What search syntax is supported?**

A: Supports Everything's complete search syntax, including:
- Wildcards: `*.txt`
- Path search: `C:\Users\Documents\`
- Extension: `ext:pdf`
- Regular expressions: `regex:.*\.log$`
- More syntax: [Everything Search Syntax](https://www.voidtools.com/support/everything/searching/)

## License

MIT License

## Contributing

Issues and Pull Requests are welcome!

## Changelog

### v1.2.0 (2026-01-12)

- ✨ **Added 3 file system browsing tools**, total 14 tools
  - `list_drives` - List all drives
  - `list_directory` - Browse directory contents
  - `get_file_info` - Get detailed file information
- 🎯 Support browsing file system like Explorer
- 📁 Can browse directories level by level from drives
- 📊 Display detailed file and folder information

### v1.1.0 (2026-01-12)

- ✨ **Added 8 search tools**, total 11 tools
  - `search_by_size` - Search by file size
  - `search_by_date` - Search by date
  - `search_recent_files` - Search recently modified files
  - `search_large_files` - Search large files
  - `search_empty_files` - Search empty files/folders
  - `search_by_content_type` - Search by content type
  - `search_with_regex` - Regular expression search
  - `search_duplicate_names` - Search duplicate filenames
- ✨ Added file size formatting
- 📝 Added complete tool documentation (TOOLS.md)

### v1.0.1 (2026-01-12)

- 🐛 Fixed URL port not being added correctly
- 🐛 Fixed Everything HTTP API returning HTML instead of JSON
- ✨ Added JSON format support (`json=1` parameter)
- ✨ Added debug mode (`EVERYTHING_DEBUG` environment variable)
- ✨ Improved error messages, especially 401 auth errors
- 📝 Added test client (`test-client`)
- 📝 Improved documentation and troubleshooting guide
- 🏗️ Refactored project structure, adopted standard Go project layout
- 🔧 Added Makefile to simplify build and test

### v1.0.0 (2026-01-11)

- 🎉 Initial release
- ✨ Implemented three search tools: `search_files`, `search_by_extension`, `search_by_path`
- ✨ Support HTTP Basic Authentication
- ✨ Support MCP protocol 2024-11-05
- 📝 Complete unit test coverage

## Related Links

- [Everything Official Website](https://www.voidtools.com/)
- [Everything HTTP API Documentation](https://www.voidtools.com/support/everything/http/)
- [Everything Search Syntax](https://www.voidtools.com/support/everything/searching/)
- [MCP Protocol Specification](https://modelcontextprotocol.io/)
- [mcp-go Library](https://github.com/mark3labs/mcp-go)

## Acknowledgments

- [Everything](https://www.voidtools.com/) - Fast file search tool
- [mcp-go](https://github.com/mark3labs/mcp-go) - Go implementation of MCP protocol
- [Anthropic](https://www.anthropic.com/) - Creator of MCP protocol
