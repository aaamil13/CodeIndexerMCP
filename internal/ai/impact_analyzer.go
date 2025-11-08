package ai

import (
	"fmt"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

// ImpactAnalyzer analyzes the impact of code changes
type ImpactAnalyzer struct {
	db *database.Manager
}

// NewImpactAnalyzer creates a new impact analyzer
func NewImpactAnalyzer(db *database.Manager) *ImpactAnalyzer {
	return &ImpactAnalyzer{db: db}
}

// AnalyzeChangeImpact analyzes the impact of changing a symbol
func (ia *ImpactAnalyzer) AnalyzeChangeImpact(symbolName string) (*model.ChangeImpact, error) {
	// TODO: Implement after DB methods are available
	// // Get the symbol
	// symbol, err := ia.db.GetSymbolByName(symbolName)
	// if err != nil {
	// 	return nil, err
	// }
	// if symbol == nil {
	// 	return nil, fmt.Errorf("symbol not found: %s", symbolName)
	// }

	// // Get direct references
	// references, err := ia.db.GetReferencesBySymbol(symbol.ID)
	// if err != nil {
	// 	return nil, err
	// }

	// directReferences := len(references)

	// // Get affected files
	// affectedFilesMap := make(map[string]bool)
	// for _, ref := range references {
	// 	affectedFilesMap[ref.FilePath] = true
	// }

	// affectedFiles := []string{}
	// for file := range affectedFilesMap {
	// 	affectedFiles = append(affectedFiles, file)
	// }

	// // Get affected symbols (symbols that reference this one)
	// affectedSymbols := []*model.Symbol{}
	// for _, ref := range references {
	// 	symbols, err := ia.db.GetSymbolsByFile(ref.FilePath)
	// 	if err != nil {
	// 		continue
	// 	}
	// 	for _, sym := range symbols {
	// 		// Check if this symbol contains the reference
	// 		if sym.Range.Start.Line <= ref.Line && sym.Range.End.Line >= ref.Line {
	// 			affectedSymbols = append(affectedSymbols, sym)
	// 			break
	// 		}
	// 	}
	// }

	// // Determine risk level
	// riskLevel := ia.calculateRiskLevel(directReferences, len(affectedFiles), symbol)

	// // Generate suggestions
	// suggestions := ia.generateSuggestions(symbol, directReferences, affectedFiles)

	// // Check if this would be a breaking change
	// breakingChanges := ia.isBreakingChange(symbol, directReferences)

	// // Calculate indirect references (transitive)
	// indirectReferences := ia.calculateIndirectReferences(affectedSymbols)

	// return &ChangeImpact{
	// 	Symbol:             symbol,
	// 	DirectReferences:   directReferences,
	// 	IndirectReferences: indirectReferences,
	// 	AffectedFiles:      affectedFiles,
	// 	AffectedSymbols:    affectedSymbols,
	// 	RiskLevel:          riskLevel,
	// 	Suggestions:        suggestions,
	// 	BreakingChanges:    breakingChanges,
	// }, nil
	return nil, fmt.Errorf("not implemented")
}

// calculateRiskLevel determines the risk level of a change
func (ia *ImpactAnalyzer) calculateRiskLevel(directRefs, affectedFiles int, symbol *model.Symbol) string {
	// High risk criteria
	if directRefs > 50 || affectedFiles > 20 {
		return "high"
	}

	// Public/exported symbols are higher risk
	isExported := strings.ToUpper(symbol.Name[0:1]) == symbol.Name[0:1]
	if isExported {
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
func (ia *ImpactAnalyzer) generateSuggestions(symbol *model.Symbol, directRefs int, affectedFiles []string) []string {
	suggestions := []string{}

	if directRefs > 50 {
		suggestions = append(suggestions, "Consider deprecation period before removal")
		suggestions = append(suggestions, "Add migration guide for users")
		suggestions = append(suggestions, "Create wrapper function for backward compatibility")
	}

	isExported := strings.ToUpper(symbol.Name[0:1]) == symbol.Name[0:1]
	if isExported && directRefs > 10 {
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

	// TODO: Check symbol kind for interface/class
	// if symbol.Type == types.SymbolTypeInterface || symbol.Type == types.SymbolTypeClass {
	// 	suggestions = append(suggestions, "Check all implementations/subclasses")
	// }

	return suggestions
}

// isBreakingChange checks if a change would be breaking
func (ia *ImpactAnalyzer) isBreakingChange(symbol *model.Symbol, directRefs int) bool {
	// Exported symbols with references are breaking changes
	isExported := strings.ToUpper(symbol.Name[0:1]) == symbol.Name[0:1]
	if isExported && directRefs > 0 {
		return true
	}

	// Public API methods are breaking changes
	// TODO: Check symbol kind for method and visibility
	// if symbol.Visibility == types.VisibilityPublic && symbol.Type == types.SymbolTypeMethod {
	// 	return true
	// }

	return false
}

// calculateIndirectReferences calculates transitive references
func (ia *ImpactAnalyzer) calculateIndirectReferences(affectedSymbols []*model.Symbol) int {
	// TODO: Implement after DB methods are available
	// count := 0
	// seen := make(map[string]bool)

	// for _, symbol := range affectedSymbols {
	// 	if seen[symbol.ID] {
	// 		continue
	// 	}
	// 	seen[symbol.ID] = true

	// 	// Count references to this affected symbol
	// 	if refs, err := ia.db.GetReferencesBySymbol(symbol.ID); err == nil {
	// 		count += len(refs)
	// 	}
	// }

	// return count
	return 0
}

// AnalyzeBulkImpact analyzes impact of changing multiple symbols
func (ia *ImpactAnalyzer) AnalyzeBulkImpact(symbolNames []string) (map[string]*ChangeImpact, error) {
	// TODO: Implement after DB methods are available
	// impacts := make(map[string]*ChangeImpact)

	// for _, name := range symbolNames {
	// 	impact, err := ia.AnalyzeChangeImpact(name)
	// 	if err != nil {
	// 		continue // Skip errors
	// 	}
	// 	impacts[name] = impact
	// }

	// return impacts, nil
	return nil, fmt.Errorf("not implemented")
}

// SuggestRefactorings suggests refactoring opportunities based on impact analysis
func (ia *ImpactAnalyzer) SuggestRefactorings(symbolName string) ([]*model.RefactoringOpportunity, error) {
	// TODO: Implement after DB methods are available
	// impact, err := ia.AnalyzeChangeImpact(symbolName)
	// if err != nil {
	// 	return nil, err
	// }

	// opportunities := []*model.RefactoringOpportunity{}

	// // High usage but low visibility - should be more visible
	// if impact.DirectReferences > 20 && impact.Symbol.Visibility == types.VisibilityPrivate {
	// 	opportunities = append(opportunities, &model.RefactoringOpportunity{
	// 		Type:        "increase_visibility",
	// 		Symbol:      impact.Symbol,
	// 		Description: "Consider making this symbol public - it's heavily used",
	// 		Reason:      fmt.Sprintf("Used %d times but marked as private", impact.DirectReferences),
	// 		Impact:      "medium",
	// 		Effort:      "low",
	// 		Benefits:    []string{"Better API surface", "More discoverable"},
	// 		Risks:       []string{"Increased API surface to maintain"},
	// 	})
	// }

	// // Very high usage - consider splitting
	// if impact.DirectReferences > 100 {
	// 	opportunities = append(opportunities, &model.RefactoringOpportunity{
	// 		Type:        "extract_interface",
	// 		Symbol:      impact.Symbol,
	// 		Description: "Consider extracting interface - very high usage",
	// 		Reason:      fmt.Sprintf("Used %d times - hard to change", impact.DirectReferences),
	// 		Impact:      "high",
	// 		Effort:      "high",
	// 		Benefits:    []string{"Better abstraction", "Easier to test", "More flexible"},
	// 		Risks:       []string{"More complex codebase", "Migration effort"},
	// 	})
	// }

	// // Spread across many files - might need better organization
	// if len(impact.AffectedFiles) > 30 {
	// 	opportunities = append(opportunities, &model.RefactoringOpportunity{
	// 		Type:        "consolidate_usage",
	// 		Symbol:      impact.Symbol,
	// 		Description: "Usage spread across too many files",
	// 		Reason:      fmt.Sprintf("Used in %d files - might indicate coupling", len(impact.AffectedFiles)),
	// 		Impact:      "medium",
	// 		Effort:      "high",
	// 		Benefits:    []string{"Reduced coupling", "Better modularity"},
	// 		Risks:       []string{"Large refactoring effort"},
	// 	})
	// }

	// return opportunities, nil
	return nil, fmt.Errorf("not implemented")
}