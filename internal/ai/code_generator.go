package ai

import (
	"context"
	"fmt"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/database"
)

// CodeGenerator provides functionality to generate code based on a given prompt and context.
type CodeGenerator struct {
	dbManager *database.Manager
	// Potentially add a client for an external AI service here
	// For now, it will be a mock or a simple rule-based generator
}

// NewCodeGenerator creates a new instance of CodeGenerator.
func NewCodeGenerator(dbManager *database.Manager) *CodeGenerator {
	return &CodeGenerator{
		dbManager: dbManager,
	}
}

// GenerateCode generates code based on the provided prompt and context symbols.
// It can optionally take a targetSymbolID to focus the generation on a specific code element.
func (cg *CodeGenerator) GenerateCode(ctx context.Context, prompt string, targetSymbolID string) (string, error) {
	// In a real-world scenario, this would interact with a sophisticated AI model
	// and use the indexed code as context.

	// For demonstration, let's create a simple mock generation.
	// If a targetSymbolID is provided, fetch its details and include in the context.
	var contextString string
	if targetSymbolID != "" {
		symbol, err := cg.dbManager.GetSymbol(targetSymbolID)
		if err != nil {
			return "", fmt.Errorf("failed to get target symbol: %w", err)
		}
		if symbol != nil {
			contextString = fmt.Sprintf("Context from symbol '%s' (%s):\nKind: %s\nFile: %s\nSignature: %s\nDocumentation: %s\n",
				symbol.Name, symbol.ID, symbol.Kind, symbol.File, symbol.Signature, symbol.Documentation)
			
			// If it's a function, get its details
			if symbol.Kind == model.SymbolKindFunction || symbol.Kind == model.SymbolKindMethod {
				function, err := cg.dbManager.GetFunctionDetails(targetSymbolID)
				if err != nil {
					return "", fmt.Errorf("failed to get function details for '%s': %w", targetSymbolID, err)
				}
				if function != nil {
					contextString += fmt.Sprintf("Body: %s\nParameters: %+v\n", function.Body, function.Parameters)
				}
			} else if symbol.Kind == model.SymbolKindClass {
				class, err := cg.dbManager.GetClassDetails(targetSymbolID)
				if err != nil {
					return "", fmt.Errorf("failed to get class details for '%s': %w", targetSymbolID, err)
				}
				if class != nil {
					contextString += fmt.Sprintf("Fields: %+v\n", class.Fields)
				}
			}
		}
	}

	generatedCode := fmt.Sprintf("// Generated code based on prompt: \"%s\"\n", prompt)
	if contextString != "" {
		generatedCode += fmt.Sprintf("// Using context:\n// %s\n", contextString)
	}
	generatedCode += `
func ExampleFunction() {
	// This is a placeholder for AI-generated code.
	// In a real scenario, an AI model would produce meaningful code here.
}
`
	return generatedCode, nil
}
