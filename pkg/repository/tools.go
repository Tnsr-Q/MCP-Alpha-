package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// GetFileList returns every file path in the repository with optional filtering and pagination
func GetFileList() server.ServerTool {
	tool, handler := getFileListImpl()
	return server.ServerTool{
		Tool:    tool,
		Handler: handler,
	}
}

func getFileListImpl() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_file_list",
			mcp.WithDescription("Return every file path in the default branch of the *current* repo (paginated)."),
			mcp.WithNumber("per_page",
				mcp.Description("Items per page (max 100)"),
				mcp.DefaultNumber(100),
			),
			mcp.WithNumber("page",
				mcp.Description("Page number"),
				mcp.DefaultNumber(1),
			),
			mcp.WithString("extension",
				mcp.Description("Optional filter, e.g. 'py', 'js', 'ts'"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return handleGetFileList(ctx, request)
		}
}

// GetFileContent returns the UTF-8 decoded content of any file in the repository
func GetFileContent() server.ServerTool {
	tool, handler := getFileContentImpl()
	return server.ServerTool{
		Tool:    tool,
		Handler: handler,
	}
}

func getFileContentImpl() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("get_file_content",
			mcp.WithDescription("Return the UTF-8 decoded content of any file in the current repo (default branch)."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Repository-relative path, e.g. 'src/utils.py'"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return handleGetFileContent(ctx, request)
		}
}

// ExtractSignatures parses source code and emits function/class signatures with docstrings
func ExtractSignatures() server.ServerTool {
	tool, handler := extractSignaturesImpl()
	return server.ServerTool{
		Tool:    tool,
		Handler: handler,
	}
}

func extractSignaturesImpl() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("extract_signatures",
			mcp.WithDescription("Parse Python or JavaScript/TypeScript source and emit every top-level function/class with its signature + docstring."),
			mcp.WithString("code",
				mcp.Required(),
				mcp.Description("Full source code to analyse"),
			),
			mcp.WithString("language",
				mcp.Required(),
				mcp.Description("Language of the code"),
				mcp.Enum("python", "javascript", "typescript"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return handleExtractSignatures(ctx, request)
		}
}

// EmitToolJSON converts function descriptors into OpenAI-style tool descriptions
func EmitToolJSON() server.ServerTool {
	tool, handler := emitToolJSONImpl()
	return server.ServerTool{
		Tool:    tool,
		Handler: handler,
	}
}

func emitToolJSONImpl() (mcp.Tool, server.ToolHandlerFunc) {
	return mcp.NewTool("emit_tool_json",
			mcp.WithDescription("Convert a list of function/class descriptors into a single JSON array of OpenAI-style tool descriptions."),
			mcp.WithArray("functions",
				mcp.Required(),
				mcp.Description("Each item must have: name, description, parameters (object), required (array[string])"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return handleEmitToolJSON(ctx, request)
		}
}

func handleGetFileList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	perPage, err := OptionalParam[float64](req, "per_page")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if perPage == 0 {
		perPage = 100
	}
	
	page, err := OptionalParam[float64](req, "page")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if page == 0 {
		page = 1
	}
	
	extension, err := OptionalParam[string](req, "extension")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Set limits
	if perPage > 100 {
		perPage = 100
	}
	if perPage <= 0 {
		perPage = 100
	}
	if page <= 0 {
		page = 1
	}

	// Get current working directory as repository root
	repoRoot, err := os.Getwd()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get working directory: %v", err)), nil
	}

	var allFiles []string

	// Walk through all files in the repository
	err = filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and common build/dependency directories
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			if name == "node_modules" || name == "vendor" || name == "__pycache__" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}

		// Skip hidden files
		if strings.HasPrefix(filepath.Base(relPath), ".") {
			return nil
		}

		// Filter by extension if specified
		if extension != "" {
			ext := strings.TrimPrefix(filepath.Ext(relPath), ".")
			if ext != extension {
				return nil
			}
		}

		allFiles = append(allFiles, relPath)
		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to walk directory: %v", err)), nil
	}

	// Calculate pagination
	start := (int(page) - 1) * int(perPage)
	end := start + int(perPage)

	if start >= len(allFiles) {
		return mcp.NewToolResultText("[]"), nil
	}

	if end > len(allFiles) {
		end = len(allFiles)
	}

	pageFiles := allFiles[start:end]

	// Convert to JSON
	result, err := json.MarshalIndent(pageFiles, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal file list: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

func handleGetFileContent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := RequiredParam[string](req, "path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get current working directory as repository root
	repoRoot, err := os.Getwd()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get working directory: %v", err)), nil
	}

	// Resolve the full path
	fullPath := filepath.Join(repoRoot, path)

	// Security: ensure the path is within the repository
	cleanPath, err := filepath.Abs(fullPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to resolve path: %v", err)), nil
	}

	cleanRepoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to resolve repository root: %v", err)), nil
	}

	if !strings.HasPrefix(cleanPath, cleanRepoRoot) {
		return mcp.NewToolResultError("path is outside repository bounds"), nil
	}

	// Read file content
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read file: %v", err)), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

func handleExtractSignatures(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, err := RequiredParam[string](req, "code")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	language, err := RequiredParam[string](req, "language")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var signatures []FunctionSignature

	switch language {
	case "python":
		signatures, err = extractPythonSignatures(code)
	case "javascript", "typescript":
		signatures, err = extractJavaScriptSignatures(code)
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unsupported language: %s", language)), nil
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to extract signatures: %v", err)), nil
	}

	result, err := json.MarshalIndent(signatures, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal signatures: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

func handleEmitToolJSON(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	functionsParam, ok := req.GetArguments()["functions"]
	if !ok {
		return mcp.NewToolResultError("functions parameter is required"), nil
	}

	// Convert to JSON bytes first then unmarshal to ensure proper type conversion
	functionsJSON, err := json.Marshal(functionsParam)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal functions parameter: %v", err)), nil
	}

	var functions []FunctionDescriptor
	if err := json.Unmarshal(functionsJSON, &functions); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to unmarshal functions: %v", err)), nil
	}

	if len(functions) == 0 {
		return mcp.NewToolResultError("functions parameter cannot be empty"), nil
	}

	tools := make([]ToolDefinition, len(functions))
	for i, fn := range functions {
		tools[i] = ToolDefinition{
			Type: "function",
			Function: FunctionDef{
				Name:        fn.Name,
				Description: fn.Description,
				Parameters:  fn.Parameters,
			},
		}
	}

	result, err := json.MarshalIndent(tools, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal tool definitions: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

// GetFileListTool returns the tool and handler separately for direct MCP server registration
func GetFileListTool() (mcp.Tool, server.ToolHandlerFunc) {
	return getFileListImpl()
}

// GetFileContentTool returns the tool and handler separately for direct MCP server registration
func GetFileContentTool() (mcp.Tool, server.ToolHandlerFunc) {
	return getFileContentImpl()
}

// ExtractSignaturesTool returns the tool and handler separately for direct MCP server registration
func ExtractSignaturesTool() (mcp.Tool, server.ToolHandlerFunc) {
	return extractSignaturesImpl()
}

// EmitToolJSONTool returns the tool and handler separately for direct MCP server registration
func EmitToolJSONTool() (mcp.Tool, server.ToolHandlerFunc) {
	return emitToolJSONImpl()
}

// Parameter helper functions (copied from github package)
func RequiredParam[T comparable](r mcp.CallToolRequest, p string) (T, error) {
	var zero T

	// Check if the parameter is present in the request
	if _, ok := r.GetArguments()[p]; !ok {
		return zero, fmt.Errorf("missing required parameter: %s", p)
	}

	// Check if the parameter is of the expected type
	val, ok := r.GetArguments()[p].(T)
	if !ok {
		return zero, fmt.Errorf("parameter %s is not of type %T", p, zero)
	}

	if val == zero {
		return zero, fmt.Errorf("missing required parameter: %s", p)
	}

	return val, nil
}

func OptionalParam[T any](r mcp.CallToolRequest, p string) (T, error) {
	var zero T

	// Check if the parameter is present in the request
	if _, ok := r.GetArguments()[p]; !ok {
		return zero, nil
	}

	// Check if the parameter is of the expected type
	if _, ok := r.GetArguments()[p].(T); !ok {
		return zero, fmt.Errorf("parameter %s is not of type %T, is %T", p, zero, r.GetArguments()[p])
	}

	return r.GetArguments()[p].(T), nil
}