package ai

import (
	"fmt"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// UsageAnalyzer analyzes symbol usage patterns
type UsageAnalyzer struct {
	db *database.DB
}

// NewUsageAnalyzer creates a new usage analyzer
func NewUsageAnalyzer(db *database.DB) *UsageAnalyzer {
	return &UsageAnalyzer{db: db}
}

// AnalyzeUsage analyzes usage statistics for a symbol
func (ua *UsageAnalyzer) AnalyzeUsage(symbolName string) (*types.SymbolUsageStats, error) {
	// Get the symbol
	symbol, err := ua.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	// Get all references
	references, err := ua.db.GetReferencesBySymbol(symbol.ID)
	if err != nil {
		return nil, err
	}

	usageCount := len(references)

	// Group by file
	usageByFile := make(map[string]int)
	fileSet := make(map[int64]bool)

	for _, ref := range references {
		fileSet[ref.FileID] = true

		// Get file for path
		if file, err := ua.db.GetFile(ref.FileID); err == nil {
			usageByFile[file.RelativePath]++
		}
	}

	fileCount := len(fileSet)

	// Detect common usage patterns
	commonPatterns := ua.detectUsagePatterns(symbol, references)

	// Check if deprecated
	isDeprecated := ua.isDeprecated(symbol)

	// Find alternatives
	alternatives := ua.findAlternatives(symbol)

	return &types.SymbolUsageStats{
		Symbol:         symbol,
		UsageCount:     usageCount,
		FileCount:      fileCount,
		UsageByFile:    usageByFile,
		CommonPatterns: commonPatterns,
		IsDeprecated:   isDeprecated,
		Alternatives:   alternatives,
	}, nil
}

// detectUsagePatterns detects common usage patterns
func (ua *UsageAnalyzer) detectUsagePatterns(symbol *types.Symbol, references []*types.Reference) []string {
	patterns := []string{}

	// Count reference types
	callCount := 0
	assignmentCount := 0
	typeRefCount := 0

	for _, ref := range references {
		switch ref.ReferenceType {
		case "call":
			callCount++
		case "assignment":
			assignmentCount++
		case "type_reference":
			typeRefCount++
		}
	}

	// Determine patterns
	if callCount > 0 {
		patterns = append(patterns, fmt.Sprintf("Called %d times", callCount))
	}
	if assignmentCount > 0 {
		patterns = append(patterns, fmt.Sprintf("Assigned %d times", assignmentCount))
	}
	if typeRefCount > 0 {
		patterns = append(patterns, fmt.Sprintf("Used as type %d times", typeRefCount))
	}

	// Pattern: heavily used
	if len(references) > 50 {
		patterns = append(patterns, "Heavily used across codebase")
	}

	// Pattern: concentrated usage
	usageByFile := make(map[int64]int)
	for _, ref := range references {
		usageByFile[ref.FileID]++
	}

	maxUsageInFile := 0
	for _, count := range usageByFile {
		if count > maxUsageInFile {
			maxUsageInFile = count
		}
	}

	if maxUsageInFile > 10 {
		patterns = append(patterns, fmt.Sprintf("Heavily used in one file (%d times)", maxUsageInFile))
	}

	// Pattern: widely spread
	if len(usageByFile) > 20 {
		patterns = append(patterns, fmt.Sprintf("Widely used across %d files", len(usageByFile)))
	}

	return patterns
}

// isDeprecated checks if a symbol is deprecated
func (ua *UsageAnalyzer) isDeprecated(symbol *types.Symbol) bool {
	// Check documentation for deprecation markers
	doc := strings.ToLower(symbol.Documentation)
	return strings.Contains(doc, "deprecated") ||
		strings.Contains(doc, "@deprecated") ||
		strings.Contains(doc, "obsolete")
}

// findAlternatives finds alternative symbols
func (ua *UsageAnalyzer) findAlternatives(symbol *types.Symbol) []string {
	alternatives := []string{}

	// If deprecated, look for alternatives in documentation
	if ua.isDeprecated(symbol) {
		doc := symbol.Documentation
		// Simple pattern matching for "use X instead"
		if strings.Contains(strings.ToLower(doc), "use") && strings.Contains(strings.ToLower(doc), "instead") {
			// Extract alternative name (simplified)
			lines := strings.Split(doc, "\n")
			for _, line := range lines {
				lower := strings.ToLower(line)
				if strings.Contains(lower, "use") && strings.Contains(lower, "instead") {
					alternatives = append(alternatives, strings.TrimSpace(line))
				}
			}
		}
	}

	return alternatives
}

// FindUnusedSymbols finds symbols that are never used
func (ua *UsageAnalyzer) FindUnusedSymbols(projectID int64) ([]*types.Symbol, error) {
	// Get all files for project
	files, err := ua.db.GetAllFilesForProject(projectID)
	if err != nil {
		return nil, err
	}

	unusedSymbols := []*types.Symbol{}

	for _, file := range files {
		// Get symbols in file
		symbols, err := ua.db.GetSymbolsByFile(file.ID)
		if err != nil {
			continue
		}

		for _, symbol := range symbols {
			// Skip exported symbols (might be used externally)
			if symbol.IsExported {
				continue
			}

			// Check references
			refs, err := ua.db.GetReferencesBySymbol(symbol.ID)
			if err != nil {
				continue
			}

			// No references = unused
			if len(refs) == 0 {
				unusedSymbols = append(unusedSymbols, symbol)
			}
		}
	}

	return unusedSymbols, nil
}

// FindMostUsedSymbols finds the most frequently used symbols
func (ua *UsageAnalyzer) FindMostUsedSymbols(projectID int64, limit int) ([]*types.SymbolUsageStats, error) {
	// Get all files for project
	files, err := ua.db.GetAllFilesForProject(projectID)
	if err != nil {
		return nil, err
	}

	usageStats := []*types.SymbolUsageStats{}

	for _, file := range files {
		symbols, err := ua.db.GetSymbolsByFile(file.ID)
		if err != nil {
			continue
		}

		for _, symbol := range symbols {
			stats, err := ua.AnalyzeUsage(symbol.Name)
			if err != nil {
				continue
			}
			usageStats = append(usageStats, stats)
		}
	}

	// Sort by usage count (simple bubble sort for small datasets)
	for i := 0; i < len(usageStats)-1; i++ {
		for j := 0; j < len(usageStats)-i-1; j++ {
			if usageStats[j].UsageCount < usageStats[j+1].UsageCount {
				usageStats[j], usageStats[j+1] = usageStats[j+1], usageStats[j]
			}
		}
	}

	// Limit results
	if len(usageStats) > limit {
		usageStats = usageStats[:limit]
	}

	return usageStats, nil
}

// AnalyzeAPIUsage analyzes usage of public API
func (ua *UsageAnalyzer) AnalyzeAPIUsage(projectID int64) (map[string]*types.SymbolUsageStats, error) {
	files, err := ua.db.GetAllFilesForProject(projectID)
	if err != nil {
		return nil, err
	}

	apiUsage := make(map[string]*types.SymbolUsageStats)

	for _, file := range files {
		symbols, err := ua.db.GetSymbolsByFile(file.ID)
		if err != nil {
			continue
		}

		for _, symbol := range symbols {
			// Only analyze exported symbols (public API)
			if !symbol.IsExported {
				continue
			}

			stats, err := ua.AnalyzeUsage(symbol.Name)
			if err != nil {
				continue
			}

			apiUsage[symbol.Name] = stats
		}
	}

	return apiUsage, nil
}
