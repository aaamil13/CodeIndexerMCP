package ai

import (
	"fmt"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

// TypeValidator validates types and finds undefined usages
type TypeValidator struct {
	db *database.Manager
}

// NewTypeValidator creates a new type validator
func NewTypeValidator(db *database.Manager) *TypeValidator {
	return &TypeValidator{
		db: db,
	}
}

// ValidateFile validates all types in a file
func (tv *TypeValidator) ValidateFile(filePath string) (*model.TypeValidation, error) {
	validation := &model.TypeValidation{
		File:             filePath,
		IsValid:          true,
		UndefinedSymbols: make([]*model.UndefinedUsage, 0),
		TypeMismatches:   make([]*model.TypeMismatch, 0),
		MissingMethods:   make([]*model.MissingMethod, 0),
		InvalidCalls:     make([]*model.InvalidCall, 0),
		UnusedImports:    make([]*model.Import, 0),
		Suggestions:      make([]string, 0),
	}

	// Get all symbols in this file
	symbols, err := tv.db.GetSymbolsByFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols for file %s: %w", filePath, err)
	}

	// Get all references in this file
	references, err := tv.db.GetReferencesByFile(filePath) // Assuming GetReferencesByFile exists and returns references where SourceSymbolID is in this file.
	if err != nil {
		fmt.Printf("Warning: Failed to get references for file %s: %v\n", filePath, err)
		references = []*model.Reference{} // Non-fatal
	}

	// Build symbol map for quick lookup
	symbolMap := make(map[string]*model.Symbol) // Map symbol name to symbol for faster lookup
	symbolIDMap := make(map[string]*model.Symbol) // Map symbol ID to symbol
	for _, sym := range symbols {
		symbolMap[sym.Name] = sym
		symbolIDMap[sym.ID] = sym
	}

	// Check each reference
	for _, ref := range references {
		// Try to resolve the referenced symbol
		referencedSymbol, err := tv.db.GetSymbolByName(ref.TargetSymbolName) // Try by name first
		if err != nil || referencedSymbol == nil {
			// If not found by name, it might be an undefined usage
			undefined := &model.UndefinedUsage{
				SymbolName:  ref.TargetSymbolName,
				FilePath:    filePath,
				Line:        ref.Line,
				Description: fmt.Sprintf("Reference to undefined symbol '%s'", ref.TargetSymbolName),
				// UsageType:   tv.inferUsageType(ref.ReferenceType), // Inferring usage type from reference type string
			}
			// Find similar symbols (typo suggestions)
			similar := tv.findSimilarSymbols(ref.TargetSymbolName, symbols) // Search within current file symbols
			if len(similar) > 0 {
				undefined.Description += fmt.Sprintf(". Did you mean '%s'?", similar[0].Name)
			}
			validation.UndefinedSymbols = append(validation.UndefinedSymbols, undefined)
			validation.IsValid = false
			continue
		}

		// Validate the reference based on symbol types (simplified for now)
		// This is where more complex type checking logic would go.
		// For example, if ref.ReferenceType indicates a function call,
		// ensure referencedSymbol is a function and check argument count/types.
		// tv.validateReference(ref, referencedSymbol, filePath, validation)
	}

	// Check for unused imports
	imports, err := tv.db.GetImportsByFile(filePath)
	if err == nil {
		for _, imp := range imports {
			if !tv.isImportUsed(imp, references) {
				validation.UnusedImports = append(validation.UnusedImports, imp)
				validation.Suggestions = append(validation.Suggestions,
					fmt.Sprintf("Import '%s' is unused and can be removed", imp.Path))
			}
		}
	}

	// Generate summary suggestions
	if len(validation.UndefinedSymbols) > 0 {
		validation.Suggestions = append(validation.Suggestions,
			fmt.Sprintf("Found %d undefined symbols. Check for typos or missing imports.",
				len(validation.UndefinedSymbols)))
	}

	if len(validation.TypeMismatches) > 0 {
		validation.Suggestions = append(validation.Suggestions,
			fmt.Sprintf("Found %d type mismatches. Review type conversions.",
				len(validation.TypeMismatches)))
	}

	return validation, nil
}

// FindUndefinedUsages finds all undefined symbols in a file
func (tv *TypeValidator) FindUndefinedUsages(filePath string) ([]*model.UndefinedUsage, error) {
	validation, err := tv.ValidateFile(filePath)
	if err != nil {
		return nil, err
	}
	return validation.UndefinedSymbols, nil
}

// CheckMethodExists checks if a method exists on a type
func (tv *TypeValidator) CheckMethodExists(typeName, methodName string, projectID string) (*model.MissingMethod, error) {
	// Find the type symbol
	typeSymbol, err := tv.db.GetSymbolByName(typeName)
	if err != nil {
		return &model.MissingMethod{
			TypeName:   typeName,
			MethodName: methodName,
			Suggestion: fmt.Sprintf("Failed to retrieve type '%s': %v", typeName, err),
		}, nil
	}
	if typeSymbol == nil {
		return &model.MissingMethod{
			TypeName:   typeName,
			MethodName: methodName,
			Suggestion: fmt.Sprintf("Type '%s' not found", typeName),
		}, nil
	}

	// Find methods of this type
	methods, err := tv.db.GetMethodsForType(typeSymbol.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get methods for type %s: %w", typeName, err)
	}

	// Check if method exists
	for _, method := range methods {
		if method.Name == methodName {
			return nil, nil // Method exists
		}
	}

	// Method doesn't exist - build response
	missing := &model.MissingMethod{
		TypeName:         typeName,
		MethodName:       methodName,
		AvailableMethods: make([]string, 0),
	}

	// List available methods
	for _, method := range methods {
		missing.AvailableMethods = append(missing.AvailableMethods, method.Name)
	}

	// Find similar method names
	similar := tv.findSimilarMethodName(methodName, methods)
	if similar != nil {
		missing.Suggestion = fmt.Sprintf("Did you mean '%s'?", similar.Name)
	} else if len(methods) > 0 {
		missing.Suggestion = fmt.Sprintf("Available methods for '%s': %s", typeName, strings.Join(missing.AvailableMethods, ", "))
	} else {
		missing.Suggestion = fmt.Sprintf("No methods found for type '%s'", typeName)
	}

	return missing, nil
}

// ValidateSymbolTypes validates types for a specific symbol
func (tv *TypeValidator) ValidateSymbolTypes(symbolID string) (*model.TypeValidation, error) {
	symbol, err := tv.db.GetSymbol(symbolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol %s: %w", symbolID, err)
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolID)
	}

	file := symbol.File

	validation := &model.TypeValidation{
		File:             file,
		Symbol:           symbol,
		IsValid:          true,
		UndefinedSymbols: make([]*model.UndefinedUsage, 0),
		TypeMismatches:   make([]*model.TypeMismatch, 0),
		MissingMethods:   make([]*model.MissingMethod, 0),
		InvalidCalls:     make([]*model.InvalidCall, 0),
		Suggestions:      make([]string, 0),
	}

	// This part needs `GetRelationshipsForSymbol` implemented in `database.Manager`.
	// For now, we will simulate or use references as a proxy.
	// Assume references where SourceSymbolID == symbol.ID are "calls" or usages by this symbol.
	references, err := tv.db.GetReferencesBySymbol(symbol.ID)
	if err != nil {
		fmt.Printf("Warning: Failed to get references for symbol %s for validation: %v\n", symbol.ID, err)
		return validation, nil // Return current validation with a warning
	}

	for _, ref := range references {
		// Only consider "call" relationships for now
		if ref.ReferenceType == model.ReferenceTypeCalls {
			calledSymbol, err := tv.db.GetSymbolByName(ref.TargetSymbolName)
			if err != nil || calledSymbol == nil {
				// Called symbol doesn't exist - undefined usage
				undefined := &model.UndefinedUsage{
					SymbolName:  ref.TargetSymbolName,
					FilePath:    file,
					Line:        ref.Line,
					Description: fmt.Sprintf("Function/method '%s' called by '%s' is undefined", ref.TargetSymbolName, symbol.Name),
				}
				validation.UndefinedSymbols = append(validation.UndefinedSymbols, undefined)
				validation.IsValid = false
			} else {
				// Basic type validation for calls
				// This would involve checking if calledSymbol is actually a function/method
				// and if the number/types of arguments match. This is highly language-specific.
				if calledSymbol.Kind != model.SymbolKindFunction && calledSymbol.Kind != model.SymbolKindMethod {
					invalidCall := &model.InvalidCall{
						CallerSymbol: symbol,
						CalledSymbol: calledSymbol,
						FilePath:     file,
						Line:         ref.Line,
						Column:       ref.Column,
						Message:      fmt.Sprintf("Symbol '%s' (kind: %s) is not a callable function/method", calledSymbol.Name, calledSymbol.Kind),
					}
					validation.InvalidCalls = append(validation.InvalidCalls, invalidCall)
					validation.IsValid = false
				}
			}
		}
	}

	return validation, nil
}

// CheckTypeCompatibility checks if two types are compatible
func (tv *TypeValidator) CheckTypeCompatibility(type1, type2 string) bool {
	// Normalize types
	t1 := strings.TrimSpace(type1)
	t2 := strings.TrimSpace(type2)

	// Direct match
	if t1 == t2 {
		return true
	}

	// Check for pointer compatibility (*T and T)
	if strings.HasPrefix(t1, "*") && t1[1:] == t2 {
		return true
	}
	if strings.HasPrefix(t2, "*") && t2[1:] == t1 {
		return true
	}

	// Check for interface{} (any type)
	if t1 == "interface{}" || t2 == "interface{}" {
		return true
	}

	// Check for numeric type compatibility
	numericTypes := map[string]bool{
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true,
	}
	if numericTypes[t1] && numericTypes[t2] {
		return true // Allow numeric conversions (with warning)
	}

	return false
}

// CalculateTypeSafetyScore calculates type safety score for a file or project
func (tv *TypeValidator) CalculateTypeSafetyScore(filePath string) (*model.TypeSafetyScore, error) {
	validation, err := tv.ValidateFile(filePath)
	if err != nil {
		return nil, err
	}

	symbols, err := tv.db.GetSymbolsByFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols for file %s: %w", filePath, err)
	}

	score := &model.TypeSafetyScore{
		TotalSymbols:   len(symbols),
		TypedSymbols:   0,
		UntypedSymbols: 0,
		ErrorCount:     len(validation.UndefinedSymbols) + len(validation.TypeMismatches),
		WarningCount:   len(validation.InvalidCalls) + len(validation.MissingMethods),
	}

	// Count typed vs untyped symbols
	for _, sym := range symbols {
		// Simplified heuristic: A symbol is "typed" if it has a non-empty Signature or Type,
		// or if it's a function/method (which inherently have return types/parameters that can be typed).
		if sym.Signature != "" || sym.Type != "" || sym.Kind == model.SymbolKindFunction || sym.Kind == model.SymbolKindMethod {
			score.TypedSymbols++
		} else {
			score.UntypedSymbols++
		}
	}

	// Calculate score (0-100)
	if score.TotalSymbols == 0 {
		score.Score = 100.0
	} else {
		typedRatio := float64(score.TypedSymbols) / float64(score.TotalSymbols)
		errorPenalty := float64(score.ErrorCount) * 5.0
		warningPenalty := float64(score.WarningCount) * 2.0

		// Base score on typed ratio, then apply penalties
		score.Score = (typedRatio * 100.0) - errorPenalty - warningPenalty
		if score.Score < 0 {
			score.Score = 0
		}
		if score.Score > 100 {
			score.Score = 100
		}
	}

	// Determine rating
	if score.Score >= 90 {
		score.Rating = "excellent"
		score.Recommendation = "Code has excellent type safety. Keep it up!"
	} else if score.Score >= 75 {
		score.Rating = "good"
		score.Recommendation = "Code has good type safety. Consider adding more type annotations."
	} else if score.Score >= 50 {
		score.Rating = "fair"
		score.Recommendation = "Code has fair type safety. Add type annotations and fix type errors."
	} else {
		score.Rating = "poor"
		score.Recommendation = "Code has poor type safety. Urgent: add types and fix errors."
	}

	return score, nil
}

// Helper methods

func (tv *TypeValidator) extractSymbolNameFromContext(context string) string {
	// Simple extraction - can be improved
	parts := strings.Fields(context)
	for _, part := range parts {
		if strings.Contains(part, "(") {
			// Likely a function call
			return strings.Split(part, "(")[0]
		}
	}
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

func (tv *TypeValidator) inferUsageType(context string) string {
	if strings.Contains(context, "(") {
		return "function"
	}
	if strings.Contains(context, ".") {
		return "method"
	}
	return "variable"
}

func (tv *TypeValidator) findSimilarSymbols(name string, symbols []*model.Symbol) []*model.Symbol {
	similar := make([]*model.Symbol, 0)

	for _, sym := range symbols {
		if tv.levenshteinDistance(name, sym.Name) <= 2 {
			similar = append(similar, sym)
		}
	}

	return similar
}

func (tv *TypeValidator) findSimilarMethodName(name string, methods []*model.Method) *model.Method {
	for _, method := range methods {
		if tv.levenshteinDistance(name, method.Name) <= 2 {
			return method
		}
	}
	return nil
}

func (tv *TypeValidator) levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func (tv *TypeValidator) validateReference(ref *model.Reference, refSymbol *model.Symbol, file string, validation *model.TypeValidation) {
	// Check if this is a method call on an object
	refType := ref.ReferenceType
	if strings.Contains(refType, ".") {
		parts := strings.Split(refType, ".")
		if len(parts) >= 2 {
			// This is potentially a method call
			// We'd need to check if the method exists on the type
			// For now, just note it
		}
	}

	// Check if this is a function call with wrong number of arguments
	if strings.Contains(refType, "(") && strings.Contains(refType, ")") {
		// Extract arguments count from context
		// This is simplified - real implementation would parse properly
		if refSymbol.Kind == "function" || refSymbol.Kind == "method" { // Use Kind from model.Symbol
			// Could validate parameter count here
		}
	}
}

func (tv *TypeValidator) isImportUsed(imp *model.Import, references []*model.Reference) bool {
	// Check if any reference uses this import
	importName := imp.Path
	if strings.Contains(importName, "/") {
		// Get last part of import path
		parts := strings.Split(importName, "/")
		importName = parts[len(parts)-1]
	}

	for _, ref := range references {
		// Check if import is used in reference type
		if strings.Contains(ref.ReferenceType, importName) {
			return true
		}
	}

	// If no explicit usage found, assume it's used (to avoid false positives)
	return len(references) > 0
}
