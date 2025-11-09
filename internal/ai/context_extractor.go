package ai

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

// ContextExtractor extracts code context for AI analysis
type ContextExtractor struct {
	db *database.Manager
}

// NewContextExtractor creates a new context extractor
func NewContextExtractor(db *database.Manager) *ContextExtractor {
	return &ContextExtractor{db: db}
}

// ExtractContext extracts comprehensive context for a symbol
func (ce *ContextExtractor) ExtractContext(symbolName string, depth int) (*model.CodeContext, error) {
	// Get the symbol
	symbol, err := ce.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol by name %s: %w", symbolName, err)
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	// Get file information
	file, err := ce.db.GetFileByID(symbol.FileID)
	if err != nil || file == nil {
		return nil, fmt.Errorf("file not found for symbol %s: %w", symbolName, err)
	}

	// Extract the actual code
	code, err := ce.extractCode(file.Path, symbol.Range.Start.Line, symbol.Range.End.Line)
	if err != nil {
		return nil, fmt.Errorf("failed to extract code for symbol %s: %w", symbolName, err)
	}

	// Get surrounding context (lines before and after)
	context, err := ce.extractContext(file.Path, symbol.Range.Start.Line, symbol.Range.End.Line, 5)
	if err != nil {
		context = "" // Non-fatal, just means less context
	}

	// Get dependencies (imports for this file)
	imports, err := ce.db.GetImportsByFile(file.ID)
	if err != nil {
		imports = []*model.Import{} // Non-fatal
	}

	dependencies := []string{}
	for _, imp := range imports {
		dependencies = append(dependencies, imp.Path)
	}

	// Get related symbols in the same file
	relatedSymbols, err := ce.db.GetSymbolsByFile(file.ID)
	if err != nil {
		relatedSymbols = []*model.Symbol{} // Non-fatal
	}

	// Get relationships - currently not implemented in database.Manager
	// relationships, err := ce.db.GetRelationshipsForSymbol(symbol.ID)
	// if err != nil {
	// 	relationships = []*model.Relationship{} // Non-fatal
	// }

	// // Build callers and callees lists (placeholder)
	// callers := []*model.Symbol{}
	// callees := []*model.Symbol{}

	// for _, rel := range relationships {
	// 	if rel.Type == "calls" { // TODO: use model.RelationshipCalls
	// 		if rel.FromSymbolID == symbol.ID {
	// 			// This symbol calls another
	// 			if callee, err := ce.getSymbolByID(rel.ToSymbolID); err == nil {
	// 				callees = append(callees, callee)
	// 			}
	// 		} else if rel.ToSymbolID == symbol.ID {
	// 			// Another symbol calls this
	// 			if caller, err := ce.getSymbolByID(rel.FromSymbolID); err == nil {
	// 				callers = append(callers, caller)
	// 			}
	// 		}
	// 	}
	// }

	// Get usage examples
	usageExamples, err := ce.extractUsageExamples(symbol, depth)
	if err != nil {
		usageExamples = []*model.UsageExample{} // Non-fatal
	}

	return &model.CodeContext{
		Symbol:         symbol,
		File:           file.Path, // Use file.Path
		Code:           code,
		Dependencies:   dependencies,
		RelatedSymbols: relatedSymbols,
		Callers:        []*model.Symbol{}, // Placeholder
		Callees:        []*model.Symbol{}, // Placeholder
		UsageExamples:  usageExamples,
		Documentation:  symbol.Documentation,
		Context:        context,
	}, nil
}

// extractCode extracts code from a file
func (ce *ContextExtractor) extractCode(filePath string, startLine, endLine int) (string, error) {
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

// extractContext extracts surrounding context
func (ce *ContextExtractor) extractContext(filePath string, startLine, endLine, contextLines int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	contextStart := startLine - contextLines
	if contextStart < 1 {
		contextStart = 1
	}
	contextEnd := endLine + contextLines

	var lines []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		if lineNum >= contextStart && lineNum <= contextEnd {
			prefix := "  "
			if lineNum >= startLine && lineNum <= endLine {
				prefix = "▶ " // Mark the actual symbol lines
			}
			lines = append(lines, fmt.Sprintf("%s%4d | %s", prefix, lineNum, scanner.Text()))
		}
		if lineNum > contextEnd {
			break
		}
	}
	return strings.Join(lines, "\n"), scanner.Err()
}

// extractUsageExamples extracts usage examples for a symbol
func (ce *ContextExtractor) extractUsageExamples(symbol *model.Symbol, maxExamples int) ([]*model.UsageExample, error) {
	// Get references to this symbol
	references, err := ce.db.GetReferencesBySymbol(symbol.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get references for symbol %d: %w", symbol.ID, err)
	}

	if len(references) == 0 {
		return []*model.UsageExample{}, nil
	}

	// Limit examples
	if len(references) > maxExamples {
		references = references[:maxExamples]
	}

	examples := []*model.UsageExample{}

	for _, ref := range references {
		file := ref.FilePath

		// Extract code snippet around the reference (5 lines before and after)
		code, err := ce.extractCode(file, ref.Line-5, ref.Line+5)
		if err != nil {
			continue // Skip on error
		}

		// Get more context (3 lines before and after)
		context, _ := ce.extractContext(file, ref.Line, ref.Line, 3)

		examples = append(examples, &model.UsageExample{
			FilePath:    file,
			LineNumber:  ref.Line,
			Code:        code,
			Context:     context,
			Description: fmt.Sprintf("%s at %s:%d", ref.ReferenceType, file, ref.Line),
		})
	}

	return examples, nil
}

// getSymbolByID is a helper to get symbol by ID
func (ce *ContextExtractor) getSymbolByID(symbolID string) (*model.Symbol, error) {
	symID, err := strconv.Atoi(symbolID)
	if err != nil {
		return nil, fmt.Errorf("invalid symbol ID: %w", err)
	}
	return ce.db.GetSymbolByID(symID)
}
