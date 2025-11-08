package golang

import (
	"testing"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

func TestParseFunction(t *testing.T) {
	code := `package main

import "fmt"

// Add adds two numbers
func Add(a, b int) int {
	return a + b
}
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.go")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Symbols) != 1 {
		t.Fatalf("Expected 1 symbol, got %d", len(result.Symbols))
	}

	sym := result.Symbols[0]
	if sym.Name != "Add" {
		t.Errorf("Expected name 'Add', got '%s'", sym.Name)
	}
	if sym.Kind != model.SymbolKindFunction {
		t.Errorf("Expected type function, got %s", sym.Kind)
	}
	if sym.Signature != "func Add(a, b int) int" {
		t.Errorf("Expected signature 'func Add(a, b int) int', got '%s'", sym.Signature)
	}
	if sym.Documentation != "Add adds two numbers" {
		t.Errorf("Expected documentation 'Add adds two numbers', got '%s'", sym.Documentation)
	}
	if sym.Name == "publicFunc" && sym.Visibility != model.VisibilityPublic {
		t.Error("Expected publicFunc to be exported")
	}
	if sym.Name == "privateFunc" && sym.Visibility == model.VisibilityPublic {
		t.Error("Expected privateFunc to not be exported")
	}
}

func TestParseStruct(t *testing.T) {
	code := `package main

// User represents a user
type User struct {
	ID   int
	Name string
}
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.go")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Symbols) != 1 {
		t.Fatalf("Expected 1 symbol, got %d", len(result.Symbols))
	}

	sym := result.Symbols[0]
	if sym.Name != "User" {
		t.Errorf("Expected name 'User', got '%s'", sym.Name)
	}
	if sym.Kind != model.SymbolKindStruct {
		t.Errorf("Expected type struct, got %s", sym.Kind)
	}
	if sym.Documentation != "User represents a user" {
		t.Errorf("Expected documentation 'User represents a user', got '%s'", sym.Documentation)
	}
}

func TestParseInterface(t *testing.T) {
	code := `package main

// Writer defines write operations
type Writer interface {
	Write(data []byte) error
	Close() error
}
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.go")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Symbols) != 1 {
		t.Fatalf("Expected 1 symbol, got %d", len(result.Symbols))
	}

	sym := result.Symbols[0]
	if sym.Name != "Writer" {
		t.Errorf("Expected name 'Writer', got '%s'", sym.Name)
	}
	if sym.Kind != model.SymbolKindInterface {
		t.Errorf("Expected type interface, got %s", sym.Kind)
	}
}

func TestParseMethod(t *testing.T) {
	code := `package main

type Calculator struct{}

// Calculate performs calculation
func (c *Calculator) Calculate(x, y int) int {
	return x + y
}
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.go")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have 2 symbols: Calculator struct and Calculate method
	if len(result.Symbols) != 2 {
		t.Fatalf("Expected 2 symbols, got %d", len(result.Symbols))
	}

	// Find the method
	var method *model.Symbol
	for _, sym := range result.Symbols {
		if sym.Kind == model.SymbolKindMethod {
			method = sym
			break
		}
	}

	if method == nil {
		t.Fatal("Method symbol not found")
	}

	if method.Name != "Calculate" {
		t.Errorf("Expected name 'Calculate', got '%s'", method.Name)
	}
	if method.Signature != "func (c *Calculator) Calculate(x, y int) int" {
		t.Errorf("Unexpected signature: %s", method.Signature)
	}
}

func TestParseImports(t *testing.T) {
	code := `package main

import (
	"fmt"
	"os"
	custom "github.com/user/custom"
)
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.go")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Imports) != 3 {
		t.Fatalf("Expected 3 imports, got %d", len(result.Imports))
	}

	// Check for specific imports
	imports := make(map[string]bool)
	for _, imp := range result.Imports {
		imports[imp.Path] = true
	}

	if !imports["fmt"] {
		t.Error("Expected import 'fmt'")
	}
	if !imports["os"] {
		t.Error("Expected import 'os'")
	}
	if !imports["github.com/user/custom"] {
		t.Error("Expected import 'github.com/user/custom'")
	}
}

func TestParseConstants(t *testing.T) {
	code := `package main

const (
	MaxSize = 1024
	MinSize = 10
)
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.go")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Symbols) != 2 {
		t.Fatalf("Expected 2 symbols, got %d", len(result.Symbols))
	}

	for _, sym := range result.Symbols {
		if sym.Kind != model.SymbolKindConstant {
			t.Errorf("Expected type constant, got %s", sym.Kind)
		}
	}
}

func TestParseVariables(t *testing.T) {
	code := `package main

var (
	Config map[string]string
	Logger interface{}
)
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.go")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Symbols) != 2 {
		t.Fatalf("Expected 2 symbols, got %d", len(result.Symbols))
	}

	for _, sym := range result.Symbols {
		if sym.Kind != model.SymbolKindVariable {
			t.Errorf("Expected type variable, got %s", sym.Kind)
		}
	}
}

func TestParsePrivateSymbols(t *testing.T) {
	code := `package main

func publicFunc() {}

func privateFunc() {}
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.go")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Symbols) != 2 {
		t.Fatalf("Expected 2 symbols, got %d", len(result.Symbols))
	}

	for _, sym := range result.Symbols {
			if sym.Name == "publicFunc" && sym.Visibility != model.VisibilityPublic {
				t.Error("Expected publicFunc to be exported")
			}
			if sym.Name == "privateFunc" && sym.Visibility == model.VisibilityPublic {
				t.Error("Expected privateFunc to not be exported")
			}	}
}

func TestParseMultilineDocumentation(t *testing.T) {
	code := `package main

// ProcessData processes the input data
// and returns the result.
// It handles multiple formats.
func ProcessData(data string) string {
	return data
}
`
	parser := NewParser()
	result, err := parser.Parse([]byte(code), "test.go")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Symbols) != 1 {
		t.Fatalf("Expected 1 symbol, got %d", len(result.Symbols))
	}

	sym := result.Symbols[0]
	expectedDoc := "ProcessData processes the input data\nand returns the result.\nIt handles multiple formats."
	if sym.Documentation != expectedDoc {
		t.Errorf("Expected documentation:\n%s\nGot:\n%s", expectedDoc, sym.Documentation)
	}
}

func TestParseInvalidSyntax(t *testing.T) {
	code := `package main

func invalid syntax {
`
	parser := NewParser()
	_, err := parser.Parse([]byte(code), "test.go")
	if err == nil {
		t.Error("Expected error for invalid syntax, got nil")
	}
}
