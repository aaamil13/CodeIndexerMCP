package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aaamil13/CodeIndexerMCP/internal/core"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

func setupTestMCPServer(t *testing.T) (*Server, *core.Indexer, string) {
	tmpDir := t.TempDir()
	projectPath := filepath.Join(tmpDir, "test-project")
	os.MkdirAll(projectPath, 0755)


	indexer, err := core.NewIndexer(projectPath, nil)
	if err != nil {
		t.Fatalf("Failed to create indexer: %v", err)
	}

	if err := indexer.Initialize(); err != nil {
		t.Fatalf("Failed to initialize indexer: %v", err)
	}

	server := NewServer(indexer)
	return server, indexer, projectPath
}

func TestMCPServer_HandleInitialize(t *testing.T) {
	server, indexer, _ := setupTestMCPServer(t)
	defer indexer.Close()

	result := server.handleInitialize(nil)

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map result from handleInitialize")
	}

	if resultMap["protocolVersion"] == nil {
		t.Error("Expected protocolVersion in result")
	}

	if resultMap["serverInfo"] == nil {
		t.Error("Expected serverInfo in result")
	}
}

func TestMCPServer_HandleToolsList(t *testing.T) {
	server, indexer, _ := setupTestMCPServer(t)
	defer indexer.Close()

	result := server.handleToolsList()

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map result from handleToolsList")
	}

	tools, ok := resultMap["tools"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected tools array in result")
	}

	// Should have 19 tools (8 core + 7 AI + 4 change tracking)
	if len(tools) != 19 {
		t.Errorf("Expected 19 tools, got %d", len(tools))
	}

	// Verify some tool names
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		name, ok := tool["name"].(string)
		if ok {
			toolNames[name] = true
		}
	}

	expectedTools := []string{
		"search_symbols",
		"get_file_structure",
		"get_project_overview",
		"simulate_change",
		"build_dependency_graph",
		"get_code_context",
		"analyze_change_impact",
	}

	for _, expected := range expectedTools {
		if !toolNames[expected] {
			t.Errorf("Expected tool %s not found", expected)
		}
	}
}

func TestMCPServer_HandleSearchSymbols(t *testing.T) {
	server, indexer, projectPath := setupTestMCPServer(t)
	defer indexer.Close()

	// Create a test file
	goFile := filepath.Join(projectPath, "test.go")
	code := `package main

func TestFunction() string {
	return "test"
}
`
	os.WriteFile(goFile, []byte(code), 0644)
	indexer.IndexAll()

	// Call search_symbols tool
	params := model.SearchOptions{
		Query: "TestFunction",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := server.handleSearchSymbols(paramsJSON)
	if err != nil {
		t.Fatalf("handleSearchSymbols failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map result")
	}

	_, ok = resultMap["symbols"]
	if !ok {
		t.Error("Expected symbols in result")
	}

	count, ok := resultMap["count"].(int)
	if !ok || count == 0 {
		t.Error("Expected non-zero count")
	}
}

func TestMCPServer_HandleGetFileStructure(t *testing.T) {
	server, indexer, projectPath := setupTestMCPServer(t)
	defer indexer.Close()

	// Create a test file
	goFile := filepath.Join(projectPath, "api.go")
	code := `package main

import "fmt"

type Server struct {
	port int
}

func (s *Server) Start() {
	fmt.Println("Starting")
}
`
	os.WriteFile(goFile, []byte(code), 0644)
	indexer.IndexAll()

	// Call get_file_structure tool
	params := map[string]string{"file_path": goFile}
	paramsJSON, _ := json.Marshal(params)

	result, err := server.handleGetFileStructure(paramsJSON)
	if err != nil {
		t.Fatalf("handleGetFileStructure failed: %v", err)
	}

	structure, ok := result.(*model.ParseResult)
	if !ok {
		t.Fatal("Expected ParseResult result")
	}

	if structure.FilePath == "" {
		t.Error("Expected file path in structure")
	}

	if len(structure.Symbols) == 0 {
		t.Error("Expected symbols in structure")
	}

	if len(structure.Imports) == 0 {
		t.Error("Expected imports in structure")
	}
}

func TestMCPServer_HandleGetProjectOverview(t *testing.T) {
	server, indexer, projectPath := setupTestMCPServer(t)
	defer indexer.Close()

	// Create test files
	files := map[string]string{
		"file1.go": `package main
func Func1() {}`,
		"file2.py": `def func2():
    pass`,
	}

	for name, code := range files {
		filePath := filepath.Join(projectPath, name)
		os.WriteFile(filePath, []byte(code), 0644)
	}

	indexer.IndexAll()

	// Call get_project_overview tool
	result, err := server.handleGetProjectOverview(nil)
	if err != nil {
		t.Fatalf("handleGetProjectOverview failed: %v", err)
	}

	overview, ok := result.(*model.ProjectOverview)
	if !ok {
		t.Fatal("Expected ProjectOverview result")
	}

	if overview.TotalFiles != 2 {
		t.Errorf("Expected 2 files, got %d", overview.TotalFiles)
	}

	if len(overview.Project.LanguageStats) == 0 {
		t.Error("Expected language stats in overview")
	}
}

func TestMCPServer_HandleGetSymbolDetails(t *testing.T) {
	server, indexer, projectPath := setupTestMCPServer(t)
	defer indexer.Close()

	// Create a test file
	goFile := filepath.Join(projectPath, "code.go")
	code := `package main

func Helper() string {
	return "help"
}

func Main() {
	result := Helper()
	println(result)
}
`
	os.WriteFile(goFile, []byte(code), 0644)
	indexer.IndexAll()

	// Call get_symbol_details tool
	params := map[string]string{"symbol_name": "Helper"}
	paramsJSON, _ := json.Marshal(params)

	result, err := server.handleGetSymbolDetails(paramsJSON)
	if err != nil {
		t.Fatalf("handleGetSymbolDetails failed: %v", err)
	}

	details, ok := result.(*model.SymbolDetails)
	if !ok {
		t.Fatal("Expected SymbolDetails result")
	}

	if details.Symbol == nil {
		t.Error("Expected symbol in details")
	}

	if details.Symbol.Name != "Helper" {
		t.Errorf("Expected symbol name Helper, got %s", details.Symbol.Name)
	}
}

func TestMCPServer_HandleGetCodeContext(t *testing.T) {
	server, indexer, projectPath := setupTestMCPServer(t)
	defer indexer.Close()

	// Create a test file
	goFile := filepath.Join(projectPath, "code.go")
	code := `package main

func Process(data string) string {
	return data + " processed"
}
`
	os.WriteFile(goFile, []byte(code), 0644)
	indexer.IndexAll()

	// Call get_code_context tool
	params := map[string]interface{}{
		"symbol_name": "Process",
		"depth":       5,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := server.handleGetCodeContext(paramsJSON)
	if err != nil {
		t.Fatalf("handleGetCodeContext failed: %v", err)
	}

	context, ok := result.(*model.CodeContext)
	if !ok {
		t.Fatal("Expected CodeContext result")
	}

	if context.Symbol == nil {
		t.Error("Expected symbol in context")
	}

	if context.Code == "" {
		t.Error("Expected code in context")
	}
}

func TestMCPServer_HandleAnalyzeChangeImpact(t *testing.T) {
	server, indexer, projectPath := setupTestMCPServer(t)
	defer indexer.Close()

	// Create a test file
	goFile := filepath.Join(projectPath, "code.go")
	code := `package main

func Important() string {
	return "important"
}

func User() {
	result := Important()
	println(result)
}
`
	os.WriteFile(goFile, []byte(code), 0644)
	indexer.IndexAll()

	// Call analyze_change_impact tool
	params := map[string]string{"symbol_name": "Important"}
	paramsJSON, _ := json.Marshal(params)

	result, err := server.handleAnalyzeChangeImpact(paramsJSON)
	if err != nil {
		t.Fatalf("handleAnalyzeChangeImpact failed: %v", err)
	}

	impact, ok := result.(*model.ChangeImpact)
	if !ok {
		t.Fatal("Expected ChangeImpact result")
	}

	if impact == nil {
		t.Fatal("Expected ChangeImpact result")
	}

	if impact.RiskLevel <= 0.0 {
		t.Error("Expected risk level in impact")
	}
}

func TestMCPServer_HandleGetCodeMetrics(t *testing.T) {
	server, indexer, projectPath := setupTestMCPServer(t)
	defer indexer.Close()

	// Create a test file with complex function
	goFile := filepath.Join(projectPath, "code.go")
	code := `package main

func Complex(x int) int {
	if x > 10 {
		if x > 20 {
			return x * 2
		}
		return x + 10
	}
	return x
}
`
	os.WriteFile(goFile, []byte(code), 0644)
	indexer.IndexAll()

	// Call get_code_metrics tool
	params := map[string]string{"symbol_name": "Complex"}
	paramsJSON, _ := json.Marshal(params)

	result, err := server.handleGetCodeMetrics(paramsJSON)
	if err != nil {
		t.Fatalf("handleGetCodeMetrics failed: %v", err)
	}

	metrics, ok := result.(*model.CodeMetrics)
	if !ok {
		t.Fatal("Expected CodeMetrics result")
	}

	if metrics.SymbolName == "" {
		t.Error("Expected symbol name in metrics")
	}

	if metrics.Cyclomatic == 0 {
		t.Error("Expected non-zero cyclomatic complexity")
	}
}

func TestMCPServer_HandleSimulateChange(t *testing.T) {
	server, indexer, projectPath := setupTestMCPServer(t)
	defer indexer.Close()

	// Create a test file
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
	os.WriteFile(goFile, []byte(code), 0644)
	indexer.IndexAll()

	// Call simulate_change tool
	params := map[string]string{
		"symbol_name": "OldName",
		"change_type": "rename",
		"new_value":   "NewName",
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := server.handleSimulateChange(paramsJSON)
	if err != nil {
		t.Fatalf("handleSimulateChange failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map result")
	}

	if resultMap["symbol"] != "OldName" {
		t.Error("Expected symbol name in result")
	}

	if resultMap["change_type"] != "rename" {
		t.Error("Expected change_type in result")
	}

	if resultMap["can_auto_fix"] == nil {
		t.Error("Expected can_auto_fix in result")
	}
}

func TestMCPServer_HandleBuildDependencyGraph(t *testing.T) {
	server, indexer, projectPath := setupTestMCPServer(t)
	defer indexer.Close()

	// Create a test file with dependencies
	goFile := filepath.Join(projectPath, "code.go")
	code := `package main

func Helper() string {
	return "help"
}

func Main() {
	result := Helper()
	println(result)
}
`
	os.WriteFile(goFile, []byte(code), 0644)
	indexer.IndexAll()

	// Call build_dependency_graph tool
	params := map[string]interface{}{
		"symbol_name": "Main",
		"max_depth":   2,
	}
	paramsJSON, _ := json.Marshal(params)

	result, err := server.handleBuildDependencyGraph(paramsJSON)
	if err != nil {
		t.Fatalf("handleBuildDependencyGraph failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map result")
	}

	if resultMap["symbol"] != "Main" {
		t.Error("Expected symbol name in result")
	}

	if resultMap["total_nodes"] == nil {
		t.Error("Expected total_nodes in result")
	}

	if resultMap["graph"] == nil {
		t.Error("Expected graph in result")
	}
}

func TestMCPServer_HandleListFiles(t *testing.T) {
	server, indexer, projectPath := setupTestMCPServer(t)
	defer indexer.Close()

	// Create test files
	files := []string{"file1.go", "file2.go", "file3.py"}
	for _, name := range files {
		filePath := filepath.Join(projectPath, name)
		os.WriteFile(filePath, []byte("package main"), 0644)
	}

	indexer.IndexAll()

	// Call list_files tool
	result, err := server.handleListFiles(json.RawMessage("{}"))
	if err != nil {
		t.Fatalf("handleListFiles failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map result")
	}

	count, ok := resultMap["count"].(int)
	if !ok {
		t.Fatal("Expected count in result")
	}

	if count != 3 {
		t.Errorf("Expected 3 files, got %d", count)
	}
}

func TestMCPServer_HandleInvalidToolCall(t *testing.T) {
	server, indexer, _ := setupTestMCPServer(t)
	defer indexer.Close()

	// Call non-existent tool
	params := json.RawMessage(`{"name": "non_existent_tool"}`)

	result, err := server.handleToolCall(params)
	if err == nil {
		t.Error("Expected error for non-existent tool")
	}

	if result != nil {
		t.Error("Expected nil result for non-existent tool")
	}
}

func TestMCPServer_HandleInvalidParams(t *testing.T) {
	server, indexer, _ := setupTestMCPServer(t)
	defer indexer.Close()

	// Call with invalid JSON
	params := json.RawMessage(`{invalid json}`)

	_, err := server.handleSearchSymbols(params)
	if err == nil {
		t.Error("Expected error for invalid JSON params")
	}
}

func TestMCPServer_RegisterTool(t *testing.T) {
	server, indexer, _ := setupTestMCPServer(t)
	defer indexer.Close()

	// Verify tools are registered
	if len(server.tools) != 19 {
		t.Errorf("Expected 19 registered tools, got %d", len(server.tools))
	}

	// Verify specific tools exist
	expectedTools := []string{
		"search_symbols",
		"get_code_context",
		"simulate_change",
		"build_dependency_graph",
	}

	for _, name := range expectedTools {
		if _, exists := server.tools[name]; !exists {
			t.Errorf("Expected tool %s to be registered", name)
		}
	}
}
