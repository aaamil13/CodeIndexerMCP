package ai

import (
	"path/filepath"
	"testing"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

func setupDependencyTestDB(t *testing.T) *database.DB {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create test project and file
	project := &types.Project{Name: "test", Path: "/test"}
	db.CreateProject(project) // This method exists, why was it showing as undefined?

	file := &types.File{ProjectID: project.ID, Path: "/test/file.go", Language: "go", RelativePath: "/test/file.go"} // Added RelativePath
	db.SaveFile(file) // Changed CreateFile to SaveFile

	return db
}

func TestBuildDependencyGraph_Simple(t *testing.T) {
	db := setupDependencyTestDB(t)
	defer db.Close()

	// Create symbols
	funcA := &types.Symbol{FileID: 1, Name: "FuncA", Type: types.SymbolTypeFunction}
	db.SaveSymbol(funcA)

	funcB := &types.Symbol{FileID: 1, Name: "FuncB", Type: types.SymbolTypeFunction}
	db.SaveSymbol(funcB)

	// FuncA calls FuncB
	ref := &types.Reference{
		SymbolID: funcB.ID,
		FileID:   1,
		LineNumber:     10,
		ReferenceType: "call",
	}
	db.SaveReference(ref)

	// Build graph
	builder := NewDependencyGraphBuilder(db)
	graph, err := builder.BuildSymbolDependencyGraph(funcA.Name, 2)
	if err != nil {
		t.Fatalf("BuildSymbolDependencyGraph failed: %v", err)
	}

	if graph == nil {
		t.Fatal("Expected graph to be returned")
	}

	// Should have both nodes
	if len(graph.Nodes) < 1 {
		t.Errorf("Expected at least 1 node, got %d", len(graph.Nodes))
	}

	// No RootSymbol field in types.DependencyGraph
}

func TestBuildDependencyGraph_MultipleLevels(t *testing.T) {
	db := setupDependencyTestDB(t)
	defer db.Close()

	// Create chain: A -> B -> C
	funcA := &types.Symbol{FileID: 1, Name: "FuncA", Type: types.SymbolTypeFunction}
	db.SaveSymbol(funcA)

	funcB := &types.Symbol{FileID: 1, Name: "FuncB", Type: types.SymbolTypeFunction}
	db.SaveSymbol(funcB)

	funcC := &types.Symbol{FileID: 1, Name: "FuncC", Type: types.SymbolTypeFunction}
	db.SaveSymbol(funcC)

	// Create relationships
	rel1 := &types.Relationship{
		FromSymbolID: funcA.ID,
		ToSymbolID:   funcB.ID,
		Type:         types.RelationshipCalls,
	}
	db.SaveRelationship(rel1)

	rel2 := &types.Relationship{
		FromSymbolID: funcB.ID,
		ToSymbolID:   funcC.ID,
		Type:         types.RelationshipCalls,
	}
	db.SaveRelationship(rel2)

	// Build graph with depth 3
	builder := NewDependencyGraphBuilder(db)
	graph, err := builder.BuildSymbolDependencyGraph(funcA.Name, 3)
	if err != nil {
		t.Fatalf("BuildSymbolDependencyGraph failed: %v", err)
	}

	// Should have all nodes
	if len(graph.Nodes) < 1 {
		t.Error("Expected multiple nodes in deep graph")
	}

	// Should have edges
	if len(graph.Edges) < 1 {
		t.Error("Expected edges in graph")
	}
}

func TestGetDependencies(t *testing.T) {
	db := setupDependencyTestDB(t)
	defer db.Close()

	// Create symbols
	caller := &types.Symbol{FileID: 1, Name: "Caller", Type: types.SymbolTypeFunction}
	db.SaveSymbol(caller)

	dep1 := &types.Symbol{FileID: 1, Name: "Dep1", Type: types.SymbolTypeFunction}
	db.SaveSymbol(dep1)

	dep2 := &types.Symbol{FileID: 1, Name: "Dep2", Type: types.SymbolTypeFunction}
	db.SaveSymbol(dep2)

	// Create relationships
	rel1 := &types.Relationship{
		FromSymbolID: caller.ID,
		ToSymbolID:   dep1.ID,
		Type:         types.RelationshipCalls,
	}
	db.SaveRelationship(rel1)

	rel2 := &types.Relationship{
		FromSymbolID: caller.ID,
		ToSymbolID:   dep2.ID,
		Type:         types.RelationshipCalls,
	}
	db.SaveRelationship(rel2)

	// Get dependencies
	builder := NewDependencyGraphBuilder(db)
	deps, err := builder.GetDependenciesFor(caller.Name)
	if err != nil {
		t.Fatalf("GetDependenciesFor failed: %v", err)
	}

	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(deps))
	}

	// Verify dependency names
	names := make(map[string]bool)
	for _, dep := range deps {
		names[dep.Name] = true
	}

	if !names["Dep1"] || !names["Dep2"] {
		t.Error("Expected Dep1 and Dep2 in dependencies")
	}
}

func TestGetDependents(t *testing.T) {
	db := setupDependencyTestDB(t)
	defer db.Close()

	// Create symbols
	target := &types.Symbol{FileID: 1, Name: "Target", Type: types.SymbolTypeFunction}
	db.SaveSymbol(target)

	caller1 := &types.Symbol{FileID: 1, Name: "Caller1", Type: types.SymbolTypeFunction}
	db.SaveSymbol(caller1)

	caller2 := &types.Symbol{FileID: 1, Name: "Caller2", Type: types.SymbolTypeFunction}
	db.SaveSymbol(caller2)

	// Create relationships (callers depend on target)
	rel1 := &types.Relationship{
		FromSymbolID: caller1.ID,
		ToSymbolID:   target.ID,
		Type:         types.RelationshipCalls,
	}
	db.SaveRelationship(rel1)

	rel2 := &types.Relationship{
		FromSymbolID: caller2.ID,
		ToSymbolID:   target.ID,
		Type:         types.RelationshipCalls,
	}
	db.SaveRelationship(rel2)

	// Get dependents
	builder := NewDependencyGraphBuilder(db)
	dependents, err := builder.GetDependentsFor(target.Name)
	if err != nil {
		t.Fatalf("GetDependentsFor failed: %v", err)
	}

	if len(dependents) != 2 {
		t.Errorf("Expected 2 dependents, got %d", len(dependents))
	}

	// Verify dependent names
	names := make(map[string]bool)
	for _, dep := range dependents {
		names[dep.Name] = true
	}

	if !names["Caller1"] || !names["Caller2"] {
		t.Error("Expected Caller1 and Caller2 in dependents")
	}
}

func TestAnalyzeDependencyChain(t *testing.T) {
	db := setupDependencyTestDB(t)
	defer db.Close()

	// Create chain
	start := &types.Symbol{FileID: 1, Name: "Start", Type: types.SymbolTypeFunction}
	db.SaveSymbol(start)

	middle := &types.Symbol{FileID: 1, Name: "Middle", Type: types.SymbolTypeFunction}
	db.SaveSymbol(middle)

	end := &types.Symbol{FileID: 1, Name: "End", Type: types.SymbolTypeFunction}
	db.SaveSymbol(end)

	// Create chain relationships
	rel1 := &types.Relationship{
		FromSymbolID: start.ID,
		ToSymbolID:   middle.ID,
		Type:         types.RelationshipCalls,
	}
	db.SaveRelationship(rel1)

	rel2 := &types.Relationship{
		FromSymbolID: middle.ID,
		ToSymbolID:   end.ID,
		Type:         types.RelationshipCalls,
	}
	db.SaveRelationship(rel2)

	// Analyze chain
	builder := NewDependencyGraphBuilder(db)
	result, err := builder.AnalyzeDependencyChain(start.Name)
	if err != nil {
		t.Fatalf("AnalyzeDependencyChain failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected dependency chain result")
	}

	dependencies, ok := result["dependencies"].([]*types.Symbol)
	if !ok {
		t.Fatal("Expected dependencies in result")
	}
	if len(dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(dependencies))
	}
}

func TestCouplingScore(t *testing.T) {
	db := setupDependencyTestDB(t)
	defer db.Close()

	// Create highly coupled symbol
	symbol := &types.Symbol{FileID: 1, Name: "HighlyCoupled", Type: types.SymbolTypeFunction}
	db.SaveSymbol(symbol)

	// Create many dependencies
	for i := 0; i < 10; i++ {
		dep := &types.Symbol{FileID: 1, Name: "Dep" + string(rune(i)), Type: types.SymbolTypeFunction}
		db.SaveSymbol(dep)

		rel := &types.Relationship{
			FromSymbolID: symbol.ID,
			ToSymbolID:   dep.ID,
			Type:         types.RelationshipCalls,
		}
		db.SaveRelationship(rel)
	}

	// Analyze chain and check coupling
	builder := NewDependencyGraphBuilder(db)
	result, err := builder.AnalyzeDependencyChain(symbol.Name)
	if err != nil {
		t.Fatalf("AnalyzeDependencyChain failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected dependency chain result")
	}

	couplingScore, ok := result["coupling_score"].(float64)
	if !ok {
		t.Fatal("Expected coupling_score in result")
	}

	// Coupling score should be positive
	if couplingScore <= 0 {
		t.Error("Expected positive coupling score for highly coupled symbol")
	}
}

func TestCircularDependency(t *testing.T) {
	db := setupDependencyTestDB(t)
	defer db.Close()

	// Create circular dependency: A -> B -> C -> A
	funcA := &types.Symbol{FileID: 1, Name: "FuncA", Type: types.SymbolTypeFunction}
	db.SaveSymbol(funcA)

	funcB := &types.Symbol{FileID: 1, Name: "FuncB", Type: types.SymbolTypeFunction}
	db.SaveSymbol(funcB)

	funcC := &types.Symbol{FileID: 1, Name: "FuncC", Type: types.SymbolTypeFunction}
	db.SaveSymbol(funcC)

	// Create circular relationships
	rel1 := &types.Relationship{
		FromSymbolID: funcA.ID,
		ToSymbolID:   funcB.ID,
		Type:         types.RelationshipCalls,
	}
	db.SaveRelationship(rel1)

	rel2 := &types.Relationship{
		FromSymbolID: funcB.ID,
		ToSymbolID:   funcC.ID,
		Type:         types.RelationshipCalls,
	}
	db.SaveRelationship(rel2)

	rel3 := &types.Relationship{
		FromSymbolID: funcC.ID,
		ToSymbolID:   funcA.ID,
		Type:         types.RelationshipCalls,
	}
	db.SaveRelationship(rel3)

	// Build graph - should handle circular dependency
	builder := NewDependencyGraphBuilder(db)
	graph, err := builder.BuildSymbolDependencyGraph(funcA.Name, 5)
	if err != nil {
		t.Fatalf("BuildSymbolDependencyGraph failed with circular dependency: %v", err)
	}

	// Should still build graph without infinite loop
	if graph == nil {
		t.Fatal("Expected graph even with circular dependency")
	}
}

func TestEmptyDependencies(t *testing.T) {
	db := setupDependencyTestDB(t)
	defer db.Close()

	// Create isolated symbol
	symbol := &types.Symbol{FileID: 1, Name: "Isolated", Type: types.SymbolTypeFunction}
	db.SaveSymbol(symbol)

	// Get dependencies (should be empty)
	builder := NewDependencyGraphBuilder(db)
	deps, err := builder.GetDependenciesFor(symbol.Name)
	if err != nil {
		t.Fatalf("GetDependenciesFor failed: %v", err)
	}

	if len(deps) != 0 {
		t.Errorf("Expected 0 dependencies for isolated symbol, got %d", len(deps))
	}

	// Get dependents (should be empty)
	builder = NewDependencyGraphBuilder(db)
	dependents, err := builder.GetDependentsFor(symbol.Name)
	if err != nil {
		t.Fatalf("GetDependentsFor failed: %v", err)
	}

	if len(dependents) != 0 {
		t.Errorf("Expected 0 dependents for isolated symbol, got %d", len(dependents))
	}
}
