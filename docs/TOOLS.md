# Everything MCP Server - 工具列表

本文档列出了 Everything MCP Server 提供的所有搜索工具及其使用方法。

## 返回数据格式

所有搜索工具返回的结果都包含以下信息：
- **路径**: 文件或文件夹的完整路径
- **类型**: file（文件）或 folder（文件夹）
- **大小**: 文件大小（文件夹显示为 `-`）
- **修改时间**: 最后修改日期和时间

示例输出：
```
1. C:\Users\Documents\report.pdf
   类型: file
   大小: 2.5 MB
   修改时间: 2024-01-15 10:30:45
```

## 工具总览

Everything MCP Server 现在提供 **14 个强大的工具**：

### 搜索工具 (11个)
1. **search_files** - 基本文件搜索
2. **search_by_extension** - 按扩展名搜索
3. **search_by_path** - 按路径搜索
4. **search_by_size** - 按文件大小搜索
5. **search_by_date** - 按日期搜索
6. **search_recent_files** - 搜索最近修改的文件
7. **search_large_files** - 搜索大文件
8. **search_empty_files** - 搜索空文件/文件夹
9. **search_by_content_type** - 按内容类型搜索
10. **search_with_regex** - 正则表达式搜索
11. **search_duplicate_names** - 搜索重复文件名

### 浏览工具 (3个)
12. **list_drives** - 列出所有驱动器
13. **list_directory** - 浏览目录内容
14. **get_file_info** - 获取文件详细信息

---

## 1. search_files

**描述**: 搜索文件和文件夹。支持文件名、路径、扩展名等多种搜索方式。

**返回信息**: 路径、类型(file/folder)、大小、修改时间

**参数**:
- `query` (string, 必需): 搜索关键词
- `max_results` (integer, 可选): 最大返回结果数量，默认 100

**使用示例**:
```json
{
  "name": "search_files",
  "arguments": {
    "query": "report",
    "max_results": 50
  }
}
```

**自然语言示例**:
- "帮我找所有包含 report 的文件"
- "搜索 config 文件"

---

## 2. search_by_extension

**描述**: 按文件扩展名搜索文件。例如搜索所有 .txt 或 .pdf 文件。

**返回信息**: 路径、大小、修改时间

**参数**:
- `extension` (string, 必需): 文件扩展名（不需要点号）
- `max_results` (integer, 可选): 最大返回结果数量，默认 100

**使用示例**:
```json
{
  "name": "search_by_extension",
  "arguments": {
    "extension": "pdf",
    "max_results": 20
  }
}
```

**自然语言示例**:
- "找出所有 PDF 文件"
- "搜索所有 .log 文件"

---

## 3. search_by_path

**描述**: 在指定路径中搜索文件。可以结合关键词进行更精确的搜索。

**返回信息**: 路径、类型(file/folder)、大小、修改时间

**参数**:
- `path` (string, 必需): 搜索路径
- `query` (string, 可选): 附加搜索关键词
- `max_results` (integer, 可选): 最大返回结果数量，默认 100

**使用示例**:
```json
{
  "name": "search_by_path",
  "arguments": {
    "path": "C:\\Users\\Documents",
    "query": "report",
    "max_results": 30
  }
}
```

**自然语言示例**:
- "在 Documents 文件夹中搜索 report"
- "查找 C:\\Projects 目录下的所有文件"

---

## 4. search_by_size

**描述**: 按文件大小搜索文件。可以搜索大于、小于或在特定范围内的文件。

**返回信息**: 路径、大小、修改时间

**参数**:
- `size_min` (string, 可选): 最小文件大小，例如: 1MB, 100KB, 1GB
- `size_max` (string, 可选): 最大文件大小，例如: 10MB, 1GB
- `query` (string, 可选): 附加搜索关键词
- `max_results` (integer, 可选): 最大返回结果数量，默认 100

**使用示例**:
```json
{
  "name": "search_by_size",
  "arguments": {
    "size_min": "10MB",
    "size_max": "100MB",
    "max_results": 50
  }
}
```

**自然语言示例**:
- "找出 10MB 到 100MB 之间的文件"
- "搜索大于 1GB 的文件"

---

## 5. search_by_date

**描述**: 按日期搜索文件。可以搜索特定日期范围内修改或创建的文件。

**返回信息**: 路径、大小、修改时间

**参数**:
- `date_type` (string, 可选): 日期类型，`modified` (修改日期) 或 `created` (创建日期)，默认 `modified`
- `date_from` (string, 可选): 开始日期，格式: YYYY-MM-DD
- `date_to` (string, 可选): 结束日期，格式: YYYY-MM-DD
- `query` (string, 可选): 附加搜索关键词
- `max_results` (integer, 可选): 最大返回结果数量，默认 100

**使用示例**:
```json
{
  "name": "search_by_date",
  "arguments": {
    "date_type": "modified",
    "date_from": "2024-01-01",
    "date_to": "2024-12-31",
    "max_results": 100
  }
}
```

**自然语言示例**:
- "找出 2024 年修改的所有文件"
- "搜索最近一个月创建的文件"

---

## 6. search_recent_files

**描述**: 搜索最近修改的文件。快速查找最近工作的文件。

**返回信息**: 路径、大小、修改时间

**参数**:
- `days` (integer, 可选): 最近多少天内修改的文件，默认 7 天
- `query` (string, 可选): 附加搜索关键词
- `max_results` (integer, 可选): 最大返回结果数量，默认 100

**使用示例**:
```json
{
  "name": "search_recent_files",
  "arguments": {
    "days": 3,
    "query": "report",
    "max_results": 50
  }
}
```

**自然语言示例**:
- "找出最近 3 天修改的文件"
- "搜索最近一周的 Word 文档"

---

## 7. search_large_files

**描述**: 搜索大文件。快速找出占用空间较大的文件。

**返回信息**: 路径、大小、修改时间

**参数**:
- `min_size` (string, 可选): 最小文件大小，默认 100MB
- `path` (string, 可选): 搜索路径
- `max_results` (integer, 可选): 最大返回结果数量，默认 100

**使用示例**:
```json
{
  "name": "search_large_files",
  "arguments": {
    "min_size": "500MB",
    "path": "C:\\Users",
    "max_results": 20
  }
}
```

**自然语言示例**:
- "找出所有大于 500MB 的文件"
- "在 C 盘查找占用空间最大的文件"

---

## 8. search_empty_files

**描述**: 搜索空文件或空文件夹。帮助清理无用的文件。

**返回信息**: 路径、大小、修改时间

**参数**:
- `type` (string, 可选): 搜索类型，`file` (空文件) 或 `folder` (空文件夹)，默认 `file`
- `path` (string, 可选): 搜索路径
- `max_results` (integer, 可选): 最大返回结果数量，默认 100

**使用示例**:
```json
{
  "name": "search_empty_files",
  "arguments": {
    "type": "file",
    "path": "C:\\Projects",
    "max_results": 50
  }
}
```

**自然语言示例**:
- "找出所有空文件"
- "搜索 Projects 目录下的空文件夹"

---

## 9. search_by_content_type

**描述**: 按内容类型搜索文件。例如：图片、视频、音频、文档、压缩包等。

**返回信息**: 路径、大小、修改时间

**参数**:
- `content_type` (string, 必需): 内容类型
  - `image`: 图片 (jpg, jpeg, png, gif, bmp, webp, svg, ico)
  - `video`: 视频 (mp4, avi, mkv, mov, wmv, flv, webm, m4v)
  - `audio`: 音频 (mp3, wav, flac, aac, ogg, wma, m4a)
  - `document`: 文档 (doc, docx, pdf, txt, rtf, odt, xls, xlsx, ppt, pptx)
  - `archive`: 压缩包 (zip, rar, 7z, tar, gz, bz2, xz)
  - `executable`: 可执行文件 (exe, msi, bat, cmd, sh, app, dmg)
- `query` (string, 可选): 附加搜索关键词
- `max_results` (integer, 可选): 最大返回结果数量，默认 100

**使用示例**:
```json
{
  "name": "search_by_content_type",
  "arguments": {
    "content_type": "image",
    "query": "screenshot",
    "max_results": 50
  }
}
```

**自然语言示例**:
- "找出所有图片文件"
- "搜索所有视频"
- "查找所有压缩包"

---

## 10. search_with_regex

**描述**: 使用正则表达式搜索文件。适合复杂的文件名模式匹配。

**返回信息**: 路径、大小、修改时间

**参数**:
- `regex` (string, 必需): 正则表达式模式
- `path` (string, 可选): 搜索路径
- `max_results` (integer, 可选): 最大返回结果数量，默认 100

**使用示例**:
```json
{
  "name": "search_with_regex",
  "arguments": {
    "regex": ".*\\.log$",
    "path": "C:\\Logs",
    "max_results": 100
  }
}
```

**自然语言示例**:
- "使用正则表达式搜索所有 .log 文件"
- "查找文件名符合特定模式的文件"

---

## 11. search_duplicate_names

**描述**: 搜索具有相同文件名的文件。帮助找出重复或同名文件。

**返回信息**: 路径、大小、修改时间

**参数**:
- `filename` (string, 必需): 要搜索的文件名
- `max_results` (integer, 可选): 最大返回结果数量，默认 100

**使用示例**:
```json
{
  "name": "search_duplicate_names",
  "arguments": {
    "filename": "config.txt",
    "max_results": 50
  }
}
```

**自然语言示例**:
- "找出所有名为 config.txt 的文件"
- "搜索重复的 README.md 文件"

---

## 12. list_drives

**描述**: 列出所有驱动器（C:, D:, E: 等）。类似于查看"此电脑"中的所有驱动器。

**参数**: 无

**使用示例**:
```json
{
  "name": "list_drives",
  "arguments": {}
}
```

**自然语言示例**:
- "显示所有驱动器"
- "列出电脑上的所有磁盘"
- "查看有哪些盘符"

---

## 13. list_directory

**描述**: 列出指定目录的内容（文件和文件夹）。可以一步步浏览文件系统。

**返回信息**: 名称、类型(📁文件夹/📄文件)、大小、修改时间

**参数**:
- `path` (string, 必需): 要浏览的目录路径，例如: C:\\, C:\\Users, D:\\Projects
- `max_results` (integer, 可选): 最大返回结果数量，默认 100

**使用示例**:
```json
{
  "name": "list_directory",
  "arguments": {
    "path": "C:\\Users\\Documents",
    "max_results": 50
  }
}
```

**自然语言示例**:
- "显示 C 盘的内容"
- "浏览 Documents 文件夹"
- "查看 D:\\Projects 目录下有什么"

**特点**:
- 自动分类显示文件夹和文件
- 显示文件大小和修改时间
- 使用图标区分文件夹 📁 和文件 📄
- 支持逐级浏览

---

## 14. get_file_info

**描述**: 获取文件或文件夹的详细信息（大小、日期、类型等）。

**参数**:
- `path` (string, 必需): 文件或文件夹的完整路径

**使用示例**:
```json
{
  "name": "get_file_info",
  "arguments": {
    "path": "C:\\Users\\Documents\\report.pdf"
  }
}
```

**自然语言示例**:
- "查看这个文件的详细信息"
- "获取 report.pdf 的属性"
- "显示文件夹大小"

**返回信息**:
- 文件类型（文件/文件夹）
- 文件大小（格式化显示）
- 修改日期
- 完整路径

---

## 浏览工作流示例

### 从驱动器开始浏览

```
1. 用户: "显示所有驱动器"
   工具: list_drives
   结果: C:\, D:\, E:\

2. 用户: "浏览 C 盘"
   工具: list_directory
   参数: { path: "C:\\" }
   结果: 显示 C 盘根目录的文件夹和文件

3. 用户: "进入 Users 文件夹"
   工具: list_directory
   参数: { path: "C:\\Users" }
   结果: 显示 Users 目录的内容

4. 用户: "查看某个文件的详细信息"
   工具: get_file_info
   参数: { path: "C:\\Users\\Documents\\file.txt" }
   结果: 显示文件的详细信息
```

---

## Everything 搜索语法

所有工具都支持 Everything 的强大搜索语法：

### 基本语法
- `*.txt` - 通配符
- `"exact match"` - 精确匹配
- `file1|file2` - 或运算
- `file1 file2` - 与运算
- `!file` - 非运算

### 高级语法
- `ext:pdf` - 按扩展名
- `size:>10MB` - 按大小
- `dm:today` - 今天修改
- `dc:lastweek` - 上周创建
- `path:"C:\Users"` - 指定路径
- `regex:.*\.log$` - 正则表达式

### 日期语法
- `dm:today` - 今天
- `dm:yesterday` - 昨天
- `dm:lastweek` - 上周
- `dm:last7days` - 最近 7 天
- `dm:2024` - 2024 年
- `dm:2024-01-01..2024-12-31` - 日期范围

### 大小语法
- `size:0` - 空文件
- `size:<1MB` - 小于 1MB
- `size:>100MB` - 大于 100MB
- `size:10MB..100MB` - 大小范围

更多语法请参考: https://www.voidtools.com/zh-cn/support/everything/searching/

---

## 使用技巧

### 1. 组合使用
可以在 `query` 参数中组合多个搜索条件：
```json
{
  "name": "search_files",
  "arguments": {
    "query": "report ext:pdf dm:lastweek"
  }
}
```

### 2. 文件大小单位
支持的大小单位：
- `B` - 字节
- `KB` - 千字节
- `MB` - 兆字节
- `GB` - 吉字节
- `TB` - 太字节

### 3. 日期格式
- 绝对日期: `YYYY-MM-DD` (例如: 2024-01-01)
- 相对日期: `today`, `yesterday`, `lastweek`, `last7days`

### 4. 路径格式
- Windows: `C:\\Users\\Documents`
- 使用引号包含空格: `path:"C:\\Program Files"`

---

## 性能提示

1. **限制结果数量**: 使用 `max_results` 参数限制返回结果，提高响应速度
2. **精确搜索**: 使用更具体的搜索条件减少结果数量
3. **路径限制**: 在特定路径下搜索比全盘搜索更快
4. **索引更新**: 确保 Everything 的索引是最新的

---

## 错误处理

所有工具在出错时会返回包含错误信息的响应：
```json
{
  "isError": true,
  "content": [
    {
      "type": "text",
      "text": "错误信息"
    }
  ]
}
```

常见错误：
- 参数缺失或格式错误
- Everything HTTP API 连接失败
- 认证失败 (HTTP 401)
- 搜索语法错误

---

## 更新日志

### v1.1.0 (2026-01-12)
- ✨ 新增 8 个搜索工具
- ✨ 支持按大小、日期、内容类型搜索
- ✨ 支持正则表达式和重复文件名搜索
- ✨ 添加文件大小格式化显示

### v1.0.0 (2026-01-11)
- 🎉 初始版本
- ✨ 基础搜索工具：search_files, search_by_extension, search_by_path

---

## 相关文档

- [README.md](../README.md) - 项目概览
- [QUICK_START.md](QUICK_START.md) - 快速开始
- [USAGE.md](USAGE.md) - 详细使用说明
- [Everything 搜索语法](https://www.voidtools.com/zh-cn/support/everything/searching/)
