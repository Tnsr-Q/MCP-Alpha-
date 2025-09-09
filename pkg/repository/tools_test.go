package repository

import (
	"testing"
)

func TestExtractPythonSignatures(t *testing.T) {
	code := `
def hello_world(name: str, age: int = 25) -> str:
    """
    Greets a person with their name and age.
    
    Args:
        name: The person's name
        age: The person's age (default: 25)
    
    Returns:
        A greeting string
    """
    return f"Hello {name}, you are {age} years old!"

class Person:
    """A simple person class."""
    
    def __init__(self, name: str):
        self.name = name
        
    def _private_method(self):
        """This should be skipped."""
        pass

def _private_function():
    """This should also be skipped."""
    pass
`

	signatures, err := extractPythonSignatures(code)
	if err != nil {
		t.Fatalf("Failed to extract Python signatures: %v", err)
	}

	if len(signatures) != 2 {
		t.Errorf("Expected 2 signatures, got %d", len(signatures))
	}

	// Check function signature
	found := false
	for _, sig := range signatures {
		if sig.Name == "hello_world" && sig.Type == "function" {
			found = true
			if sig.Description == "" {
				t.Error("Expected function to have description")
			}
			if len(sig.Required) == 0 {
				t.Error("Expected function to have required parameters")
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find hello_world function")
	}

	// Check class signature
	found = false
	for _, sig := range signatures {
		if sig.Name == "Person" && sig.Type == "class" {
			found = true
			if sig.Description == "" {
				t.Error("Expected class to have description")
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find Person class")
	}
}

func TestExtractJavaScriptSignatures(t *testing.T) {
	code := `
/**
 * Calculates the sum of two numbers
 * @param {number} a - First number
 * @param {number} b - Second number
 * @returns {number} The sum
 */
function add(a, b) {
    return a + b;
}

/**
 * A utility class for mathematical operations
 */
class Calculator {
    constructor() {
        this.value = 0;
    }
    
    _privateMethod() {
        // This should be skipped
    }
}

const multiply = (x, y) => {
    return x * y;
};

export const divide = (a, b) => a / b;

function _privateFunction() {
    // This should be skipped
}
`

	signatures, err := extractJavaScriptSignatures(code)
	if err != nil {
		t.Fatalf("Failed to extract JavaScript signatures: %v", err)
	}

	if len(signatures) < 3 {
		t.Errorf("Expected at least 3 signatures, got %d", len(signatures))
	}

	// Check for the add function
	found := false
	for _, sig := range signatures {
		if sig.Name == "add" && sig.Type == "function" {
			found = true
			if sig.Description == "" {
				t.Error("Expected add function to have description")
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find add function")
	}

	// Check for Calculator class
	found = false
	for _, sig := range signatures {
		if sig.Name == "Calculator" && sig.Type == "class" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find Calculator class")
	}
}

func TestParsePythonParameters(t *testing.T) {
	params := "name: str, age: int = 25, *args, **kwargs"
	
	parameters, required := parsePythonParameters(params)
	
	if len(required) != 1 {
		t.Errorf("Expected 1 required parameter, got %d", len(required))
	}
	
	if required[0] != "name" {
		t.Errorf("Expected 'name' to be required, got %s", required[0])
	}
	
	props, ok := parameters["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}
	
	if len(props) < 2 {
		t.Errorf("Expected at least 2 properties, got %d", len(props))
	}
	
	if _, ok := props["name"]; !ok {
		t.Error("Expected 'name' property to exist")
	}
	
	if _, ok := props["age"]; !ok {
		t.Error("Expected 'age' property to exist")
	}
}

func TestParseJavaScriptParameters(t *testing.T) {
	params := "a, b = 10, ...rest"
	
	parameters, required := parseJavaScriptParameters(params)
	
	if len(required) != 1 {
		t.Errorf("Expected 1 required parameter, got %d", len(required))
	}
	
	if required[0] != "a" {
		t.Errorf("Expected 'a' to be required, got %s", required[0])
	}
	
	props, ok := parameters["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}
	
	if len(props) < 2 {
		t.Errorf("Expected at least 2 properties, got %d", len(props))
	}
}