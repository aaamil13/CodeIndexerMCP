package ai

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// SnippetExtractor extracts smart code snippets
type SnippetExtractor struct {
	db *database.DB
}

// NewSnippetExtractor creates a new snippet extractor
func NewSnippetExtractor(db *database.DB) *SnippetExtractor {
	return &SnippetExtractor{db: db}
}

// ExtractSmartSnippet extracts a self-contained code snippet
func (se *SnippetExtractor) ExtractSmartSnippet(symbolName string, includeTests bool) (*types.SmartSnippet, error) {
	// Get the symbol
	symbol, err := se.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	// Get file
	file, err := se.db.GetFile(symbol.FileID)
	if err != nil {
		return nil, err
	}

	// Extract main code
	code, err := se.extractCode(file.Path, symbol.StartLine, symbol.EndLine)
	if err != nil {
		return nil, err
	}

	// Get dependencies (imports)
	imports, err := se.db.GetImportsByFile(file.ID)
	if err != nil {
		return nil, err
	}

	dependencies := []string{}
	for _, imp := range imports {
		dependencies = append(dependencies, imp.Source)
	}

	// Get related code (helper functions, types used by this symbol)
	relatedCode, err := se.extractRelatedCode(symbol, file)
	if err != nil {
		relatedCode = []string{} // Non-fatal
	}

	// Generate usage hints
	usageHints := se.generateUsageHints(symbol, file.Language)

	// Check if complete (has all dependencies resolved)
	complete := se.isComplete(symbol, dependencies)

	return &types.SmartSnippet{
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
func (se *SnippetExtractor) extractRelatedCode(symbol *types.Symbol, file *types.File) ([]string, error) {
	relatedCode := []string{}

	// Get all symbols in the file
	symbols, err := se.db.GetSymbolsByFile(file.ID)
	if err != nil {
		return relatedCode, err
	}

	// Find symbols that this symbol might depend on
	// (This is simplified - in production we'd parse the code to find actual dependencies)
	for _, sym := range symbols {
		// Skip the symbol itself
		if sym.ID == symbol.ID {
			continue
		}

		// Include private helpers that might be used
		if sym.Visibility == types.VisibilityPrivate && sym.Type == types.SymbolTypeFunction {
			code, err := se.extractCode(file.Path, sym.StartLine, sym.EndLine)
			if err == nil {
				relatedCode = append(relatedCode, code)
			}
		}

		// Include types/interfaces this might use
		if sym.Type == types.SymbolTypeType || sym.Type == types.SymbolTypeInterface || sym.Type == types.SymbolTypeStruct {
			code, err := se.extractCode(file.Path, sym.StartLine, sym.EndLine)
			if err == nil {
				relatedCode = append(relatedCode, code)
			}
		}
	}

	return relatedCode, nil
}

// generateUsageHints generates hints on how to use the symbol
func (se *SnippetExtractor) generateUsageHints(symbol *types.Symbol, language string) []string {
	hints := []string{}

	switch symbol.Type {
	case types.SymbolTypeFunction:
		if symbol.Signature != "" {
			hints = append(hints, fmt.Sprintf("Call with: %s", symbol.Signature))
		}
		if symbol.IsAsync {
			hints = append(hints, "This is an async function - use await or handle promise")
		}
		if symbol.IsExported {
			hints = append(hints, "This is exported and can be imported from other modules")
		}

	case types.SymbolTypeClass:
		hints = append(hints, fmt.Sprintf("Instantiate with: new %s()", symbol.Name))
		if symbol.IsExported {
			hints = append(hints, "This class is exported and can be imported")
		}

	case types.SymbolTypeInterface:
		hints = append(hints, "This is an interface - implement it in your class/struct")

	case types.SymbolTypeType:
		hints = append(hints, "This is a type definition - use it for type annotations")
	}

	// Language-specific hints
	switch language {
	case "go":
		if symbol.IsExported {
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
func (se *SnippetExtractor) isComplete(symbol *types.Symbol, dependencies []string) bool {
	// Simple heuristic - if it has few dependencies and is self-contained
	if len(dependencies) > 10 {
		return false
	}

	// Functions without external deps are usually complete
	if symbol.Type == types.SymbolTypeFunction && len(dependencies) < 5 {
		return true
	}

	return false
}

// ExtractMinimalSnippet extracts the absolute minimum code needed
func (se *SnippetExtractor) ExtractMinimalSnippet(symbolName string) (string, error) {
	symbol, err := se.db.GetSymbolByName(symbolName)
	if err != nil {
		return "", err
	}
	if symbol == nil {
		return "", fmt.Errorf("symbol not found: %s", symbolName)
	}

	file, err := se.db.GetFile(symbol.FileID)
	if err != nil {
		return "", err
	}

	return se.extractCode(file.Path, symbol.StartLine, symbol.EndLine)
}

// ExtractWithContext extracts code with surrounding context
func (se *SnippetExtractor) ExtractWithContext(symbolName string, contextLines int) (string, error) {
	symbol, err := se.db.GetSymbolByName(symbolName)
	if err != nil {
		return "", err
	}
	if symbol == nil {
		return "", fmt.Errorf("symbol not found: %s", symbolName)
	}

	file, err := se.db.GetFile(symbol.FileID)
	if err != nil {
		return "", err
	}

	startLine := symbol.StartLine - contextLines
	if startLine < 1 {
		startLine = 1
	}
	endLine := symbol.EndLine + contextLines

	return se.extractCode(file.Path, startLine, endLine)
}

func init() {
	// In production: Load and compile Tree-sitter queries
	fmt.Println("TypeScript parser initialized (using fallback mode until Tree-sitter is fully integrated)")
}
