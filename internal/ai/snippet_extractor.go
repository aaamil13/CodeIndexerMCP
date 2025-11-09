package ai

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

// SnippetExtractor extracts smart code snippets
type SnippetExtractor struct {
	db *database.Manager
}

// NewSnippetExtractor creates a new snippet extractor
func NewSnippetExtractor(db *database.Manager) *SnippetExtractor {
	return &SnippetExtractor{db: db}
}

// ExtractSmartSnippet extracts a self-contained code snippet
func (se *SnippetExtractor) ExtractSmartSnippet(symbol *model.Symbol, includeTests bool) (*model.SmartSnippet, error) {
	if symbol == nil {
		return nil, fmt.Errorf("symbol cannot be nil")
	}

	// Get file path and ID
	filePath := symbol.FilePath
	fileID := symbol.FileID

	// Extract main code
	code, err := se.extractCode(filePath, symbol.Range.Start.Line, symbol.Range.End.Line)
	if err != nil {
		return nil, fmt.Errorf("failed to extract code for symbol %s: %w", symbol.Name, err)
	}

	// Get dependencies (imports)
	imports, err := se.db.GetImportsByFile(fileID) // Pass FileID
	if err != nil {
		// Log error but continue, imports might not be critical for a snippet
		fmt.Printf("Warning: Failed to get imports for file ID %d: %v\n", fileID, err)
		imports = []*model.Import{}
	}

	dependencies := []string{}
	for _, imp := range imports {
		dependencies = append(dependencies, imp.Path)
	}

	// Get related code (helper functions, types used by this symbol)
	relatedCode, err := se.extractRelatedCode(symbol, filePath) // Corrected call
	if err != nil {
		fmt.Printf("Warning: Failed to extract related code for symbol %s: %v\n", symbol.Name, err)
		relatedCode = []string{} // Non-fatal
	}

	// Generate usage hints
	usageHints := se.generateUsageHints(symbol)

	// Check if complete (has all dependencies resolved)
	complete := se.isComplete(symbol, dependencies)

	return &model.SmartSnippet{
		Symbol:        symbol,
		Code:          code,
		Dependencies:  dependencies,
		RelatedCode:   relatedCode,
		Documentation: symbol.Documentation,
		UsageHints:    usageHints,
		Complete:      complete,
	}, nil
}

// extractCode extracts code from a file
func (se *SnippetExtractor) extractCode(filePath string, startLine, endLine int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		if lineNum >= startLine && lineNum <= endLine {
			lines = append(lines, scanner.Text())
		}
		if lineNum > endLine {
			break
		}
	}

	return strings.Join(lines, "\n"), scanner.Err()
}

// extractRelatedCode extracts related code (helper functions, types)
func (se *SnippetExtractor) extractRelatedCode(symbol *model.Symbol, filePath string) ([]string, error) {
	relatedCode := []string{}

	// Get all symbols in the file
	symbols, err := se.db.GetSymbolsByFile(symbol.FileID) // Use symbol.FileID
	if err != nil {
		return relatedCode, fmt.Errorf("failed to get symbols for file ID %d: %w", symbol.FileID, err)
	}

	// Find symbols that this symbol might depend on or relate to
	for _, sym := range symbols {
		// Skip the symbol itself
		if sym.ID == symbol.ID {
			continue
		}

		// Heuristic: Include symbols that are defined close by or are of a related kind
		// This is a simplification; a real implementation would analyze usage within the snippet.
		if (sym.Kind == model.SymbolKindFunction || sym.Kind == model.SymbolKindMethod ||
			sym.Kind == model.SymbolKindClass || model.SymbolKindInterface ||
			sym.Kind == model.SymbolKindStruct || sym.Kind == model.SymbolKindType) &&
			(sym.Range.Start.Line >= symbol.Range.Start.Line-10 && sym.Range.End.Line <= symbol.Range.End.Line+10) { // Within 10 lines
			code, err := se.extractCode(filePath, sym.Range.Start.Line, sym.Range.End.Line) // Use filePath consistently
			if err == nil && code != "" {
				relatedCode = append(relatedCode, code)
			}
		}
	}

	return relatedCode, nil
}

// generateUsageHints generates hints on how to use the symbol
func (se *SnippetExtractor) generateUsageHints(symbol *model.Symbol) []string {
	hints := []string{}

	switch symbol.Kind {
	case "function":
		if symbol.Signature != "" {
			hints = append(hints, fmt.Sprintf("Call with: %s", symbol.Signature))
		}
		// TODO: Check for IsAsync
		// if symbol.IsAsync {
		// 	hints = append(hints, "This is an async function - use await or handle promise")
		// }
		isExported := strings.ToUpper(symbol.Name[0:1]) == symbol.Name[0:1]
		if isExported {
			hints = append(hints, "This is exported and can be imported from other modules")
		}

	case "class":
		hints = append(hints, fmt.Sprintf("Instantiate with: new %s()", symbol.Name))
		isExported := strings.ToUpper(symbol.Name[0:1]) == symbol.Name[0:1]
		if isExported {
			hints = append(hints, "This class is exported and can be imported")
		}

	case "interface":
		hints = append(hints, "This is an interface - implement it in your class/struct")

	case "type":
		hints = append(hints, "This is a type definition - use it for type annotations")
	}

	// Language-specific hints
	switch symbol.Language {
	case "go":
		isExported := strings.ToUpper(symbol.Name[0:1]) == symbol.Name[0:1]
		if isExported {
			hints = append(hints, "Import from package to use")
		}
	case "python":
		if symbol.Name == "__init__" {
			hints = append(hints, "This is a constructor - called automatically when creating instance")
		}
	case "typescript":
		hints = append(hints, "TypeScript provides type safety - check types when using")
	}

	return hints
}

// isComplete checks if snippet is complete and runnable
func (se *SnippetExtractor) isComplete(symbol *model.Symbol, dependencies []string) bool {
	// Simple heuristic - if it has few dependencies and is self-contained
	if len(dependencies) > 10 {
		return false
	}

	// Functions without external deps are usually complete
	if symbol.Kind == "function" && len(dependencies) < 5 {
		return true
	}

	return false
}

// ExtractMinimalSnippet extracts the absolute minimum code needed
func (se *SnippetExtractor) ExtractMinimalSnippet(symbol *model.Symbol) (string, error) {
	if symbol == nil {
		return "", fmt.Errorf("symbol cannot be nil")
	}

	filePath := symbol.FilePath

	return se.extractCode(filePath, symbol.Range.Start.Line, symbol.Range.End.Line)
}

// ExtractWithContext extracts code with surrounding context
func (se *SnippetExtractor) ExtractWithContext(symbol *model.Symbol, contextLines int) (string, error) {
	if symbol == nil {
		return "", fmt.Errorf("symbol cannot be nil")
	}

	filePath := symbol.FilePath

	startLine := symbol.Range.Start.Line - contextLines
	if startLine < 1 {
		startLine = 1
	}
	endLine := symbol.Range.End.Line + contextLines

	return se.extractCode(filePath, startLine, endLine)
}

func init() {
	// In production: Load and compile Tree-sitter queries
	// fmt.Println("TypeScript parser initialized (using fallback mode until Tree-sitter is fully integrated)")
}
