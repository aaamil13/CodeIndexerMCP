package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

func setupTestIndexer(t *testing.T) (*Indexer, string) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")
	os.MkdirAll(projectPath, 0755)

	// dbPath is not used, remove it
	// dbPath := filepath.Join(tmpDir, "test.db")

	indexer, err := NewIndexer(projectPath, nil)
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}

	if err := indexer.Initialize(); err != nil {
		t.Fatalf("Failed to initialize indexer: %v", err)
	}

	return indexer, projectPath
}

func TestIndexer_IndexGoFile(t *testing.T) {
	indexer, projectPath := setupTestIndexer(t)
	defer indexer.Close()

	// Create a Go file
	goFile := filepath.Join(projectPath, "test.go")
	code := `package main

import "fmt"

// Greet prints a greeting
func Greet(name string) {
	fmt.Println("Hello", name)
}

func main() {
	Greet("World")
}
`
	if err := os.WriteFile(goFile, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Index the project
	if err := indexer.IndexAll(); err != nil {
		t.Fatalf("IndexAll failed: %v", err)
	}

	// Search for symbols
	symbols, err := indexer.SearchSymbols(model.SearchOptions{
		Query: "Greet",
	})
	if err != nil {
		t.Fatalf("SearchSymbols failed: %v", err)
	}

	if len(symbols) == 0 {
		t.Error("Expected to find Greet function")
	}

	// Check symbol details
	if symbols[0].Name != "Greet" {
		t.Errorf("Expected symbol name Greet, got %s", symbols[0].Name)
	}
	if symbols[0].Kind != model.SymbolKindFunction {
		t.Errorf("Expected function type, got %s", symbols[0].Kind)
	}
}

func TestIndexer_IndexPythonFile(t *testing.T) {
	indexer, projectPath := setupTestIndexer(t)
	defer indexer.Close()

	// Create a Python file
	pyFile := filepath.Join(projectPath, "test.py")
	code := `
class Calculator:
    """A simple calculator."""

    def add(self, a, b):
        """Add two numbers."""
        return a + b

    def subtract(self, a, b):
        """Subtract two numbers."""
        return a - b
`
	if err := os.WriteFile(pyFile, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Index the project
	if err := indexer.IndexAll(); err != nil {
		t.Fatalf("IndexAll failed: %v", err)
	}

	// Search for class
	symbols, err := indexer.SearchSymbols(model.SearchOptions{
		Query: "Calculator",
	})
	if err != nil {
		t.Fatalf("SearchSymbols failed: %v", err)
	}

	if len(symbols) == 0 {
		t.Error("Expected to find Calculator class")
	}

	// Search for methods
	methodType := model.SymbolKindMethod
	symbols, err = indexer.SearchSymbols(model.SearchOptions{
		Query: "add",
		Kind:  string(methodType),
	})
	if err != nil {
		t.Fatalf("SearchSymbols failed: %v", err)
	}

	if len(symbols) == 0 {
		t.Error("Expected to find add method")
	}
}

func TestIndexer_MultipleFiles(t *testing.T) {
	indexer, projectPath := setupTestIndexer(t)
	defer indexer.Close()

	// Create multiple Go files
	files := map[string]string{
		"math.go": `package main

func Add(a, b int) int {
	return a + b
}
`,
		"string.go": `package main

func Concat(a, b string) string {
	return a + b
}
`,
		"types.go": `package main

type Result struct {
	Value int
	Error error
}
`,
	}

	for name, code := range files {
		filePath := filepath.Join(projectPath, name)
		if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", name, err)
		}
	}

	// Index the project
	if err := indexer.IndexAll(); err != nil {
		t.Fatalf("IndexAll failed: %v", err)
	}

	// Get project overview
	overview, err := indexer.GetProjectOverview()
	if err != nil {
		t.Fatalf("GetProjectOverview failed: %v", err)
	}

	if overview.TotalFiles != 3 {
		t.Errorf("Expected 3 files, got %d", overview.TotalFiles)
	}

	if overview.TotalSymbols < 3 {
		t.Errorf("Expected at least 3 symbols, got %d", overview.TotalSymbols)
	}
}

func TestIndexer_GetFileStructure(t *testing.T) {
	indexer, projectPath := setupTestIndexer(t)
	defer indexer.Close()

	// Create a Go file
	goFile := filepath.Join(projectPath, "api.go")
	code := `package main

import "net/http"

type Server struct {
	port int
}

func (s *Server) Start() error {
	return nil
}

func NewServer(port int) *Server {
	return &Server{port: port}
}
`
	if err := os.WriteFile(goFile, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Index the project
	if err := indexer.IndexAll(); err != nil {
		t.Fatalf("IndexAll failed: %v", err)
	}

	// Get file structure
	structure, err := indexer.GetFileStructure(goFile)
	if err != nil {
		t.Fatalf("GetFileStructure failed: %v", err)
	}

	if structure.FilePath == "" { // File is not directly embedded, check FilePath
		t.Fatal("Expected file path in structure")
	}

	if len(structure.Symbols) < 3 {
		t.Errorf("Expected at least 3 symbols, got %d", len(structure.Symbols))
	}

	if len(structure.Imports) < 1 {
		t.Errorf("Expected at least 1 import, got %d", len(structure.Imports))
	}
}

func TestIndexer_SymbolDetails(t *testing.T) {
	indexer, projectPath := setupTestIndexer(t)
	defer indexer.Close()

	// Create a Go file with references
	goFile := filepath.Join(projectPath, "code.go")
	code := `package main

func Helper() string {
	return "help"
}

func Main() {
	result := Helper()
	println(result)
	Helper()
}
`
	if err := os.WriteFile(goFile, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Index the project
	if err := indexer.IndexAll(); err != nil {
		t.Fatalf("IndexAll failed: %v", err)
	}

	// Get symbol details
	details, err := indexer.GetSymbolDetails("Helper")
	if err != nil {
		t.Fatalf("GetSymbolDetails failed: %v", err)
	}

	if details.Symbol == nil {
		t.Fatal("Expected symbol in details")
	}
	if details.Symbol.Name != "Helper" {
		t.Errorf("Expected symbol name Helper, got %s", details.Symbol.Name)
	}
}

func TestIndexer_AIFeatures(t *testing.T) {
	indexer, projectPath := setupTestIndexer(t)
	defer indexer.Close()

	// Create a Go file
	goFile := filepath.Join(projectPath, "code.go")
	code := `package main

func Calculate(x, y int) int {
	if x > y {
		if x > 100 {
			return x * y
		}
		return x + y
	}
	return y - x
}
`
	if err := os.WriteFile(goFile, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Index the project
	if err := indexer.IndexAll(); err != nil {
		t.Fatalf("IndexAll failed: %v", err)
	}

	// Test GetCodeMetrics
	metrics, err := indexer.GetCodeMetrics("Calculate")
	if err == nil || err.Error() != "not implemented" {
		t.Fatalf("Expected 'not implemented' error, got %v", err)
	}
	if metrics != nil {
		t.Fatal("Expected nil metrics when not implemented")
	}

	// Test GetCodeContext
	context, err := indexer.GetCodeContext("Calculate", 5)
	if err == nil || err.Error() != "not implemented" {
		t.Fatalf("Expected 'not implemented' error, got %v", err)
	}
	if context != nil {
		t.Fatal("Expected nil context when not implemented")
	}

	// Test AnalyzeChangeImpact
	impact, err := indexer.AnalyzeChangeImpact("Calculate")
	if err == nil || err.Error() != "not implemented" {
		t.Fatalf("Expected 'not implemented' error, got %v", err)
	}
	if impact != nil {
		t.Fatal("Expected nil impact when not implemented")
	}
}

func TestIndexer_SimulateChange(t *testing.T) {
	indexer, projectPath := setupTestIndexer(t)
	defer indexer.Close()

	// Create a Go file
	goFile := filepath.Join(projectPath, "code.go")
	code := `package main

func OldName() string {
	return "test"
}

func Caller() {
	result := OldName()
	println(result)
}
`
	if err := os.WriteFile(goFile, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Index the project
	if err := indexer.IndexAll(); err != nil {
		t.Fatalf("IndexAll failed: %v", err)
	}

	// Simulate rename
	result, err := indexer.SimulateSymbolChange("OldName", model.ChangeTypeRename, "NewName")
	if err == nil || err.Error() != "not implemented" {
		t.Fatalf("Expected 'not implemented' error, got %v", err)
	}
	if result != nil {
		t.Fatal("Expected nil result from SimulateSymbolChange when not implemented")
	}
}

func TestIndexer_DependencyGraph(t *testing.T) {
	indexer, projectPath := setupTestIndexer(t)
	defer indexer.Close()

	// Create a Go file with dependencies
	goFile := filepath.Join(projectPath, "code.go")
	code := `package main

func Helper1() string {
	return "help1"
}

func Helper2() string {
	return "help2"
}

func Main() {
	a := Helper1()
	b := Helper2()
	println(a, b)
}
`
	if err := os.WriteFile(goFile, []byte(code), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Index the project
	if err := indexer.IndexAll(); err != nil {
		t.Fatalf("IndexAll failed: %v", err)
	}

	// Build dependency graph
	graph, err := indexer.BuildDependencyGraph("Main", 2)
	if err == nil || err.Error() != "not implemented" {
		t.Fatalf("Expected 'not implemented' error, got %v", err)
	}
	if graph != nil {
		t.Fatal("Expected nil graph when not implemented")
	}
}

func TestIndexer_UnsupportedLanguage(t *testing.T) {
	indexer, projectPath := setupTestIndexer(t)
	defer indexer.Close()

	// Create an unsupported file
	unsupportedFile := filepath.Join(projectPath, "test.xyz")
	if err := os.WriteFile(unsupportedFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Index should skip unsupported files without error
	if err := indexer.IndexAll(); err != nil {
		t.Fatalf("IndexAll should not fail on unsupported files: %v", err)
	}

	// Should have 0 files indexed
	overview, _ := indexer.GetProjectOverview()
	if overview.TotalFiles != 0 {
		t.Errorf("Expected 0 files for unsupported language, got %d", overview.TotalFiles)
	}
}

func TestIndexer_Close(t *testing.T) {
	indexer, _ := setupTestIndexer(t)

	// Close indexer
	if err := indexer.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Operations after close should fail
	err := indexer.IndexAll()
	if err == nil {
		t.Error("Expected error when using indexer after Close")
	}
}
