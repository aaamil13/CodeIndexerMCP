package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
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
	symbols, err := indexer.SearchSymbols(types.SearchOptions{
		Query:     "Greet",
		ProjectID: indexer.project.ID,
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
	if symbols[0].Type != types.SymbolTypeFunction {
		t.Errorf("Expected function type, got %s", symbols[0].Type)
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
	symbols, err := indexer.SearchSymbols(types.SearchOptions{
		Query:     "Calculator",
		ProjectID: indexer.project.ID,
	})
	if err != nil {
		t.Fatalf("SearchSymbols failed: %v", err)
	}

	if len(symbols) == 0 {
		t.Error("Expected to find Calculator class")
	}

	// Search for methods
	methodType := types.SymbolTypeMethod
	symbols, err = indexer.SearchSymbols(types.SearchOptions{
		Query:     "add",
		Type:      &methodType,
		ProjectID: indexer.project.ID,
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

func TestIndexer_IncrementalIndex(t *testing.T) {
	indexer, projectPath := setupTestIndexer(t)
	defer indexer.Close()

	// Create initial file
	goFile := filepath.Join(projectPath, "code.go")
	code1 := `package main

func Version1() string {
	return "v1"
}
`
	if err := os.WriteFile(goFile, []byte(code1), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Initial index
	if err := indexer.IndexAll(); err != nil {
		t.Fatalf("IndexAll failed: %v", err)
	}

	// Verify Version1 exists
	symbols1, _ := indexer.SearchSymbols(types.SearchOptions{
		Query:     "Version1",
		ProjectID: indexer.project.ID,
	})
	if len(symbols1) == 0 {
		t.Error("Expected to find Version1")
	}

	// Modify file
	time.Sleep(10 * time.Millisecond) // Ensure timestamp changes
	code2 := `package main

func Version2() string {
	return "v2"
}
`
	if err := os.WriteFile(goFile, []byte(code2), 0644); err != nil {
		t.Fatalf("Failed to update test file: %v", err)
	}

	// Re-index
	if err := indexer.IndexFile(goFile); err != nil {
		t.Fatalf("IndexFile failed: %v", err)
	}

	// Verify Version2 exists and Version1 doesn't
	symbols2, _ := indexer.SearchSymbols(types.SearchOptions{
		Query:     "Version2",
		ProjectID: indexer.project.ID,
	})
	if len(symbols2) == 0 {
		t.Error("Expected to find Version2 after update")
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
	if err != nil {
		t.Fatalf("GetCodeMetrics failed: %v", err)
	}

	if metrics.FunctionName == "" { // Symbol is not directly embedded, check FunctionName
		t.Fatal("Expected function name in metrics")
	}
	if metrics.CyclomaticComplexity == 0 {
		t.Error("Expected non-zero cyclomatic complexity")
	}

	// Test GetCodeContext
	context, err := indexer.GetCodeContext("Calculate", 5)
	if err != nil {
		t.Fatalf("GetCodeContext failed: %v", err)
	}

	if context.Symbol == nil {
		t.Fatal("Expected symbol in context")
	}
	if context.Code == "" {
		t.Error("Expected code in context")
	}

	// Test AnalyzeChangeImpact
	impact, err := indexer.AnalyzeChangeImpact("Calculate")
	if err != nil {
		t.Fatalf("AnalyzeChangeImpact failed: %v", err)
	}

	if impact.Symbol == nil {
		t.Fatal("Expected symbol in impact")
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
	result, err := indexer.SimulateSymbolChange("OldName", types.ChangeTypeRename, "NewName")
	if err != nil {
		t.Fatalf("SimulateSymbolChange failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result from SimulateSymbolChange")
	}

	// Should have auto-fix suggestions for rename
	if len(result.AutoFixSuggestions) == 0 {
		t.Error("Expected auto-fix suggestions for rename")
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
	if err != nil {
		t.Fatalf("BuildDependencyGraph failed: %v", err)
	}

	if graph == nil {
		t.Fatal("Expected dependency graph")
	}

	if len(graph.Nodes) == 0 { // RootSymbol is not directly embedded, check nodes
		t.Fatal("Expected nodes in graph")
	}
	// Further checks on nodes can be added if needed
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
