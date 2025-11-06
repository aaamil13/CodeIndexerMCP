package ai

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// ContextExtractor extracts code context for AI analysis
type ContextExtractor struct {
	db *database.DB
}

// NewContextExtractor creates a new context extractor
func NewContextExtractor(db *database.DB) *ContextExtractor {
	return &ContextExtractor{db: db}
}

// ExtractContext extracts comprehensive context for a symbol
func (ce *ContextExtractor) ExtractContext(symbolName string, depth int) (*types.CodeContext, error) {
	// Get the symbol
	symbol, err := ce.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	// Get file information
	file, err := ce.db.GetFile(symbol.FileID)
	if err != nil {
		return nil, err
	}

	// Extract the actual code
	code, err := ce.extractCode(file.Path, symbol.StartLine, symbol.EndLine)
	if err != nil {
		return nil, err
	}

	// Get surrounding context (lines before and after)
	context, err := ce.extractContext(file.Path, symbol.StartLine, symbol.EndLine, 5)
	if err != nil {
		context = "" // Non-fatal
	}

	// Get dependencies (imports for this file)
	imports, err := ce.db.GetImportsByFile(file.ID)
	if err != nil {
		imports = []*types.Import{} // Non-fatal
	}

	dependencies := []string{}
	for _, imp := range imports {
		dependencies = append(dependencies, imp.Source)
	}

	// Get related symbols in the same file
	relatedSymbols, err := ce.db.GetSymbolsByFile(file.ID)
	if err != nil {
		relatedSymbols = []*types.Symbol{} // Non-fatal
	}

	// Get relationships
	relationships, err := ce.db.GetRelationshipsForSymbol(symbol.ID)
	if err != nil {
		relationships = []*types.Relationship{} // Non-fatal
	}

	// Build callers and callees lists
	callers := []*types.Symbol{}
	callees := []*types.Symbol{}

	for _, rel := range relationships {
		if rel.Type == types.RelationshipCalls {
			if rel.FromSymbolID == symbol.ID {
				// This symbol calls another
				if callee, err := ce.getSymbolByID(rel.ToSymbolID); err == nil {
					callees = append(callees, callee)
				}
			} else if rel.ToSymbolID == symbol.ID {
				// Another symbol calls this
				if caller, err := ce.getSymbolByID(rel.FromSymbolID); err == nil {
					callers = append(callers, caller)
				}
			}
		}
	}

	// Get usage examples
	usageExamples, err := ce.extractUsageExamples(symbol, depth)
	if err != nil {
		usageExamples = []*types.UsageExample{} // Non-fatal
	}

	return &types.CodeContext{
		Symbol:         symbol,
		File:           file,
		Code:           code,
		Dependencies:   dependencies,
		RelatedSymbols: relatedSymbols,
		Callers:        callers,
		Callees:        callees,
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
func (ce *ContextExtractor) extractUsageExamples(symbol *types.Symbol, maxExamples int) ([]*types.UsageExample, error) {
	// Get references to this symbol
	references, err := ce.db.GetReferencesBySymbol(symbol.ID)
	if err != nil {
		return nil, err
	}

	if len(references) == 0 {
		return []*types.UsageExample{}, nil
	}

	// Limit examples
	if len(references) > maxExamples {
		references = references[:maxExamples]
	}

	examples := []*types.UsageExample{}

	for _, ref := range references {
		file, err := ce.db.GetFile(ref.FileID)
		if err != nil {
			continue // Skip on error
		}

		// Extract code snippet around the reference
		code, err := ce.extractCode(file.Path, ref.LineNumber-2, ref.LineNumber+2)
		if err != nil {
			continue // Skip on error
		}

		// Get more context
		context, _ := ce.extractContext(file.Path, ref.LineNumber, ref.LineNumber, 3)

		examples = append(examples, &types.UsageExample{
			FilePath:    file.RelativePath,
			LineNumber:  ref.LineNumber,
			Code:        code,
			Context:     context,
			Description: fmt.Sprintf("%s at %s:%d", ref.ReferenceType, file.RelativePath, ref.LineNumber),
		})
	}

	return examples, nil
}

// getSymbolByID is a helper to get symbol by ID
func (ce *ContextExtractor) getSymbolByID(symbolID int64) (*types.Symbol, error) {
	// This is a simplified version - in production we'd have a proper DB query
	// For now, we'll skip this implementation
	return nil, fmt.Errorf("not implemented")
}
