package ai

import (
	"fmt"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

// SemanticAnalyzer performs semantic analysis across files
type SemanticAnalyzer struct {
	db            *database.Manager
	typeValidator *TypeValidator
}

// NewSemanticAnalyzer creates a new semantic analyzer
func NewSemanticAnalyzer(db *database.Manager) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		db:            db,
		typeValidator: NewTypeValidator(db),
	}
}

// AnalyzeProject performs full semantic analysis on a project
func (sa *SemanticAnalyzer) AnalyzeProject(projectID string) (*model.SemanticAnalysisResult, error) {
	result := &model.SemanticAnalysisResult{
		ProjectID:           projectID,
		TypeErrors:          make([]*model.TypeMismatch, 0),
		UndefinedReferences: make([]*model.UndefinedUsage, 0),
		UnusedSymbols:       make([]*model.Symbol, 0),
		CircularDeps:        make([]*model.CircularDependency, 0),
		Warnings:            make([]string, 0),
		Metrics:             make(map[string]interface{}),
	}

	// Convert projectID to int (assuming projectID in DB is int)
	projID, err := strconv.Atoi(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	// Get all files in project
	files, err := sa.db.GetAllFilesForProject(projID)
	if err != nil {
		return nil, fmt.Errorf("failed to get files for project %s: %w", projectID, err)
	}

	// Analyze each file
	for _, file := range files {
		// Type validation (assuming typeValidator.ValidateFile accepts file path)
		validation, err := sa.typeValidator.ValidateFile(file.Path)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to validate file %s: %v", file.Path, err))
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
	} else {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to find unused symbols: %v", err))
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
func (sa *SemanticAnalyzer) InferType(symbolID string) (*model.TypeInference, error) {
	symbol, err := sa.db.GetSymbol(symbolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol %s: %w", symbolID, err)
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolID)
	}

	inference := &model.TypeInference{
		SymbolName: symbol.Name,
		Confidence: 0.5, // Default confidence
	}

	// Type inference based on symbol type
	switch symbol.Kind {
	case model.SymbolKindFunction, model.SymbolKindMethod:
		// Need to fetch full function details for return type
		fn, err := sa.db.GetFunctionDetails(symbolID)
		if err != nil {
			return nil, fmt.Errorf("failed to get function details for symbol %s: %w", symbolID, err)
		}
		if fn != nil && fn.ReturnType != "" {
			inference.InferredType = fn.ReturnType
		} else {
			inference.InferredType = sa.inferFunctionType(symbol)
		}
		inference.Confidence = 0.8
		inference.Reasoning = "Inferred from function signature and/or database"

	// For other kinds, infer from signature or name.
	// For variables, more advanced analysis (e.g., from assignment or usage) would be needed.
	// For simplicity, we'll use a basic inference for now.
	case model.SymbolKindVariable, model.SymbolKindParameter, model.SymbolKindField:
		if symbol.Type != "" {
			inference.InferredType = symbol.Type
			inference.Confidence = 0.9
			inference.Reasoning = "Directly from symbol type information"
		} else {
			inference.InferredType = sa.inferVariableType(symbol)
			inference.Confidence = 0.6
			inference.Reasoning = "Inferred from usage context or basic signature parsing"
		}

	case model.SymbolKindClass, model.SymbolKindInterface, model.SymbolKindStruct, model.SymbolKindEnum:
		inference.InferredType = symbol.Name // Class/Interface name is its own type
		inference.Confidence = 1.0
		inference.Reasoning = "From definition"

	default:
		inference.InferredType = "unknown"
		inference.Confidence = 0.0
		inference.Reasoning = "Unable to infer type"
	}

	return inference, nil
}

func (sa *SemanticAnalyzer) inferFunctionType(symbol *model.Symbol) string {
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

func (sa *SemanticAnalyzer) inferVariableType(symbol *model.Symbol) string {
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

// AnalyzeProject performs full semantic analysis on a project
func (sa *SemanticAnalyzer) AnalyzeProject(projectID string) (*model.SemanticAnalysisResult, error) {
	result := &model.SemanticAnalysisResult{
		ProjectID:           projectID,
		TypeErrors:          make([]*model.TypeMismatch, 0),
		UndefinedReferences: make([]*model.UndefinedUsage, 0),
		UnusedSymbols:       make([]*model.Symbol, 0),
		CircularDeps:        make([]*model.CircularDependency, 0),
		Warnings:            make([]string, 0),
		Metrics:             make(map[string]interface{}),
	}

	// Convert projectID to int (assuming projectID in DB is int)
	projID, err := strconv.Atoi(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	// Get all files in project
	files, err := sa.db.GetAllFilesForProject(projID)
	if err != nil {
		return nil, fmt.Errorf("failed to get files for project %s: %w", projectID, err)
	}

	// Analyze each file
	for _, file := range files {
		// Type validation (assuming typeValidator.ValidateFile accepts file path)
		validation, err := sa.typeValidator.ValidateFile(file.Path)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to validate file %s: %v", file.Path, err))
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
	} else {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to find unused symbols: %v", err))
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
func (sa *SemanticAnalyzer) InferType(symbolID string) (*model.TypeInference, error) {
	symbol, err := sa.db.GetSymbol(symbolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol %s: %w", symbolID, err)
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolID)
	}

	inference := &model.TypeInference{
		SymbolName: symbol.Name,
		Confidence: 0.5, // Default confidence
	}

	// Type inference based on symbol type
	switch symbol.Kind {
	case model.SymbolKindFunction, model.SymbolKindMethod:
		// Need to fetch full function details for return type
		fn, err := sa.db.GetFunctionDetails(symbolID)
		if err != nil {
			return nil, fmt.Errorf("failed to get function details for symbol %s: %w", symbolID, err)
		}
		if fn != nil && fn.ReturnType != "" {
			inference.InferredType = fn.ReturnType
		} else {
			inference.InferredType = sa.inferFunctionType(symbol)
		}
		inference.Confidence = 0.8
		inference.Reasoning = "Inferred from function signature and/or database"

	// For other kinds, infer from signature or name.
	// For variables, more advanced analysis (e.g., from assignment or usage) would be needed.
	// For simplicity, we'll use a basic inference for now.
	case model.SymbolKindVariable, model.SymbolKindParameter, model.SymbolKindField:
		if symbol.Type != "" {
			inference.InferredType = symbol.Type
			inference.Confidence = 0.9
			inference.Reasoning = "Directly from symbol type information"
		} else {
			inference.InferredType = sa.inferVariableType(symbol)
			inference.Confidence = 0.6
			inference.Reasoning = "Inferred from usage context or basic signature parsing"
		}

	case model.SymbolKindClass, model.SymbolKindInterface, model.SymbolKindStruct, model.SymbolKindEnum:
		inference.InferredType = symbol.Name // Class/Interface name is its own type
		inference.Confidence = 1.0
		inference.Reasoning = "From definition"

	default:
		inference.InferredType = "unknown"
		inference.Confidence = 0.0
		inference.Reasoning = "Unable to infer type"
	}

	return inference, nil
}

func (sa *SemanticAnalyzer) inferFunctionType(symbol *model.Symbol) string {
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

func (sa *SemanticAnalyzer) inferVariableType(symbol *model.Symbol) string {
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
func (sa *SemanticAnalyzer) ResolveCrossFileReference(symbolName string, projectID string) ([]*model.Symbol, error) {
	// Search for symbol across all files in project
	searchOptions := model.SearchOptions{
		Query: symbolName,
		// Assuming SearchSymbols can filter by projectID internally or that symbols are unique enough
		// For now, projectID is not directly used in SearchSymbols signature from manager.go
		// We would need to add projectID to SearchOptions and modify SearchSymbols in database/manager.go
	}
	symbols, err := sa.db.SearchSymbols(searchOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to search symbols for %s: %w", symbolName, err)
	}

	// Filter by project ID if SearchSymbols doesn't support it directly
	filteredSymbols := []*model.Symbol{}
	// This would be more efficient if db.SearchSymbols took projectID
	for _, s := range symbols {
		// Need to get the file associated with the symbol to check its projectID
		// This requires a GetFileByPath or similar in dbManager that returns file with ProjectID
		// For now, assuming direct projectID is not enforced in SearchSymbols
		// As a workaround, we'd need to fetch files for the project and then match symbol files.
		// Skipping precise projectID filtering for now due to complexity without proper db methods.
		filteredSymbols = append(filteredSymbols, s)
	}

	return filteredSymbols, nil
}

func (sa *SemanticAnalyzer) analyzeCrossDependencies(projectID string, result *model.SemanticAnalysisResult) {
	projID, err := strconv.Atoi(projectID)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Invalid project ID for cross-dependency analysis: %v", err))
		return
	}

	// Get all files in project
	files, err := sa.db.GetAllFilesForProject(projID)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to get files for cross-dependency analysis: %v", err))
		return
	}

	// Build dependency map
	dependencyMap := make(map[string][]string) // file_path -> []dependent_file_paths

	for _, file := range files {
		imports, err := sa.db.GetImportsByFile(file.Path)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to get imports for file %s: %v", file.Path, err))
			continue
		}

		// For each import, try to find the corresponding file
		for _, imp := range imports {
			targetFiles, err := sa.findFilesByImport(imp.Path, projectID)
			if err != nil {
				// Log warning but continue
				result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to find files for import %s in file %s: %v", imp.Path, file.Path, err))
				continue
			}

			for _, targetFile := range targetFiles {
				dependencyMap[file.Path] = append(dependencyMap[file.Path], targetFile)
			}
		}
	}

	// Store dependency information in result
	result.Metrics["dependency_map"] = dependencyMap
}

func (sa *SemanticAnalyzer) findFilesByImport(importPath string, projectID string) ([]string, error) {
	// Simplified import resolution
	// In production, this would handle:
	// - Relative imports (./module, ../module)
	// - Absolute imports (package/module)
	// - Node modules (@scope/package)
	// - Python packages (django.db)

	projID, err := strconv.Atoi(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	files, err := sa.db.GetAllFilesForProject(projID)
	if err != nil {
		return nil, err
	}

	matching := make([]string, 0)
	for _, file := range files {
		// Simple name matching - could be improved with proper module resolution
		if strings.Contains(file.Path, importPath) || strings.Contains(file.RelativePath, importPath) {
			matching = append(matching, file.Path)
		}
	}

	return matching, nil
}

func (sa *SemanticAnalyzer) findUnusedSymbols(projectID string) ([]*model.Symbol, error) {
	unused := make([]*model.Symbol, 0)

	projID, err := strconv.Atoi(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	files, err := sa.db.GetAllFilesForProject(projID)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		symbols, err := sa.db.GetSymbolsByFile(file.Path)
		if err != nil {
			fmt.Printf("Warning: Failed to get symbols for file %s: %v\n", file.Path, err)
			continue
		}

		for _, symbol := range symbols {
			// Skip exported symbols (might be used externally)
			if strings.ToUpper(symbol.Name[0:1]) == symbol.Name[0:1] { // Check for exported
				continue
			}

			// Check if symbol has any references
			refs, err := sa.db.GetReferencesBySymbol(symbol.ID)
			if err != nil {
				fmt.Printf("Warning: Failed to get references for symbol %s: %v\n", symbol.ID, err)
				continue
			}

			// Check if symbol is referenced by itself (e.g., function calling itself)
			// or if it's referenced by anything else
			isUsed := false
			for _, ref := range refs {
				if ref.SourceSymbolID != symbol.ID { // If referenced by something other than itself
					isUsed = true
					break
				}
			}

			if !isUsed {
				unused = append(unused, symbol)
			}
		}
	}

	return unused, nil
}

func (sa *SemanticAnalyzer) detectCircularDependencies(projectID string) []*model.CircularDependency {
	circular := make([]*model.CircularDependency, 0)

	projID, err := strconv.Atoi(projectID)
	if err != nil {
		fmt.Printf("Warning: Invalid project ID for circular dependency detection: %v\n", err)
		return circular
	}

	files, err := sa.db.GetAllFilesForProject(projID)
	if err != nil {
		fmt.Printf("Warning: Failed to get files for circular dependency detection: %v\n", err)
		return circular
	}

	// Build adjacency list (filePath -> []dependentFilePaths)
	graph := make(map[string][]string)
	for _, file := range files {
		imports, err := sa.db.GetImportsByFile(file.Path)
		if err != nil {
			fmt.Printf("Warning: Failed to get imports for file %s for circular dependency detection: %v\n", file.Path, err)
			continue
		}

		for _, imp := range imports {
			targetFiles, err := sa.findFilesByImport(imp.Path, projectID)
			if err != nil {
				fmt.Printf("Warning: Failed to find files for import %s in file %s: %v\n", imp.Path, file.Path, err)
				continue
			}

			for _, target := range targetFiles {
				graph[file.Path] = append(graph[file.Path], target)
			}
		}
	}

	// Detect cycles using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(filePath string, path []string) []string
	dfs = func(filePath string, path []string) []string {
		visited[filePath] = true
		recStack[filePath] = true
		path = append(path, filePath)

		for _, neighbor := range graph[filePath] {
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

		recStack[filePath] = false
		return nil
	}

	for _, file := range files {
		if !visited[file.Path] {
			if cycle := dfs(file.Path, make([]string, 0)); cycle != nil {
				circular = append(circular, &model.CircularDependency{
					Files:       cycle,
					Description: fmt.Sprintf("Circular dependency detected: %s", strings.Join(cycle, " -> ")),
					Severity:    "warning",
				})
			}
		}
	}

	return circular
}

func (sa *SemanticAnalyzer) calculateQualityScore(result *model.SemanticAnalysisResult) float64 {
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
func (sa *SemanticAnalyzer) AnalyzeCallGraph(projectID string) (*model.CallGraph, error) {
	callGraph := &model.CallGraph{
		Nodes: make([]*model.CallGraphNode, 0),
		Edges: make([]*model.CallGraphEdge, 0),
	}

	projID, err := strconv.Atoi(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	files, err := sa.db.GetAllFilesForProject(projID)
	if err != nil {
		return nil, fmt.Errorf("failed to get files for call graph analysis: %w", err)
	}

	// Build nodes (all functions/methods)
	symbolToNode := make(map[string]*model.CallGraphNode)

	for _, file := range files {
		symbols, err := sa.db.GetSymbolsByFile(file.Path)
		if err != nil {
			fmt.Printf("Warning: Failed to get symbols for file %s for call graph: %v\n", file.Path, err)
			continue
		}

		for _, symbol := range symbols {
			if symbol.Kind == model.SymbolKindFunction || symbol.Kind == model.SymbolKindMethod {
				node := &model.CallGraphNode{
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
		symbols, err := sa.db.GetSymbolsByFile(file.Path)
		if err != nil {
			continue
		}

		for _, symbol := range symbols {
			references, err := sa.db.GetReferencesBySymbol(symbol.ID)
			if err != nil {
				fmt.Printf("Warning: Failed to get references for symbol %s for call graph: %v\n", symbol.ID, err)
				continue
			}

			for _, ref := range references {
				if ref.ReferenceType == model.ReferenceTypeCalls { // Assuming ReferenceTypeCalls is defined
					edge := &model.CallGraphEdge{
						FromSymbolID: symbol.ID,
						ToSymbolID:   ref.TargetSymbolName, // TargetSymbolName might be ID or name, depending on how GetReferencesBySymbol works
						CallSites:    1,                    // Simplified
					}
					callGraph.Edges = append(callGraph.Edges, edge)

					// Update call count
					if node, ok := symbolToNode[ref.TargetSymbolName]; ok { // Use TargetSymbolName to find node
						node.CallCount++
					}
				}
			}
		}
	}

	return callGraph, nil
}

</final_file_content>

IMPORTANT: For any future changes to this file, use the final_file_content shown above as your reference. This content reflects the current state of the file, including any auto-formatting (e.g., if you used single quotes but the formatter converted them to double quotes). Always base your SEARCH/REPLACE operations on this final version to ensure accuracy.

<environment_details>
# Visual Studio Code Visible Files
internal/ai/semantic_analyzer.go

# Visual Studio Code Open Tabs
NewPlan.md
internal/core/indexer.go
internal/core/indexer_test.go
internal/core/watcher.go
internal/parsing/grammar_manager_test.go
internal/parsing/ast_provider_test.go
internal/database/schema.sql
internal/database/sqlite_fts.go
go.mod
internal/database/manager_test.go
internal/parsing/extractors/python_extractor.go
internal/database/manager.go
internal/ai/code_generator.go
internal/ai/status_tracker.go
internal/ai/change_tracker.go
internal/ai/context_extractor.go
internal/ai/dependency_graph.go
internal/ai/impact_analyzer.go
internal/ai/metrics.go
internal/ai/semantic_analyzer.go

# Recently Modified Files
These files have been modified since you last accessed them (file was just edited so you may need to re-read it before editing):
internal/ai/semantic_analyzer.go

# Current Time
11/9/2025, 6:40:24 PM (Europe/Sofia, UTC+2:00)

# Context Window Usage
260,939 / 550K tokens used (47%)

# Current Mode
ACT MODE
