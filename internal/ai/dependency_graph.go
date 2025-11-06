package ai

import (
	"fmt"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// DependencyGraphBuilder builds dependency graphs
type DependencyGraphBuilder struct {
	db *database.DB
}

// NewDependencyGraphBuilder creates a new dependency graph builder
func NewDependencyGraphBuilder(db *database.DB) *DependencyGraphBuilder {
	return &DependencyGraphBuilder{db: db}
}

// BuildSymbolDependencyGraph builds a dependency graph for a symbol
func (dgb *DependencyGraphBuilder) BuildSymbolDependencyGraph(symbolName string, maxDepth int) (*types.DependencyGraph, error) {
	symbol, err := dgb.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	file, err := dgb.db.GetFile(symbol.FileID)
	if err != nil {
		return nil, err
	}

	graph := &types.DependencyGraph{
		Nodes: []*types.DependencyNode{},
		Edges: []*types.DependencyEdge{},
	}

	visited := make(map[int64]bool)

	// Add root node
	rootNode := &types.DependencyNode{
		Symbol: symbol,
		File:   file,
		Type:   "symbol",
		Level:  0,
	}
	graph.Nodes = append(graph.Nodes, rootNode)
	visited[symbol.ID] = true

	// Build graph recursively
	dgb.buildGraphRecursive(symbol, graph, visited, 0, maxDepth)

	return graph, nil
}

// buildGraphRecursive builds graph recursively
func (dgb *DependencyGraphBuilder) buildGraphRecursive(symbol *types.Symbol, graph *types.DependencyGraph, visited map[int64]bool, currentDepth, maxDepth int) {
	if currentDepth >= maxDepth {
		return
	}

	// Get relationships
	relationships, err := dgb.db.GetRelationshipsForSymbol(symbol.ID)
	if err != nil {
		return
	}

	for _, rel := range relationships {
		var targetSymbolID int64
		var edgeFrom, edgeTo string

		if rel.FromSymbolID == symbol.ID {
			// This symbol depends on another
			targetSymbolID = rel.ToSymbolID
			edgeFrom = fmt.Sprintf("symbol_%d", symbol.ID)
			edgeTo = fmt.Sprintf("symbol_%d", targetSymbolID)
		} else {
			// Another symbol depends on this
			targetSymbolID = rel.FromSymbolID
			edgeFrom = fmt.Sprintf("symbol_%d", targetSymbolID)
			edgeTo = fmt.Sprintf("symbol_%d", symbol.ID)
		}

		// Skip if already visited
		if visited[targetSymbolID] {
			continue
		}

		// Get target symbol (simplified - in production we'd have a better query)
		// For now, skip if we can't get it
		visited[targetSymbolID] = true

		// Add edge
		edge := &types.DependencyEdge{
			From:   edgeFrom,
			To:     edgeTo,
			Type:   string(rel.Type),
			Weight: 1,
		}
		graph.Edges = append(graph.Edges, edge)
	}
}

// BuildFileDependencyGraph builds a dependency graph for a file
func (dgb *DependencyGraphBuilder) BuildFileDependencyGraph(filePath string, maxDepth int) (*types.DependencyGraph, error) {
	graph := &types.DependencyGraph{
		Nodes: []*types.DependencyNode{},
		Edges: []*types.DependencyEdge{},
	}

	// Implementation would track file-level dependencies through imports

	return graph, nil
}

// GetDependenciesFor gets all dependencies for a symbol
func (dgb *DependencyGraphBuilder) GetDependenciesFor(symbolName string) ([]*types.Symbol, error) {
	symbol, err := dgb.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	dependencies := []*types.Symbol{}
	visited := make(map[int64]bool)

	dgb.collectDependencies(symbol, &dependencies, visited, 0, 5)

	return dependencies, nil
}

// collectDependencies collects dependencies recursively
func (dgb *DependencyGraphBuilder) collectDependencies(symbol *types.Symbol, deps *[]*types.Symbol, visited map[int64]bool, depth, maxDepth int) {
	if depth >= maxDepth || visited[symbol.ID] {
		return
	}

	visited[symbol.ID] = true

	relationships, err := dgb.db.GetRelationshipsForSymbol(symbol.ID)
	if err != nil {
		return
	}

	for _, rel := range relationships {
		// Only follow outgoing dependencies (what this symbol uses)
		if rel.FromSymbolID == symbol.ID {
			// In production, we'd fetch the target symbol properly
			// For now, this is a placeholder
			visited[rel.ToSymbolID] = true
		}
	}
}

// GetDependentsFor gets all dependents (who depends on this symbol)
func (dgb *DependencyGraphBuilder) GetDependentsFor(symbolName string) ([]*types.Symbol, error) {
	symbol, err := dgb.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	// Get all references (simpler approach)
	references, err := dgb.db.GetReferencesBySymbol(symbol.ID)
	if err != nil {
		return nil, err
	}

	dependents := []*types.Symbol{}
	seenFiles := make(map[int64]bool)

	for _, ref := range references {
		if seenFiles[ref.FileID] {
			continue
		}
		seenFiles[ref.FileID] = true

		// Get all symbols in this file that might depend on our symbol
		symbols, err := dgb.db.GetSymbolsByFile(ref.FileID)
		if err != nil {
			continue
		}

		for _, sym := range symbols {
			// Check if this symbol contains the reference
			if sym.StartLine <= ref.LineNumber && sym.EndLine >= ref.LineNumber {
				dependents = append(dependents, sym)
				break
			}
		}
	}

	return dependents, nil
}

// AnalyzeDependencyChain analyzes the full dependency chain
func (dgb *DependencyGraphBuilder) AnalyzeDependencyChain(symbolName string) (map[string]interface{}, error) {
	dependencies, err := dgb.GetDependenciesFor(symbolName)
	if err != nil {
		return nil, err
	}

	dependents, err := dgb.GetDependentsFor(symbolName)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"symbol":           symbolName,
		"dependencies":     dependencies,
		"dependency_count": len(dependencies),
		"dependents":       dependents,
		"dependent_count":  len(dependents),
		"coupling_score":   float64(len(dependents)) / float64(len(dependencies)+1),
	}, nil
}
