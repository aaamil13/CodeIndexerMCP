package ai

import (
	"fmt"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// ImpactAnalyzer analyzes the impact of code changes
type ImpactAnalyzer struct {
	db *database.DB
}

// NewImpactAnalyzer creates a new impact analyzer
func NewImpactAnalyzer(db *database.DB) *ImpactAnalyzer {
	return &ImpactAnalyzer{db: db}
}

// AnalyzeChangeImpact analyzes the impact of changing a symbol
func (ia *ImpactAnalyzer) AnalyzeChangeImpact(symbolName string) (*types.ChangeImpact, error) {
	// Get the symbol
	symbol, err := ia.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	// Get direct references
	references, err := ia.db.GetReferencesBySymbol(symbol.ID)
	if err != nil {
		return nil, err
	}

	directReferences := len(references)

	// Get affected files
	affectedFilesMap := make(map[int64]*types.File)
	for _, ref := range references {
		if _, exists := affectedFilesMap[ref.FileID]; !exists {
			if file, err := ia.db.GetFile(ref.FileID); err == nil {
				affectedFilesMap[ref.FileID] = file
			}
		}
	}

	affectedFiles := []*types.File{}
	for _, file := range affectedFilesMap {
		affectedFiles = append(affectedFiles, file)
	}

	// Get affected symbols (symbols that reference this one)
	affectedSymbols := []*types.Symbol{}
	for _, ref := range references {
		symbols, err := ia.db.GetSymbolsByFile(ref.FileID)
		if err != nil {
			continue
		}
		for _, sym := range symbols {
			// Check if this symbol contains the reference
			if sym.StartLine <= ref.LineNumber && sym.EndLine >= ref.LineNumber {
				affectedSymbols = append(affectedSymbols, sym)
				break
			}
		}
	}

	// Determine risk level
	riskLevel := ia.calculateRiskLevel(directReferences, len(affectedFiles), symbol)

	// Generate suggestions
	suggestions := ia.generateSuggestions(symbol, directReferences, affectedFiles)

	// Check if this would be a breaking change
	breakingChanges := ia.isBreakingChange(symbol, directReferences)

	// Calculate indirect references (transitive)
	indirectReferences := ia.calculateIndirectReferences(affectedSymbols)

	return &types.ChangeImpact{
		Symbol:             symbol,
		DirectReferences:   directReferences,
		IndirectReferences: indirectReferences,
		AffectedFiles:      affectedFiles,
		AffectedSymbols:    affectedSymbols,
		RiskLevel:          riskLevel,
		Suggestions:        suggestions,
		BreakingChanges:    breakingChanges,
	}, nil
}

// calculateRiskLevel determines the risk level of a change
func (ia *ImpactAnalyzer) calculateRiskLevel(directRefs, affectedFiles int, symbol *types.Symbol) string {
	// High risk criteria
	if directRefs > 50 || affectedFiles > 20 {
		return "high"
	}

	// Public/exported symbols are higher risk
	if symbol.IsExported {
		if directRefs > 20 || affectedFiles > 10 {
			return "high"
		}
		if directRefs > 5 || affectedFiles > 3 {
			return "medium"
		}
	}

	// Medium risk
	if directRefs > 10 || affectedFiles > 5 {
		return "medium"
	}

	// Low risk
	return "low"
}

// generateSuggestions generates refactoring suggestions
func (ia *ImpactAnalyzer) generateSuggestions(symbol *types.Symbol, directRefs int, affectedFiles []*types.File) []string {
	suggestions := []string{}

	if directRefs > 50 {
		suggestions = append(suggestions, "Consider deprecation period before removal")
		suggestions = append(suggestions, "Add migration guide for users")
		suggestions = append(suggestions, "Create wrapper function for backward compatibility")
	}

	if symbol.IsExported && directRefs > 10 {
		suggestions = append(suggestions, "Mark as deprecated first with @deprecated tag")
		suggestions = append(suggestions, "Update all examples in documentation")
	}

	if len(affectedFiles) > 10 {
		suggestions = append(suggestions, "Use automated refactoring tools")
		suggestions = append(suggestions, "Create comprehensive tests before refactoring")
		suggestions = append(suggestions, "Refactor in stages across multiple PRs")
	}

	if directRefs > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Update %d references across %d files", directRefs, len(affectedFiles)))
	}

	if symbol.Type == types.SymbolTypeInterface || symbol.Type == types.SymbolTypeClass {
		suggestions = append(suggestions, "Check all implementations/subclasses")
	}

	return suggestions
}

// isBreakingChange checks if a change would be breaking
func (ia *ImpactAnalyzer) isBreakingChange(symbol *types.Symbol, directRefs int) bool {
	// Exported symbols with references are breaking changes
	if symbol.IsExported && directRefs > 0 {
		return true
	}

	// Public API methods are breaking changes
	if symbol.Visibility == types.VisibilityPublic && symbol.Type == types.SymbolTypeMethod {
		return true
	}

	return false
}

// calculateIndirectReferences calculates transitive references
func (ia *ImpactAnalyzer) calculateIndirectReferences(affectedSymbols []*types.Symbol) int {
	count := 0
	seen := make(map[int64]bool)

	for _, symbol := range affectedSymbols {
		if seen[symbol.ID] {
			continue
		}
		seen[symbol.ID] = true

		// Count references to this affected symbol
		if refs, err := ia.db.GetReferencesBySymbol(symbol.ID); err == nil {
			count += len(refs)
		}
	}

	return count
}

// AnalyzeBulkImpact analyzes impact of changing multiple symbols
func (ia *ImpactAnalyzer) AnalyzeBulkImpact(symbolNames []string) (map[string]*types.ChangeImpact, error) {
	impacts := make(map[string]*types.ChangeImpact)

	for _, name := range symbolNames {
		impact, err := ia.AnalyzeChangeImpact(name)
		if err != nil {
			continue // Skip errors
		}
		impacts[name] = impact
	}

	return impacts, nil
}

// SuggestRefactorings suggests refactoring opportunities based on impact analysis
func (ia *ImpactAnalyzer) SuggestRefactorings(symbolName string) ([]*types.RefactoringOpportunity, error) {
	impact, err := ia.AnalyzeChangeImpact(symbolName)
	if err != nil {
		return nil, err
	}

	opportunities := []*types.RefactoringOpportunity{}

	// High usage but low visibility - should be more visible
	if impact.DirectReferences > 20 && impact.Symbol.Visibility == types.VisibilityPrivate {
		opportunities = append(opportunities, &types.RefactoringOpportunity{
			Type:        "increase_visibility",
			Symbol:      impact.Symbol,
			Description: "Consider making this symbol public - it's heavily used",
			Reason:      fmt.Sprintf("Used %d times but marked as private", impact.DirectReferences),
			Impact:      "medium",
			Effort:      "low",
			Benefits:    []string{"Better API surface", "More discoverable"},
			Risks:       []string{"Increased API surface to maintain"},
		})
	}

	// Very high usage - consider splitting
	if impact.DirectReferences > 100 {
		opportunities = append(opportunities, &types.RefactoringOpportunity{
			Type:        "extract_interface",
			Symbol:      impact.Symbol,
			Description: "Consider extracting interface - very high usage",
			Reason:      fmt.Sprintf("Used %d times - hard to change", impact.DirectReferences),
			Impact:      "high",
			Effort:      "high",
			Benefits:    []string{"Better abstraction", "Easier to test", "More flexible"},
			Risks:       []string{"More complex codebase", "Migration effort"},
		})
	}

	// Spread across many files - might need better organization
	if len(impact.AffectedFiles) > 30 {
		opportunities = append(opportunities, &types.RefactoringOpportunity{
			Type:        "consolidate_usage",
			Symbol:      impact.Symbol,
			Description: "Usage spread across too many files",
			Reason:      fmt.Sprintf("Used in %d files - might indicate coupling", len(impact.AffectedFiles)),
			Impact:      "medium",
			Effort:      "high",
			Benefits:    []string{"Reduced coupling", "Better modularity"},
			Risks:       []string{"Large refactoring effort"},
		})
	}

	return opportunities, nil
}
