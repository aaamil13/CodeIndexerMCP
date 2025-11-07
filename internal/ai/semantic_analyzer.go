package ai

import (
	"fmt"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// SemanticAnalyzer performs semantic analysis across files
type SemanticAnalyzer struct {
	db            *database.Database
	typeValidator *TypeValidator
}

// NewSemanticAnalyzer creates a new semantic analyzer
func NewSemanticAnalyzer(db *database.Database) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		db:            db,
		typeValidator: NewTypeValidator(db),
	}
}

// AnalyzeProject performs full semantic analysis on a project
func (sa *SemanticAnalyzer) AnalyzeProject(projectID int64) (*types.SemanticAnalysisResult, error) {
	result := &types.SemanticAnalysisResult{
		ProjectID:           projectID,
		TypeErrors:          make([]*types.TypeMismatch, 0),
		UndefinedReferences: make([]*types.UndefinedUsage, 0),
		UnusedSymbols:       make([]*types.Symbol, 0),
		CircularDeps:        make([]*types.CircularDependency, 0),
		Warnings:            make([]string, 0),
		Metrics:             make(map[string]interface{}),
	}

	// Get all files in project
	files, err := sa.db.GetAllFilesForProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}

	// Analyze each file
	for _, file := range files {
		// Type validation
		validation, err := sa.typeValidator.ValidateFile(file.ID)
		if err != nil {
			continue
		}

		result.TypeErrors = append(result.TypeErrors, validation.TypeMismatches...)
		result.UndefinedReferences = append(result.UndefinedReferences, validation.UndefinedSymbols...)
	}

	// Cross-file analysis
	sa.analyzeCrossDependencies(projectID, result)

	// Find unused symbols
	unusedSymbols, err := sa.findUnusedSymbols(projectID)
	if err == nil {
		result.UnusedSymbols = unusedSymbols
	}

	// Detect circular dependencies
	circular := sa.detectCircularDependencies(projectID)
	result.CircularDeps = circular

	// Calculate metrics
	result.Metrics["total_files"] = len(files)
	result.Metrics["type_errors"] = len(result.TypeErrors)
	result.Metrics["undefined_references"] = len(result.UndefinedReferences)
	result.Metrics["unused_symbols"] = len(result.UnusedSymbols)
	result.Metrics["circular_dependencies"] = len(result.CircularDeps)

	// Calculate semantic quality score
	result.QualityScore = sa.calculateQualityScore(result)

	return result, nil
}

// InferType infers the type of a symbol
func (sa *SemanticAnalyzer) InferType(symbolID int64) (*types.TypeInference, error) {
	symbol, err := sa.db.GetSymbol(symbolID)
	if err != nil {
		return nil, err
	}

	inference := &types.TypeInference{
		SymbolName: symbol.Name,
		Confidence: 0.5, // Default confidence
	}

	// Type inference based on symbol type
	switch symbol.Type {
	case types.SymbolTypeFunction, types.SymbolTypeMethod:
		inference.InferredType = sa.inferFunctionType(symbol)
		inference.Confidence = 0.8
		inference.Reasoning = "Inferred from function signature"

	case types.SymbolTypeVariable:
		inference.InferredType = sa.inferVariableType(symbol)
		inference.Confidence = 0.6
		inference.Reasoning = "Inferred from usage context"

	case types.SymbolTypeClass:
		inference.InferredType = symbol.Name // Class name is its type
		inference.Confidence = 1.0
		inference.Reasoning = "Class definition"

	case types.SymbolTypeInterface:
		inference.InferredType = symbol.Name
		inference.Confidence = 1.0
		inference.Reasoning = "Interface definition"

	default:
		inference.InferredType = "unknown"
		inference.Confidence = 0.0
		inference.Reasoning = "Unable to infer type"
	}

	return inference, nil
}

func (sa *SemanticAnalyzer) inferFunctionType(symbol *types.Symbol) string {
	// Parse function signature to extract return type
	signature := symbol.Signature

	// For Go: func Name() returnType
	if strings.Contains(signature, ")") {
		parts := strings.Split(signature, ")")
		if len(parts) >= 2 {
			returnPart := strings.TrimSpace(parts[1])
			if returnPart != "" && returnPart != "{" {
				return returnPart
			}
		}
	}

	// For TypeScript: (params): returnType =>
	if strings.Contains(signature, "):") {
		parts := strings.Split(signature, "):")
		if len(parts) >= 2 {
			returnPart := strings.TrimSpace(parts[1])
			returnPart = strings.Split(returnPart, "=>")[0]
			return strings.TrimSpace(returnPart)
		}
	}

	return "unknown"
}

func (sa *SemanticAnalyzer) inferVariableType(symbol *types.Symbol) string {
	// Try to infer from signature or context
	signature := symbol.Signature

	// For typed languages: var name type = value
	parts := strings.Fields(signature)
	if len(parts) >= 3 {
		// Second part might be the type
		potentialType := parts[1]
		if potentialType != "=" && potentialType != ":=" {
			return potentialType
		}
	}

	// For TypeScript: const name: type = value
	if strings.Contains(signature, ":") {
		parts := strings.Split(signature, ":")
		if len(parts) >= 2 {
			typePart := strings.Split(parts[1], "=")[0]
			return strings.TrimSpace(typePart)
		}
	}

	return "unknown"
}

// ResolveCrossFileReference resolves a reference across files
func (sa *SemanticAnalyzer) ResolveCrossFileReference(symbolName string, projectID int64) ([]*types.Symbol, error) {
	// Search for symbol across all files in project
	symbols, err := sa.db.SearchSymbols(types.SearchOptions{
		Query:     symbolName,
		ProjectID: projectID,
	})
	if err != nil {
		return nil, err
	}

	return symbols, nil
}

func (sa *SemanticAnalyzer) analyzeCrossDependencies(projectID int64, result *types.SemanticAnalysisResult) {
	// Get all symbols in project
	files, err := sa.db.GetAllFilesForProject(projectID)
	if err != nil {
		return
	}

	// Build dependency map
	dependencyMap := make(map[int64][]int64) // fileID -> []dependentFileIDs

	for _, file := range files {
		imports, err := sa.db.GetImportsByFile(file.ID)
		if err != nil {
			continue
		}

		// For each import, try to find the corresponding file
		for _, imp := range imports {
			// This is simplified - would need proper module resolution
			targetFiles, err := sa.findFilesByImport(imp.Source, projectID)
			if err != nil {
				continue
			}

			for _, targetFile := range targetFiles {
				dependencyMap[file.ID] = append(dependencyMap[file.ID], targetFile.ID)
			}
		}
	}

	// Store dependency information in result
	result.Metrics["dependency_map"] = dependencyMap
}

func (sa *SemanticAnalyzer) findFilesByImport(importPath string, projectID int64) ([]*types.File, error) {
	// Simplified import resolution
	// In production, this would handle:
	// - Relative imports (./module, ../module)
	// - Absolute imports (package/module)
	// - Node modules (@scope/package)
	// - Python packages (django.db)

	files, err := sa.db.GetAllFilesForProject(projectID)
	if err != nil {
		return nil, err
	}

	matching := make([]*types.File, 0)
	for _, file := range files {
		// Simple name matching
		if strings.Contains(file.Path, importPath) {
			matching = append(matching, file)
		}
	}

	return matching, nil
}

func (sa *SemanticAnalyzer) findUnusedSymbols(projectID int64) ([]*types.Symbol, error) {
	unused := make([]*types.Symbol, 0)

	files, err := sa.db.GetAllFilesForProject(projectID)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		symbols, err := sa.db.GetSymbolsByFile(file.ID)
		if err != nil {
			continue
		}

		for _, symbol := range symbols {
			// Skip exported symbols (might be used externally)
			if symbol.IsExported {
				continue
			}

			// Check if symbol has any references
			refs, err := sa.db.GetReferencesBySymbol(symbol.ID)
			if err != nil {
				continue
			}

			if len(refs) == 0 {
				unused = append(unused, symbol)
			}
		}
	}

	return unused, nil
}

func (sa *SemanticAnalyzer) detectCircularDependencies(projectID int64) []*types.CircularDependency {
	circular := make([]*types.CircularDependency, 0)

	files, err := sa.db.GetAllFilesForProject(projectID)
	if err != nil {
		return circular
	}

	// Build adjacency list
	graph := make(map[int64][]int64)
	for _, file := range files {
		imports, err := sa.db.GetImportsByFile(file.ID)
		if err != nil {
			continue
		}

		for _, imp := range imports {
			targetFiles, _ := sa.findFilesByImport(imp.Source, projectID)
			for _, target := range targetFiles {
				graph[file.ID] = append(graph[file.ID], target.ID)
			}
		}
	}

	// Detect cycles using DFS
	visited := make(map[int64]bool)
	recStack := make(map[int64]bool)

	var dfs func(fileID int64, path []int64) []int64
	dfs = func(fileID int64, path []int64) []int64 {
		visited[fileID] = true
		recStack[fileID] = true
		path = append(path, fileID)

		for _, neighbor := range graph[fileID] {
			if !visited[neighbor] {
				if cycle := dfs(neighbor, path); cycle != nil {
					return cycle
				}
			} else if recStack[neighbor] {
				// Found cycle
				cycleStart := -1
				for i, id := range path {
					if id == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					return path[cycleStart:]
				}
			}
		}

		recStack[fileID] = false
		return nil
	}

	for _, file := range files {
		if !visited[file.ID] {
			if cycle := dfs(file.ID, make([]int64, 0)); cycle != nil {
				// Convert file IDs to file paths
				cyclePaths := make([]string, 0)
				for _, fileID := range cycle {
					for _, f := range files {
						if f.ID == fileID {
							cyclePaths = append(cyclePaths, f.Path)
							break
						}
					}
				}

				circular = append(circular, &types.CircularDependency{
					Files:       cyclePaths,
					Description: fmt.Sprintf("Circular dependency detected: %s", strings.Join(cyclePaths, " -> ")),
					Severity:    "warning",
				})
			}
		}
	}

	return circular
}

func (sa *SemanticAnalyzer) calculateQualityScore(result *types.SemanticAnalysisResult) float64 {
	// Calculate semantic quality score (0-100)
	totalFiles := result.Metrics["total_files"].(int)
	if totalFiles == 0 {
		return 100.0
	}

	// Penalties
	typeErrorPenalty := float64(len(result.TypeErrors)) * 5.0
	undefinedPenalty := float64(len(result.UndefinedReferences)) * 10.0
	unusedPenalty := float64(len(result.UnusedSymbols)) * 1.0
	circularPenalty := float64(len(result.CircularDeps)) * 15.0

	totalPenalty := typeErrorPenalty + undefinedPenalty + unusedPenalty + circularPenalty

	score := 100.0 - (totalPenalty / float64(totalFiles))
	if score < 0 {
		score = 0
	}

	return score
}

// AnalyzeCallGraph builds a call graph for the project
func (sa *SemanticAnalyzer) AnalyzeCallGraph(projectID int64) (*types.CallGraph, error) {
	callGraph := &types.CallGraph{
		Nodes: make([]*types.CallGraphNode, 0),
		Edges: make([]*types.CallGraphEdge, 0),
	}

	files, err := sa.db.GetAllFilesForProject(projectID)
	if err != nil {
		return nil, err
	}

	// Build nodes (all functions/methods)
	symbolToNode := make(map[int64]*types.CallGraphNode)

	for _, file := range files {
		symbols, err := sa.db.GetSymbolsByFile(file.ID)
		if err != nil {
			continue
		}

		for _, symbol := range symbols {
			if symbol.Type == types.SymbolTypeFunction || symbol.Type == types.SymbolTypeMethod {
				node := &types.CallGraphNode{
					SymbolID:   symbol.ID,
					SymbolName: symbol.Name,
					FilePath:   file.Path,
					CallCount:  0,
				}
				callGraph.Nodes = append(callGraph.Nodes, node)
				symbolToNode[symbol.ID] = node
			}
		}
	}

	// Build edges (function calls)
	for _, file := range files {
		symbols, err := sa.db.GetSymbolsByFile(file.ID)
		if err != nil {
			continue
		}

		for _, symbol := range symbols {
			relationships, err := sa.db.GetRelationshipsForSymbol(symbol.ID)
			if err != nil {
				continue
			}

			for _, rel := range relationships {
				if rel.Type == types.RelationshipTypeCalls {
					edge := &types.CallGraphEdge{
						FromSymbolID: symbol.ID,
						ToSymbolID:   rel.ToSymbolID,
						CallSites:    1, // Simplified
					}
					callGraph.Edges = append(callGraph.Edges, edge)

					// Update call count
					if node, ok := symbolToNode[rel.ToSymbolID]; ok {
						node.CallCount++
					}
				}
			}
		}
	}

	return callGraph, nil
}
