package repository

// FunctionSignature represents a function or class extracted from source code
type FunctionSignature struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // "function" or "class"
	Signature   string                 `json:"signature"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Required    []string               `json:"required,omitempty"`
}

// FunctionDescriptor represents a function descriptor for tool generation
type FunctionDescriptor struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Required    []string               `json:"required"`
}

// ToolDefinition represents an OpenAI-compatible tool definition
type ToolDefinition struct {
	Type     string      `json:"type"`
	Function FunctionDef `json:"function"`
}

// FunctionDef represents the function part of a tool definition
type FunctionDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}