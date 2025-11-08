package python

import (
	"testing"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

func TestParseFunction(t *testing.T) {
	code := `
def add(a, b):
    """Add two numbers."""
    return a + b
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.py")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Symbols) != 1 {
		t.Fatalf("Expected 1 symbol, got %d", len(result.Symbols))
	}

	sym := result.Symbols[0]
	if sym.Name != "add" {
		t.Errorf("Expected name 'add', got '%s'", sym.Name)
	}
	if sym.Type != model.SymbolTypeFunction {
		t.Errorf("Expected type function, got %s", sym.Type)
	}
	if sym.Signature != "def add(a, b):" {
		t.Errorf("Expected signature 'def add(a, b):', got '%s'", sym.Signature)
	}
	if sym.Documentation != "Add two numbers." {
		t.Errorf("Expected documentation 'Add two numbers.', got '%s'", sym.Documentation)
	}
	if sym.Visibility != model.VisibilityPublic {
		t.Error("Expected public visibility")
	}
}

func TestParseClass(t *testing.T) {
	code := `
class User:
    """Represents a user."""

    def __init__(self, name):
        self.name = name

    def greet(self):
        return f"Hello, {self.name}"
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.py")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have class + 2 methods
	if len(result.Symbols) != 3 {
		t.Fatalf("Expected 3 symbols, got %d", len(result.Symbols))
	}

	// Find the class
	var class *model.Symbol
	for _, sym := range result.Symbols {
		if sym.Type == model.SymbolTypeClass {
			class = sym
			break
		}
	}

	if class == nil {
		t.Fatal("Class symbol not found")
	}

	if class.Name != "User" {
		t.Errorf("Expected name 'User', got '%s'", class.Name)
	}
	if class.Documentation != "Represents a user." {
		t.Errorf("Expected documentation 'Represents a user.', got '%s'", class.Documentation)
	}
}

func TestParseMethod(t *testing.T) {
	code := `
class Calculator:
    def calculate(self, x, y):
        """Perform calculation."""
        return x + y
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.py")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have class + method
	if len(result.Symbols) != 2 {
		t.Fatalf("Expected 2 symbols, got %d", len(result.Symbols))
	}

	// Find the method
	var method *model.Symbol
	for _, sym := range result.Symbols {
		if sym.Type == model.SymbolTypeMethod {
			method = sym
			break
		}
	}

	if method == nil {
		t.Fatal("Method symbol not found")
	}

	if method.Name != "calculate" {
		t.Errorf("Expected name 'calculate', got '%s'", method.Name)
	}
	if method.Documentation != "Perform calculation." {
		t.Errorf("Expected documentation 'Perform calculation.', got '%s'", method.Documentation)
	}
}

func TestParseImports(t *testing.T) {
	code := `
import os
import sys
from typing import List, Dict
from .local import helper
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.py")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Imports) != 4 {
		t.Fatalf("Expected 4 imports, got %d", len(result.Imports))
	}

	imports := make(map[string]bool)
	for _, imp := range result.Imports {
		imports[imp.Source] = true
	}

	if !imports["os"] {
		t.Error("Expected import 'os'")
	}
	if !imports["sys"] {
		t.Error("Expected import 'sys'")
	}
	if !imports["typing"] {
		t.Error("Expected import 'typing'")
	}
}

func TestParseAsyncFunction(t *testing.T) {
	code := `
async def fetch_data(url):
    """Fetch data asynchronously."""
    return await request(url)
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.py")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Symbols) != 1 {
		t.Fatalf("Expected 1 symbol, got %d", len(result.Symbols))
	}

	sym := result.Symbols[0]
	if sym.Name != "fetch_data" {
		t.Errorf("Expected name 'fetch_data', got '%s'", sym.Name)
	}
	if sym.Type != model.SymbolTypeFunction {
		t.Errorf("Expected type function, got %s", sym.Type)
	}
	if sym.Signature != "async def fetch_data(url):" {
		t.Errorf("Unexpected signature: %s", sym.Signature)
	}
}

func TestParseDecorators(t *testing.T) {
	code := `
@staticmethod
def static_method():
    """A static method."""
    pass

@property
def prop(self):
    """A property."""
    return self._value
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.py")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Symbols) != 2 {
		t.Fatalf("Expected 2 symbols, got %d", len(result.Symbols))
	}

	for _, sym := range result.Symbols {
		if sym.Type != model.SymbolTypeFunction {
			t.Errorf("Expected type function, got %s", sym.Type)
		}
	}
}

func TestParsePrivateMethods(t *testing.T) {
	code := `
class MyClass:
    def public_method(self):
        pass

    def _private_method(self):
        pass

    def __internal_method(self):
        pass
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.py")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have class + 3 methods
	if len(result.Symbols) != 4 {
		t.Fatalf("Expected 4 symbols, got %d", len(result.Symbols))
	}

	for _, sym := range result.Symbols {
		if sym.Type != model.SymbolTypeMethod && sym.Type != model.SymbolTypeClass {
			continue
		}

		if sym.Name == "public_method" {
			if sym.Visibility != model.VisibilityPublic {
				t.Error("Expected public_method to be public")
			}
		}
		if sym.Name == "_private_method" {
			if sym.Visibility != model.VisibilityPrivate {
				t.Error("Expected _private_method to be private")
			}
		}
		if sym.Name == "__internal_method" {
			if sym.Visibility != model.VisibilityInternal {
				t.Error("Expected __internal_method to be internal")
			}
		}
	}
}

func TestParseVariables(t *testing.T) {
	code := `
MAX_SIZE = 1024
min_size = 10
_private_var = "secret"
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.py")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Symbols) != 3 {
		t.Fatalf("Expected 3 symbols, got %d", len(result.Symbols))
	}

	for _, sym := range result.Symbols {
		if sym.Type != model.SymbolTypeVariable {
			t.Errorf("Expected type variable, got %s", sym.Type)
		}

		if sym.Name == "MAX_SIZE" && sym.Visibility != model.VisibilityPublic {
			t.Error("Expected MAX_SIZE to be public")
		}
		if sym.Name == "_private_var" && sym.Visibility != model.VisibilityPrivate {
			t.Error("Expected _private_var to be private")
		}
	}
}

func TestParseMultilineDocstring(t *testing.T) {
	code := `
def process_data(data):
    """
    Process the input data.

    This function handles multiple formats
    and returns the processed result.
    """
    return data
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.py")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Symbols) != 1 {
		t.Fatalf("Expected 1 symbol, got %d", len(result.Symbols))
	}

	sym := result.Symbols[0]
	if sym.Documentation == "" {
		t.Error("Expected documentation to be present")
	}
	if len(sym.Documentation) < 50 {
		t.Errorf("Expected longer documentation, got: %s", sym.Documentation)
	}
}

func TestParseClassWithConstructor(t *testing.T) {
	code := `
class Person:
    """A person class."""

    def __init__(self, name, age):
        """Initialize a person."""
        self.name = name
        self.age = age
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.py")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have class + constructor
	if len(result.Symbols) != 2 {
		t.Fatalf("Expected 2 symbols, got %d", len(result.Symbols))
	}

	// Find the constructor
	var constructor *model.Symbol
	for _, sym := range result.Symbols {
		if sym.Name == "__init__" {
			constructor = sym
			break
		}
	}

	if constructor == nil {
		t.Fatal("Constructor not found")
	}

	if constructor.Type != model.SymbolTypeMethod {
		t.Errorf("Expected constructor to be a method, got %s", constructor.Type)
	}
}

func TestParseEmptyFile(t *testing.T) {
	code := `# Just a comment`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.py")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Symbols) != 0 {
		t.Errorf("Expected 0 symbols for empty file, got %d", len(result.Symbols))
	}
}
