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
func (se *SnippetExtractor) ExtractSmartSnippet(symbol *model.Symbol, includeTests bool) (*SmartSnippet, error) {
	// TODO: Implement after DB methods are available
	// // Get the symbol
	// if symbol == nil {
	// 	return nil, fmt.Errorf("symbol cannot be nil")
	// }

	// // Get file
	// file := symbol.File

	// // Extract main code
	// code, err := se.extractCode(file, symbol.Range.Start.Line, symbol.Range.End.Line)
	// if err != nil {
	// 	return nil, err
	// }

	// // Get dependencies (imports)
	// imports, err := se.db.GetImportsByFile(file)
	// if err != nil {
	// 	return nil, err
	// }

	// dependencies := []string{}
	// for _, imp := range imports {
	// 	dependencies = append(dependencies, imp.Path)
	// }

	// // Get related code (helper functions, types used by this symbol)
	// relatedCode, err := se.extractRelatedCode(symbol, file)
	// if err != nil {
	// 	relatedCode = []string{} // Non-fatal
	// }

	// // Generate usage hints
	// usageHints := se.generateUsageHints(symbol)

	// // Check if complete (has all dependencies resolved)
	// complete := se.isComplete(symbol, dependencies)

	// return &SmartSnippet{
	// 	Symbol:        symbol,
	// 	Code:          code,
	// 	Dependencies:  dependencies,
	// 	RelatedCode:   relatedCode,
	// 	Documentation: symbol.Documentation,
	// 	UsageHints:    usageHints,
	// 	Complete:      complete,
	// }, nil
	return nil, fmt.Errorf("not implemented")
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
func (se *SnippetExtractor) extractRelatedCode(symbol *model.Symbol, file string) ([]string, error) {
	// TODO: Implement after DB methods are available
	// relatedCode := []string{}

	// // Get all symbols in the file
	// symbols, err := se.db.GetSymbolsByFile(file)
	// if err != nil {
	// 	return relatedCode, err
	// }

	// // Find symbols that this symbol might depend on
	// // (This is simplified - in production we'd parse the code to find actual dependencies)
	// for _, sym := range symbols {
	// 	// Skip the symbol itself
	// 	if sym.ID == symbol.ID {
	// 		continue
	// 	}

	// 	// Include private helpers that might be used
	// 	isExported := strings.ToUpper(sym.Name[0:1]) == sym.Name[0:1]
	// 	if !isExported && (sym.Kind == "function" || sym.Kind == "method") { // Check for private function/method
	// 		code, err := se.extractCode(file, sym.Range.Start.Line, sym.Range.End.Line)
	// 		if err == nil {
	// 			relatedCode = append(relatedCode, code)
	// 		}
	// 	}

	// 	// Include types/interfaces this might use
	// 	if sym.Kind == "type" || sym.Kind == "interface" || sym.Kind == "struct" { // Check for type/interface/struct
	// 		code, err := se.extractCode(file, sym.Range.Start.Line, sym.Range.End.Line)
	// 		if err == nil {
	// 			relatedCode = append(relatedCode, code)
	// 		}
	// 	}
	// }

	// return relatedCode, nil
	return nil, fmt.Errorf("not implemented")
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
	// TODO: Implement after DB methods are available
	// if symbol == nil {
	// 	return "", fmt.Errorf("symbol cannot be nil")
	// }

	// file := symbol.File

	// return se.extractCode(file, symbol.Range.Start.Line, symbol.Range.End.Line)
	return "", fmt.Errorf("not implemented")
}

// ExtractWithContext extracts code with surrounding context
func (se *SnippetExtractor) ExtractWithContext(symbol *model.Symbol, contextLines int) (string, error) {
	// TODO: Implement after DB methods are available
	// if symbol == nil {
	// 	return "", fmt.Errorf("symbol cannot be nil")
	// }

	// file := symbol.File

	// startLine := symbol.Range.Start.Line - contextLines
	// if startLine < 1 {
	// 	startLine = 1
	// }
	// endLine := symbol.Range.End.Line + contextLines

	// return se.extractCode(file, startLine, endLine)
	return "", fmt.Errorf("not implemented")
}

func init() {
	// In production: Load and compile Tree-sitter queries
	// fmt.Println("TypeScript parser initialized (using fallback mode until Tree-sitter is fully integrated)")
}