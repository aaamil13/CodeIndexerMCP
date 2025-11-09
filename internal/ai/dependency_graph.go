package ai

import (
	"fmt"
	"strconv"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

// DependencyGraphBuilder builds dependency graphs
type DependencyGraphBuilder struct {
	db *database.Manager
}

// NewDependencyGraphBuilder creates a new dependency graph builder
func NewDependencyGraphBuilder(db *database.Manager) *DependencyGraphBuilder {
	return &DependencyGraphBuilder{db: db}
}

// BuildSymbolDependencyGraph builds a dependency graph for a symbol
func (dgb *DependencyGraphBuilder) BuildSymbolDependencyGraph(symbolName string, maxDepth int) (*model.DependencyGraph, error) {
	symbol, err := dgb.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	// Retrieve the file associated with the symbol
	file, err := dgb.db.GetFileByID(symbol.FileID)
	if err != nil || file == nil {
		return nil, fmt.Errorf("file not found for symbol %d: %w", symbol.ID, err)
	}

	graph := &model.DependencyGraph{
		Nodes: []*model.DependencyNode{},
		Edges: []*model.DependencyEdge{},
	}

	visitedNodes := make(map[string]bool)
	visitedEdges := make(map[string]bool) // To prevent duplicate edges

	// Add root node
	rootNode := &model.DependencyNode{
		SymbolID: strconv.Itoa(symbol.ID), // Convert int to string
		Name:     symbol.Name,
		Kind:     string(symbol.Kind), // Cast to string
		File:     file.Path,
		Type:     "symbol",
		Level:    0,
	}
	graph.Nodes = append(graph.Nodes, rootNode)
	visitedNodes[strconv.Itoa(symbol.ID)] = true

	dgb.buildGraphRecursive(symbol, graph, visitedNodes, visitedEdges, 0, maxDepth)

	return graph, nil
}

// buildGraphRecursive builds graph recursively
func (dgb *DependencyGraphBuilder) buildGraphRecursive(currentSymbol *model.Symbol, graph *model.DependencyGraph, visitedNodes map[string]bool, visitedEdges map[string]bool, currentDepth, maxDepth int) {
	if currentDepth >= maxDepth {
		return
	}

	// Get references where currentSymbol is the source (i.e., currentSymbol depends on target)
	outgoingReferences, err := dgb.db.GetReferencesBySymbol(currentSymbol.ID)
	if err != nil {
		// Log error but continue
		fmt.Printf("Error getting outgoing references for symbol %s: %v\n", currentSymbol.ID, err)
	} else {
		for _, ref := range outgoingReferences {
			targetSymbol, err := dgb.db.GetSymbolByName(ref.TargetSymbolName) // Assuming target_symbol_name can be resolved to a symbol
			if err != nil || targetSymbol == nil {
				// Couldn't resolve target symbol, skip
				continue
			}

			// Add target node if not visited
			if !visitedNodes[strconv.Itoa(targetSymbol.ID)] {
				// Retrieve the file associated with the target symbol
				targetFile, err := dgb.db.GetFileByID(targetSymbol.FileID)
				if err != nil || targetFile == nil {
					fmt.Printf("Error: File not found for target symbol %d: %v\n", targetSymbol.ID, err)
					continue
				}

				targetNode := &model.DependencyNode{
					SymbolID: strconv.Itoa(targetSymbol.ID), // Convert int to string
					Name:     targetSymbol.Name,
					Kind:     string(targetSymbol.Kind), // Cast to string
					File:     targetFile.Path,
					Type:     "symbol",
					Level:    currentDepth + 1,
				}
				graph.Nodes = append(graph.Nodes, targetNode)
				visitedNodes[strconv.Itoa(targetSymbol.ID)] = true
			}

			// Add edge
			edgeKey := fmt.Sprintf("%s-%s-%s", strconv.Itoa(currentSymbol.ID), strconv.Itoa(targetSymbol.ID), ref.ReferenceType)
			if !visitedEdges[edgeKey] {
				edge := &model.DependencyEdge{
					From:   strconv.Itoa(currentSymbol.ID),
					To:     strconv.Itoa(targetSymbol.ID),
					Type:   string(ref.ReferenceType), // Convert to string
					Weight: 1, // Default weight
				}
				graph.Edges = append(graph.Edges, edge)
				visitedEdges[edgeKey] = true
			}

			// Recursively build graph for target symbol
			dgb.buildGraphRecursive(targetSymbol, graph, visitedNodes, visitedEdges, currentDepth+1, maxDepth)
		}
	}

	// Get references where currentSymbol is the target (i.e., target depends on currentSymbol)
	// This part would require a GetDependentsBySymbol method in database.Manager, or a more complex query.
	// For now, we focus on outgoing dependencies as 'dependencies' typically refers to what a component needs.
}

// BuildFileDependencyGraph builds a dependency graph for a file
func (dgb *DependencyGraphBuilder) BuildFileDependencyGraph(filePath string, maxDepth int) (*model.DependencyGraph, error) {
	graph := &model.DependencyGraph{
		Nodes: []*model.DependencyNode{},
		Edges: []*model.DependencyEdge{},
	}

	// Implementation would track file-level dependencies through imports

	return graph, nil
}

// GetDependenciesFor gets all dependencies for a symbol
func (dgb *DependencyGraphBuilder) GetDependenciesFor(symbolName string) ([]*model.Symbol, error) {
	symbol, err := dgb.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol by name %s: %w", symbolName, err)
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	dependencies := []*model.Symbol{}
	visited := make(map[string]bool) // Track visited symbols to avoid infinite loops and duplicates

	// Collect direct outgoing dependencies (what this symbol references)
	dgb.collectDependencies(symbol, &dependencies, visited, 0, 5) // Max depth of 5 for collecting

	return dependencies, nil
}

// collectDependencies collects dependencies recursively
func (dgb *DependencyGraphBuilder) collectDependencies(currentSymbol *model.Symbol, deps *[]*model.Symbol, visited map[string]bool, depth, maxDepth int) {
	if depth >= maxDepth || visited[strconv.Itoa(currentSymbol.ID)] {
		return
	}

	visited[strconv.Itoa(currentSymbol.ID)] = true
	*deps = append(*deps, currentSymbol)

	// Get outgoing references (what currentSymbol uses)
	references, err := dgb.db.GetReferencesBySymbol(currentSymbol.ID)
	if err != nil {
		fmt.Printf("Error getting references for symbol %s: %v\n", currentSymbol.ID, err)
		return
	}

	for _, ref := range references {
		targetSymbol, err := dgb.db.GetSymbolByName(ref.TargetSymbolName) // Assuming target_symbol_name is resolvable
		if err != nil || targetSymbol == nil {
			continue // Skip if target symbol cannot be found
		}
		dgb.collectDependencies(targetSymbol, deps, visited, depth+1, maxDepth)
	}
}

// GetDependentsFor gets all dependents (who depends on this symbol)
func (dgb *DependencyGraphBuilder) GetDependentsFor(symbolName string) ([]*model.Symbol, error) {
	symbol, err := dgb.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol by name %s: %w", symbolName, err)
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	// Get all references where this symbol is the target
	// This requires a new database method to efficiently query references by target_symbol_name
	// For now, we'll iterate through all symbols and their references (inefficient but works for now)
	// A more efficient approach would be to have an index on target_symbol_name in the references table.
	// (Which we have already created as idx_code_reference_target)
	
	dependents := []*model.Symbol{}
	visited := make(map[string]bool)

	// This assumes GetReferencesBySymbol returns references where either source or target matches symbolID
	// which is how GetReferencesBySymbol is currently implemented.
	references, err := dgb.db.GetReferencesBySymbol(symbol.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get references for symbol %s: %w", symbol.ID, err)
	}

	for _, ref := range references {
		// If our symbol is the target, then ref.SourceSymbolID is a dependent
		// Note: The previous logic assumed TargetSymbolName stores the actual name, not ID.
		// If TargetSymbolName is expected to be the ID, this comparison needs to be updated.
		// For now, let's assume it's the name and that GetSymbolByName is robust enough.
		targetSym, err := dgb.db.GetSymbolByName(ref.TargetSymbolName) // Need a way to ensure this is the *correct* target symbol
		if err != nil || targetSym == nil || targetSym.ID != symbol.ID {
			// This reference is not for our symbol, or the target symbol couldn't be resolved.
			continue
		}

		dependentSymbol, err := dgb.db.GetSymbolByID(ref.SourceSymbolID)
		if err != nil || dependentSymbol == nil {
			continue // Skip if source symbol cannot be found
		}
		if !visited[strconv.Itoa(dependentSymbol.ID)] { // Convert int to string for map key
			dependents = append(dependents, dependentSymbol)
			visited[strconv.Itoa(dependentSymbol.ID)] = true
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
