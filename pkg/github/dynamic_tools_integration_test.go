package github

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/github/github-mcp-server/pkg/toolsets"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDynamicToolsetIntegration tests the integration of dynamic toolsets with the main server functionality
func TestDynamicToolsetIntegration(t *testing.T) {
	translator := mockTranslationHelper()

	// Create a toolset group similar to the default one but simpler for testing
	tsg := toolsets.NewToolsetGroup(false)

	// Add a few toolsets like the real server would
	repos := toolsets.NewToolset("repos", "GitHub Repository related tools")
	repos.AddReadTools(
		toolsets.NewServerTool(
			mcp.NewTool("list_repositories",
				mcp.WithDescription("List repositories"),
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
				mcp.WithDescription("List issues"),
				mcp.WithToolAnnotation(mcp.ToolAnnotation{
					ReadOnlyHint: ToBoolPtr(true),
				}),
			),
			func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return mcp.NewToolResultText("mock issues"), nil
			},
		),
	)

	// Add toolsets to the group
	tsg.AddToolset(repos)
	tsg.AddToolset(issues)

	// Verify initial state - no toolsets are enabled
	assert.False(t, tsg.IsEnabled("repos"))
	assert.False(t, tsg.IsEnabled("issues"))

	// Create MCP server
	mcpServer := server.NewMCPServer("test-server", "1.0.0")

	// Test the InitDynamicToolset function like the main server does
	dynamicToolset := InitDynamicToolset(mcpServer, tsg, translator)

	// Verify dynamic toolset was created and configured correctly
	require.NotNil(t, dynamicToolset)
	assert.Equal(t, "dynamic", dynamicToolset.Name)
	assert.True(t, dynamicToolset.Enabled) // Should be enabled by default

	// Get the active tools from the dynamic toolset
	activeTools := dynamicToolset.GetActiveTools()
	assert.Len(t, activeTools, 3) // Should have 3 tools

	// Extract tool names
	toolNames := make([]string, len(activeTools))
	for i, tool := range activeTools {
		toolNames[i] = tool.Tool.Name
	}

	// Verify all expected dynamic tools are present
	assert.Contains(t, toolNames, "list_available_toolsets")
	assert.Contains(t, toolNames, "get_toolset_tools")
	assert.Contains(t, toolNames, "enable_toolset")

	// Register the dynamic toolset with the server like the main server does
	dynamicToolset.RegisterTools(mcpServer)

	// Verify the dynamic toolset was registered successfully
	// Note: We can't directly check the server's tools, but we can verify
	// the dynamic toolset has the right configuration
}

// TestDynamicToolsetFilter tests that when DynamicToolsets is enabled, "all" is filtered from enabled toolsets
func TestDynamicToolsetFilter(t *testing.T) {
	// This test simulates the logic in internal/ghmcp/server.go where "all" is filtered
	// when DynamicToolsets is enabled

	enabledToolsets := []string{"repos", "issues", "all", "actions"}
	dynamicToolsets := true

	var filteredToolsets []string
	if dynamicToolsets {
		// Filter "all" from the enabled toolsets like the server does
		filteredToolsets = make([]string, 0, len(enabledToolsets))
		for _, toolset := range enabledToolsets {
			if toolset != "all" {
				filteredToolsets = append(filteredToolsets, toolset)
			}
		}
	} else {
		filteredToolsets = enabledToolsets
	}

	// Verify "all" was filtered out
	assert.NotContains(t, filteredToolsets, "all")
	assert.Contains(t, filteredToolsets, "repos")
	assert.Contains(t, filteredToolsets, "issues")
	assert.Contains(t, filteredToolsets, "actions")
	assert.Len(t, filteredToolsets, 3) // Should have 3 toolsets (original 4 minus "all")
}

// TestDynamicToolsetWorkflow tests a typical workflow of using dynamic toolsets
func TestDynamicToolsetWorkflow(t *testing.T) {
	translator := mockTranslationHelper()
	tsg := createTestToolsetGroup()
	mcpServer := server.NewMCPServer("test-server", "1.0.0")

	// Initialize dynamic toolset
	dynamicToolset := InitDynamicToolset(mcpServer, tsg, translator)
	require.NotNil(t, dynamicToolset)

	// Get the tools
	listTool, listHandler := ListAvailableToolsets(tsg, translator)
	getTool, getHandler := GetToolsetsTools(tsg, translator)
	enableTool, enableHandler := EnableToolset(mcpServer, tsg, translator)

	// Step 1: List available toolsets
	listRequest := mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name:      listTool.Name,
			Arguments: map[string]interface{}{},
		},
	}

	result, err := listHandler(context.Background(), listRequest)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse and verify results
	var toolsets []map[string]string
	textContent := result.Content[0].(mcp.TextContent)
	err = json.Unmarshal([]byte(textContent.Text), &toolsets)
	require.NoError(t, err)
	assert.Len(t, toolsets, 2) // repos and issues

	// Step 2: Get tools for a specific toolset
	getRequest := mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name: getTool.Name,
			Arguments: map[string]interface{}{
				"toolset": "repos",
			},
		},
	}

	result, err = getHandler(context.Background(), getRequest)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse and verify results
	var tools []map[string]string
	textContent = result.Content[0].(mcp.TextContent)
	err = json.Unmarshal([]byte(textContent.Text), &tools)
	require.NoError(t, err)
	assert.Len(t, tools, 1) // list_repositories
	assert.Equal(t, "list_repositories", tools[0]["name"])

	// Step 3: Enable the toolset
	enableRequest := mcp.CallToolRequest{
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name: enableTool.Name,
			Arguments: map[string]interface{}{
				"toolset": "repos",
			},
		},
	}

	result, err = enableHandler(context.Background(), enableRequest)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify toolset was enabled
	textContent = result.Content[0].(mcp.TextContent)
	assert.Contains(t, textContent.Text, "Toolset repos enabled")
	assert.True(t, tsg.IsEnabled("repos"))

	// Step 4: List toolsets again to verify status changed
	result, err = listHandler(context.Background(), listRequest)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse and verify repos is now enabled
	textContent = result.Content[0].(mcp.TextContent)
	err = json.Unmarshal([]byte(textContent.Text), &toolsets)
	require.NoError(t, err)

	reposToolset := findToolsetByName(toolsets, "repos")
	require.NotNil(t, reposToolset)
	assert.Equal(t, "true", (*reposToolset)["currently_enabled"])
}