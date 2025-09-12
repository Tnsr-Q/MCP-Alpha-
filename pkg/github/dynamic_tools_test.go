package github

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/github/github-mcp-server/pkg/toolsets"
	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a mock translation helper
func mockTranslationHelper() translations.TranslationHelperFunc {
	return func(key string, fallback string) string {
		return fallback
	}
}

// Helper function to create a test toolset group
func createTestToolsetGroup() *toolsets.ToolsetGroup {
	tsg := toolsets.NewToolsetGroup(false)

	// Add some test toolsets
	repos := toolsets.NewToolset("repos", "GitHub Repository related tools")
	repos.AddReadTools(
		toolsets.NewServerTool(
			mcp.NewTool("list_repositories",
				mcp.WithDescription("Mock list repositories tool"),
				mcp.WithToolAnnotation(mcp.ToolAnnotation{
					ReadOnlyHint: ToBoolPtr(true),
				}),
			), 
			func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return mcp.NewToolResultText("mock repos"), nil
			},
		),
	)

	issues := toolsets.NewToolset("issues", "GitHub Issues related tools")
	issues.AddReadTools(
		toolsets.NewServerTool(
			mcp.NewTool("list_issues",
				mcp.WithDescription("Mock list issues tool"),
				mcp.WithToolAnnotation(mcp.ToolAnnotation{
					ReadOnlyHint: ToBoolPtr(true),
				}),
			), 
			func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return mcp.NewToolResultText("mock issues"), nil
			},
		),
	)

	// Add toolsets to group
	tsg.AddToolset(repos)
	tsg.AddToolset(issues)

	return tsg
}

func TestListAvailableToolsets(t *testing.T) {
	tsg := createTestToolsetGroup()
	translator := mockTranslationHelper()

	tool, handler := ListAvailableToolsets(tsg, translator)

	// Test tool properties
	assert.Equal(t, "list_available_toolsets", tool.Name)
	assert.NotEmpty(t, tool.Description)

	// Test handler with empty request
	request := mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name:      "list_available_toolsets",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := handler(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse the JSON result
	var toolsets []map[string]string
	err = json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &toolsets)
	require.NoError(t, err)

	// Verify we have the expected toolsets
	assert.Len(t, toolsets, 2)

	// Check repos toolset
	reposToolset := findToolsetByName(toolsets, "repos")
	require.NotNil(t, reposToolset, "repos toolset should be found")
	assert.Equal(t, "repos", (*reposToolset)["name"])
	assert.Equal(t, "GitHub Repository related tools", (*reposToolset)["description"])
	assert.Equal(t, "true", (*reposToolset)["can_enable"])
	assert.Equal(t, "false", (*reposToolset)["currently_enabled"]) // Should be disabled by default

	// Check issues toolset
	issuesToolset := findToolsetByName(toolsets, "issues")
	require.NotNil(t, issuesToolset, "issues toolset should be found")
	assert.Equal(t, "issues", (*issuesToolset)["name"])
	assert.Equal(t, "GitHub Issues related tools", (*issuesToolset)["description"])
	assert.Equal(t, "true", (*issuesToolset)["can_enable"])
	assert.Equal(t, "false", (*issuesToolset)["currently_enabled"]) // Should be disabled by default
}

func TestListAvailableToolsetsWithEnabledToolset(t *testing.T) {
	tsg := createTestToolsetGroup()
	translator := mockTranslationHelper()

	// Enable one toolset
	err := tsg.EnableToolset("repos")
	require.NoError(t, err)

	_, handler := ListAvailableToolsets(tsg, translator)

	request := mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name:      "list_available_toolsets",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := handler(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse the JSON result
	var toolsets []map[string]string
	err = json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &toolsets)
	require.NoError(t, err)

	// Find the repos toolset and verify it's enabled
	reposToolset := findToolsetByName(toolsets, "repos")
	require.NotNil(t, reposToolset, "repos toolset should be found")
	assert.Equal(t, "true", (*reposToolset)["currently_enabled"]) // Should be enabled now

	// Find the issues toolset and verify it's still disabled
	issuesToolset := findToolsetByName(toolsets, "issues")
	require.NotNil(t, issuesToolset, "issues toolset should be found")
	assert.Equal(t, "false", (*issuesToolset)["currently_enabled"]) // Should still be disabled
}

func TestGetToolsetsTools(t *testing.T) {
	tsg := createTestToolsetGroup()
	translator := mockTranslationHelper()

	tool, handler := GetToolsetsTools(tsg, translator)

	// Test tool properties
	assert.Equal(t, "get_toolset_tools", tool.Name)
	assert.NotEmpty(t, tool.Description)

	// Test with valid toolset
	request := mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name: "get_toolset_tools",
			Arguments: map[string]interface{}{
				"toolset": "repos",
			},
		},
	}

	result, err := handler(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse the JSON result
	var tools []map[string]string
	err = json.Unmarshal([]byte(result.Content[0].(mcp.TextContent).Text), &tools)
	require.NoError(t, err)

	// Verify we have the expected tools
	assert.Len(t, tools, 1)
	assert.Equal(t, "list_repositories", tools[0]["name"])
	assert.Equal(t, "true", tools[0]["can_enable"])
	assert.Equal(t, "repos", tools[0]["toolset"])
}

func TestGetToolsetsToolsInvalidToolset(t *testing.T) {
	tsg := createTestToolsetGroup()
	translator := mockTranslationHelper()

	_, handler := GetToolsetsTools(tsg, translator)

	// Test with invalid toolset
	request := mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name: "get_toolset_tools",
			Arguments: map[string]interface{}{
				"toolset": "nonexistent",
			},
		},
	}

	result, err := handler(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return an error result
	assert.Len(t, result.Content, 1)
	errorContent := result.Content[0].(mcp.TextContent)
	assert.Contains(t, errorContent.Text, "Toolset nonexistent not found")
	assert.True(t, result.IsError)
}

func TestGetToolsetsToolsMissingParameter(t *testing.T) {
	tsg := createTestToolsetGroup()
	translator := mockTranslationHelper()

	_, handler := GetToolsetsTools(tsg, translator)

	// Test without toolset parameter
	request := mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name:      "get_toolset_tools",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := handler(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return an error result for missing parameter
	assert.Len(t, result.Content, 1)
	errorContent := result.Content[0].(mcp.TextContent)
	assert.Contains(t, errorContent.Text, "missing required parameter: toolset")
	assert.True(t, result.IsError)
}

func TestEnableToolset(t *testing.T) {
	tsg := createTestToolsetGroup()
	translator := mockTranslationHelper()

	// Create a mock MCP server
	mcpServer := server.NewMCPServer("test-server", "1.0.0")

	tool, handler := EnableToolset(mcpServer, tsg, translator)

	// Test tool properties
	assert.Equal(t, "enable_toolset", tool.Name)
	assert.NotEmpty(t, tool.Description)

	// Verify toolset is initially disabled
	assert.False(t, tsg.IsEnabled("repos"))

	// Test enabling valid toolset
	request := mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name: "enable_toolset",
			Arguments: map[string]interface{}{
				"toolset": "repos",
			},
		},
	}

	result, err := handler(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify success response
	assert.Len(t, result.Content, 1)
	textContent := result.Content[0].(mcp.TextContent)
	assert.Contains(t, textContent.Text, "Toolset repos enabled")

	// Verify toolset is now enabled
	assert.True(t, tsg.IsEnabled("repos"))
}

func TestEnableToolsetAlreadyEnabled(t *testing.T) {
	tsg := createTestToolsetGroup()
	translator := mockTranslationHelper()

	// Pre-enable the toolset
	err := tsg.EnableToolset("repos")
	require.NoError(t, err)

	// Create a mock MCP server
	mcpServer := server.NewMCPServer("test-server", "1.0.0")

	_, handler := EnableToolset(mcpServer, tsg, translator)

	// Test enabling already enabled toolset
	request := mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name: "enable_toolset",
			Arguments: map[string]interface{}{
				"toolset": "repos",
			},
		},
	}

	result, err := handler(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify response indicates already enabled
	assert.Len(t, result.Content, 1)
	textContent := result.Content[0].(mcp.TextContent)
	assert.Contains(t, textContent.Text, "Toolset repos is already enabled")
}

func TestEnableToolsetInvalidToolset(t *testing.T) {
	tsg := createTestToolsetGroup()
	translator := mockTranslationHelper()

	// Create a mock MCP server
	mcpServer := server.NewMCPServer("test-server", "1.0.0")

	_, handler := EnableToolset(mcpServer, tsg, translator)

	// Test enabling invalid toolset
	request := mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name: "enable_toolset",
			Arguments: map[string]interface{}{
				"toolset": "nonexistent",
			},
		},
	}

	result, err := handler(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return an error result
	assert.Len(t, result.Content, 1)
	errorContent := result.Content[0].(mcp.TextContent)
	assert.Contains(t, errorContent.Text, "Toolset nonexistent not found")
	assert.True(t, result.IsError)
}

func TestEnableToolsetMissingParameter(t *testing.T) {
	tsg := createTestToolsetGroup()
	translator := mockTranslationHelper()

	// Create a mock MCP server
	mcpServer := server.NewMCPServer("test-server", "1.0.0")

	_, handler := EnableToolset(mcpServer, tsg, translator)

	// Test without toolset parameter
	request := mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name:      "enable_toolset",
			Arguments: map[string]interface{}{},
		},
	}

	result, err := handler(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should return an error result for missing parameter
	assert.Len(t, result.Content, 1)
	errorContent := result.Content[0].(mcp.TextContent)
	assert.Contains(t, errorContent.Text, "missing required parameter: toolset")
	assert.True(t, result.IsError)
}

func TestToolsetEnum(t *testing.T) {
	tsg := createTestToolsetGroup()

	enumOption := ToolsetEnum(tsg)

	// The enum should contain all toolset names
	// Note: We can't easily inspect the internal enum values due to the mcp library's structure,
	// but we can verify it doesn't panic and returns a valid option
	assert.NotNil(t, enumOption)
}

func TestInitDynamicToolset(t *testing.T) {
	tsg := createTestToolsetGroup()
	translator := mockTranslationHelper()

	// Create a mock MCP server
	mcpServer := server.NewMCPServer("test-server", "1.0.0")

	// Initialize dynamic toolset
	dynamicToolset := InitDynamicToolset(mcpServer, tsg, translator)

	// Verify dynamic toolset properties
	assert.Equal(t, "dynamic", dynamicToolset.Name)
	assert.Contains(t, dynamicToolset.Description, "Discover GitHub MCP tools")
	assert.True(t, dynamicToolset.Enabled) // Should be enabled by default

	// Verify it has the expected tools
	tools := dynamicToolset.GetActiveTools()
	assert.Len(t, tools, 3) // Should have 3 tools: list_available_toolsets, get_toolset_tools, enable_toolset

	toolNames := make([]string, len(tools))
	for i, tool := range tools {
		toolNames[i] = tool.Tool.Name
	}

	assert.Contains(t, toolNames, "list_available_toolsets")
	assert.Contains(t, toolNames, "get_toolset_tools")
	assert.Contains(t, toolNames, "enable_toolset")
}

// Helper function to find a toolset by name in the JSON result
func findToolsetByName(toolsets []map[string]string, name string) *map[string]string {
	for _, ts := range toolsets {
		if ts["name"] == name {
			return &ts
		}
	}
	return nil
}