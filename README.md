# MCP PRIME - Repository to MCP Conversion Tool

MCP PRIME is a Model Context Protocol (MCP) server that converts any repository into an MCP-compatible tool. It provides a comprehensive toolkit for analyzing codebases and automatically generating MCP tool definitions from existing functions and classes.

### Use Cases

- **Repository Analysis**: Analyze any codebase to extract function signatures and documentation
- **MCP Tool Generation**: Automatically convert Python, JavaScript, and TypeScript functions into MCP tool definitions
- **Code Structure Discovery**: Navigate and understand project structure across any repository
- **Function Signature Extraction**: Parse source code to identify callable functions with their parameters and documentation
- **Tool Definition Export**: Generate OpenAI-compatible tool JSON definitions for AI integration

Built for developers who want to convert existing codebases into MCP-compatible tools, enabling AI agents to interact with any repository's functionality through natural language.

---

## Core Tools

MCP PRIME provides four essential tools for repository analysis and MCP conversion:

### 1. `get_file_list`
Return every file path in a repository with optional filtering and pagination.

**Parameters:**
- `per_page` (integer, default: 100) - Items per page (max 100)
- `page` (integer, default: 1) - Page number for pagination  
- `extension` (string, optional) - Filter by file extension (e.g., 'py', 'js', 'ts')

### 2. `get_file_content`
Return the UTF-8 decoded content of any file in the repository.

**Parameters:**
- `path` (string, required) - Repository-relative path (e.g., 'src/utils.py')

### 3. `extract_signatures`
Parse Python, JavaScript, or TypeScript source code and extract top-level function and class signatures with their docstrings.

**Parameters:**
- `code` (string, required) - Full source code to analyze
- `language` (string, required) - Language of the code ('python', 'javascript', 'typescript')

### 4. `emit_tool_json`
Convert a list of function/class descriptors into a JSON array of OpenAI-compatible tool descriptions.

**Parameters:**
- `functions` (array, required) - Function descriptors with name, description, parameters, and required fields

---

## Installation & Usage

### Prerequisites
- Go 1.23+ installed
- Compatible MCP host (Claude Desktop, VS Code, Cursor, etc.)

### Build from Source
```bash
git clone https://github.com/Tnsr-Q/MCP-PRIME.git
cd MCP-PRIME
go build -o mcp-prime ./cmd/mcp-prime
```

### Run as MCP Server
```bash
./mcp-prime stdio
```

### Example Configuration for Claude Desktop
Add to your Claude Desktop config:

```json
{
  "mcpServers": {
    "mcp-prime": {
      "command": "/path/to/mcp-prime",
      "args": ["stdio"]
    }
  }
}
```

---

## Typical Workflow

1. **List repository files**: Use `get_file_list(extension="py")` to find Python files
2. **Read source code**: For each file, call `get_file_content(path)`  
3. **Extract signatures**: Parse code with `extract_signatures(code, language="python")`
4. **Generate tools**: Convert results using `emit_tool_json(functions=...)`

This creates a complete MCP tool definition from any repository's codebase.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
