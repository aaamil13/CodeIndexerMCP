package ai

import (
	"fmt"
	"strings"
	"strconv" // Added
	"sort"    // Added

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

// UsageAnalyzer analyzes symbol usage patterns
type UsageAnalyzer struct {
	db *database.Manager
}

// NewUsageAnalyzer creates a new usage analyzer
func NewUsageAnalyzer(db *database.Manager) *UsageAnalyzer {
	return &UsageAnalyzer{db: db}
}

// AnalyzeUsage analyzes usage statistics for a symbol
func (ua *UsageAnalyzer) AnalyzeUsage(symbol *model.Symbol) (*model.SymbolUsageStats, error) {
	if symbol == nil {
		return nil, fmt.Errorf("symbol cannot be nil")
	}

	// Get all references where this symbol is the target
	allReferences, err := ua.db.GetReferencesBySymbol(symbol.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get references for symbol %s: %w", symbol.ID, err)
	}

	// Filter references to only count where 'symbol' is the target (i.e., being used)
	referencesToThisSymbol := []*model.Reference{}
	for _, ref := range allReferences {
		// GetReferencesBySymbol returns references where current symbol is either source or target.
		// We only want to count where it's the target for usage analysis.
		if ref.TargetSymbolName == symbol.Name {
			referencesToThisSymbol = append(referencesToThisSymbol, ref)
		}
	}

	usageCount := len(referencesToThisSymbol)

	// Group by file
	usageByFile := make(map[string]int)
	fileSet := make(map[string]bool)

	for _, ref := range referencesToThisSymbol {
		fileSet[ref.FilePath] = true
		usageByFile[ref.FilePath]++
	}

	fileCount := len(fileSet)

	// Detect common usage patterns
	commonPatterns := ua.detectUsagePatterns(symbol, referencesToThisSymbol)

	// Check if deprecated
	isDeprecated := ua.isDeprecated(symbol)

	// Find alternatives
	alternatives := ua.findAlternatives(symbol)

	return &model.SymbolUsageStats{
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
func (ua *UsageAnalyzer) detectUsagePatterns(symbol *model.Symbol, references []*model.Reference) []string {
	patterns := []string{}

	// Count reference types
	callCount := 0
	assignmentCount := 0
	typeRefCount := 0

	for _, ref := range references {
		switch ref.ReferenceType {
		case "calls": // Use "calls" from model.Reference
			callCount++
		case "assignment": // Assuming "assignment" is a valid type
			assignmentCount++
		case "type_reference": // Assuming "type_reference" is a valid type
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
	usageByFile := make(map[string]int)
	for _, ref := range references {
		usageByFile[ref.FilePath]++
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
func (ua *UsageAnalyzer) isDeprecated(symbol *model.Symbol) bool {
	// Check documentation for deprecation markers
	doc := strings.ToLower(symbol.Documentation)
	return strings.Contains(doc, "deprecated") ||
		strings.Contains(doc, "@deprecated") ||
		strings.Contains(doc, "obsolete")
}

// findAlternatives finds alternative symbols
func (ua *UsageAnalyzer) findAlternatives(symbol *model.Symbol) []string {
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
func (ua *UsageAnalyzer) FindUnusedSymbols(projectID string) ([]*model.Symbol, error) {
	projID, err := strconv.Atoi(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	files, err := ua.db.GetAllFilesForProject(projID)
	if err != nil {
		return nil, fmt.Errorf("failed to get files for project %s: %w", projectID, err)
	}

	unusedSymbols := []*model.Symbol{}

	for _, file := range files {
		symbols, err := ua.db.GetSymbolsByFile(file.Path)
		if err != nil {
			fmt.Printf("Warning: Failed to get symbols for file %s: %v\n", file.Path, err)
			continue
		}

		for _, symbol := range symbols {
			// Skip exported symbols (might be used externally)
			if strings.ToUpper(symbol.Name[0:1]) == symbol.Name[0:1] { // Check for exported
				continue
			}

			// Check if symbol has any external references (not self-references)
			references, err := ua.db.GetReferencesBySymbol(symbol.ID)
			if err != nil {
				fmt.Printf("Warning: Failed to get references for symbol %s: %v\n", symbol.ID, err)
				continue
			}

			isUsedExternally := false
			for _, ref := range references {
				if ref.SourceSymbolID != symbol.ID { // If referenced by something other than itself
					isUsedExternally = true
					break
				}
			}

			if !isUsedExternally {
				unusedSymbols = append(unusedSymbols, symbol)
			}
		}
	}

	return unusedSymbols, nil
}

// FindMostUsedSymbols finds the most frequently used symbols
func (ua *UsageAnalyzer) FindMostUsedSymbols(projectID string, limit int) ([]*model.SymbolUsageStats, error) {
	projID, err := strconv.Atoi(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	files, err := ua.db.GetAllFilesForProject(projID)
	if err != nil {
		return nil, fmt.Errorf("failed to get files for project %s: %w", projectID, err)
	}

	var allUsageStats []*model.SymbolUsageStats

	for _, file := range files {
		symbols, err := ua.db.GetSymbolsByFile(file.Path)
		if err != nil {
			fmt.Printf("Warning: Failed to get symbols for file %s: %v\n", file.Path, err)
			continue
		}

		for _, symbol := range symbols {
			stats, err := ua.AnalyzeUsage(symbol)
			if err != nil {
				fmt.Printf("Warning: Failed to analyze usage for symbol %s: %v\n", symbol.ID, err)
				continue
			}
			allUsageStats = append(allUsageStats, stats)
		}
	}

	// Sort by usage count (descending)
	sort.Slice(allUsageStats, func(i, j int) bool {
		return allUsageStats[i].UsageCount > allUsageStats[j].UsageCount
	})

	// Limit results
	if len(allUsageStats) > limit {
		allUsageStats = allUsageStats[:limit]
	}

	return allUsageStats, nil
}

// AnalyzeAPIUsage analyzes usage of public API
func (ua *UsageAnalyzer) AnalyzeAPIUsage(projectID string) (map[string]*model.SymbolUsageStats, error) {
	projID, err := strconv.Atoi(projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	files, err := ua.db.GetAllFilesForProject(projID)
	if err != nil {
		return nil, fmt.Errorf("failed to get files for project %s: %w", projectID, err)
	}

	apiUsage := make(map[string]*model.SymbolUsageStats)

	for _, file := range files {
		symbols, err := ua.db.GetSymbolsByFile(file.Path)
		if err != nil {
			fmt.Printf("Warning: Failed to get symbols for file %s: %v\n", file.Path, err)
			continue
		}

		for _, symbol := range symbols {
			// Only analyze exported symbols (public API in many languages starts with an uppercase letter)
			if strings.ToUpper(symbol.Name[0:1]) == symbol.Name[0:1] {
				stats, err := ua.AnalyzeUsage(symbol)
				if err != nil {
					fmt.Printf("Warning: Failed to analyze usage for API symbol %s: %v\n", symbol.ID, err)
					continue
				}
				apiUsage[symbol.Name] = stats
			}
		}
	}

	return apiUsage, nil
}
