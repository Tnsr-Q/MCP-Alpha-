package repository

import (
	"fmt"
	"regexp"
	"strings"
)

// extractPythonSignatures extracts function and class signatures from Python source code
func extractPythonSignatures(code string) ([]FunctionSignature, error) {
	var signatures []FunctionSignature

	lines := strings.Split(code, "\n")
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		// Match function definitions
		if funcMatch := regexp.MustCompile(`^def\s+(\w+)\s*\((.*?)\)\s*(?:->.*?)?:`).FindStringSubmatch(line); funcMatch != nil {
			name := funcMatch[1]
			params := funcMatch[2]
			
			// Skip private functions (starting with _)
			if strings.HasPrefix(name, "_") {
				continue
			}
			
			// Extract docstring
			docstring := extractPythonDocstring(lines, i+1)
			
			// Parse parameters
			parameters, required := parsePythonParameters(params)
			
			signatures = append(signatures, FunctionSignature{
				Name:        name,
				Type:        "function",
				Signature:   line,
				Description: docstring,
				Parameters:  parameters,
				Required:    required,
			})
		}
		
		// Match class definitions
		if classMatch := regexp.MustCompile(`^class\s+(\w+)(?:\(.*?\))?:`).FindStringSubmatch(line); classMatch != nil {
			name := classMatch[1]
			
			// Skip private classes (starting with _)
			if strings.HasPrefix(name, "_") {
				continue
			}
			
			// Extract docstring
			docstring := extractPythonDocstring(lines, i+1)
			
			signatures = append(signatures, FunctionSignature{
				Name:        name,
				Type:        "class",
				Signature:   line,
				Description: docstring,
			})
		}
	}
	
	return signatures, nil
}

// extractJavaScriptSignatures extracts function signatures from JavaScript/TypeScript source code
func extractJavaScriptSignatures(code string) ([]FunctionSignature, error) {
	var signatures []FunctionSignature

	lines := strings.Split(code, "\n")
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		// Match function declarations: function name(params) or function name(params): returnType
		if funcMatch := regexp.MustCompile(`^(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*\((.*?)\)(?:\s*:\s*.*?)?`).FindStringSubmatch(line); funcMatch != nil {
			name := funcMatch[1]
			params := funcMatch[2]
			
			// Skip private functions (starting with _)
			if strings.HasPrefix(name, "_") {
				continue
			}
			
			// Extract JSDoc comment
			jsdoc := extractJSDocComment(lines, i)
			
			// Parse parameters
			parameters, required := parseJavaScriptParameters(params)
			
			signatures = append(signatures, FunctionSignature{
				Name:        name,
				Type:        "function",
				Signature:   line,
				Description: jsdoc,
				Parameters:  parameters,
				Required:    required,
			})
		}
		
		// Match arrow function exports: export const name = (params) => or const name = (params): returnType =>
		if arrowMatch := regexp.MustCompile(`^(?:export\s+)?const\s+(\w+)\s*=\s*(?:async\s+)?\((.*?)\)(?:\s*:\s*.*?)?\s*=>`).FindStringSubmatch(line); arrowMatch != nil {
			name := arrowMatch[1]
			params := arrowMatch[2]
			
			// Skip private functions (starting with _)
			if strings.HasPrefix(name, "_") {
				continue
			}
			
			// Extract JSDoc comment
			jsdoc := extractJSDocComment(lines, i)
			
			// Parse parameters
			parameters, required := parseJavaScriptParameters(params)
			
			signatures = append(signatures, FunctionSignature{
				Name:        name,
				Type:        "function",
				Signature:   line,
				Description: jsdoc,
				Parameters:  parameters,
				Required:    required,
			})
		}
		
		// Match class definitions
		if classMatch := regexp.MustCompile(`^(?:export\s+)?(?:abstract\s+)?class\s+(\w+)(?:\s+extends\s+\w+)?(?:\s+implements\s+.*?)?`).FindStringSubmatch(line); classMatch != nil {
			name := classMatch[1]
			
			// Skip private classes (starting with _)
			if strings.HasPrefix(name, "_") {
				continue
			}
			
			// Extract JSDoc comment
			jsdoc := extractJSDocComment(lines, i)
			
			signatures = append(signatures, FunctionSignature{
				Name:        name,
				Type:        "class",
				Signature:   line,
				Description: jsdoc,
			})
		}
	}
	
	return signatures, nil
}

// extractPythonDocstring extracts docstring from Python code starting at the given line
func extractPythonDocstring(lines []string, startLine int) string {
	if startLine >= len(lines) {
		return ""
	}
	
	line := strings.TrimSpace(lines[startLine])
	
	// Check for triple-quoted docstring
	if strings.HasPrefix(line, `"""`) || strings.HasPrefix(line, "'''") {
		quote := line[:3]
		
		// Single line docstring
		if len(line) > 6 && strings.HasSuffix(line, quote) {
			return strings.Trim(line[3:len(line)-3], " \t")
		}
		
		// Multi-line docstring
		var docLines []string
		if len(line) > 3 {
			docLines = append(docLines, line[3:])
		}
		
		for i := startLine + 1; i < len(lines); i++ {
			line := lines[i]
			if strings.Contains(line, quote) {
				// Found end quote
				endIdx := strings.Index(line, quote)
				if endIdx > 0 {
					docLines = append(docLines, line[:endIdx])
				}
				break
			}
			docLines = append(docLines, line)
		}
		
		// Clean up and join
		docstring := strings.Join(docLines, "\n")
		return strings.TrimSpace(docstring)
	}
	
	return ""
}

// extractJSDocComment extracts JSDoc comment from JavaScript/TypeScript code before the given line
func extractJSDocComment(lines []string, lineNum int) string {
	var commentLines []string
	
	// Look backwards for JSDoc comment
	for i := lineNum - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		
		if line == "" {
			continue
		}
		
		if strings.HasPrefix(line, "*/") {
			// Found end of comment block, continue looking backwards
			continue
		} else if strings.HasPrefix(line, "*") {
			// Comment line
			content := strings.TrimSpace(line[1:])
			if content != "" {
				commentLines = append([]string{content}, commentLines...)
			}
		} else if strings.HasPrefix(line, "/**") {
			// Found start of JSDoc comment
			content := strings.TrimSpace(line[3:])
			if content != "" && !strings.HasSuffix(content, "*/") {
				commentLines = append([]string{content}, commentLines...)
			}
			break
		} else {
			// Not a comment line, stop looking
			break
		}
	}
	
	return strings.Join(commentLines, " ")
}

// parsePythonParameters parses Python function parameters and returns parameter schema and required list
func parsePythonParameters(params string) (map[string]interface{}, []string) {
	if params == "" {
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}, []string{}
	}
	
	properties := make(map[string]interface{})
	var required []string
	
	// Split parameters by comma (simple approach)
	paramList := strings.Split(params, ",")
	
	for _, param := range paramList {
		param = strings.TrimSpace(param)
		if param == "" || param == "self" {
			continue
		}
		
		// Skip *args and **kwargs
		if strings.HasPrefix(param, "*") {
			continue
		}
		
		// Remove type hints for parsing
		paramName := param
		if idx := strings.Index(param, ":"); idx >= 0 {
			paramName = strings.TrimSpace(param[:idx])
		}
		
		// Check if parameter has default value
		hasDefault := strings.Contains(param, "=")
		if !hasDefault {
			// Only add to required if no default value
			required = append(required, paramName)
		}
		
		if paramName != "" {
			properties[paramName] = map[string]interface{}{
				"type":        "string",
				"description": fmt.Sprintf("Parameter %s", paramName),
			}
		}
	}
	
	return map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}, required
}

// parseJavaScriptParameters parses JavaScript/TypeScript function parameters
func parseJavaScriptParameters(params string) (map[string]interface{}, []string) {
	if params == "" {
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}, []string{}
	}
	
	properties := make(map[string]interface{})
	var required []string
	
	// Split parameters by comma (simple approach)
	paramList := strings.Split(params, ",")
	
	for _, param := range paramList {
		param = strings.TrimSpace(param)
		if param == "" {
			continue
		}
		
		// Skip rest/spread parameters
		if strings.HasPrefix(param, "...") {
			continue
		}
		
		// Remove type annotations for parsing
		paramName := param
		if idx := strings.Index(param, ":"); idx >= 0 {
			paramName = strings.TrimSpace(param[:idx])
		}
		
		// Handle optional parameters (ending with ?)
		isOptional := strings.HasSuffix(paramName, "?")
		if isOptional {
			paramName = strings.TrimSuffix(paramName, "?")
		}
		
		// Check if parameter has default value
		hasDefault := strings.Contains(param, "=")
		
		// Parameter is required if it's not optional and has no default value
		if !isOptional && !hasDefault {
			required = append(required, paramName)
		}
		
		if paramName != "" {
			properties[paramName] = map[string]interface{}{
				"type":        "string",
				"description": fmt.Sprintf("Parameter %s", paramName),
			}
		}
	}
	
	return map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}, required
}