package ai

import (
	"fmt"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// TypeValidator validates types and finds undefined usages
type TypeValidator struct {
	db *database.DB
}

// NewTypeValidator creates a new type validator
func NewTypeValidator(db *database.DB) *TypeValidator {
	return &TypeValidator{
		db: db,
	}
}

// ValidateFile validates all types in a file
func (tv *TypeValidator) ValidateFile(fileID int64) (*types.TypeValidation, error) {
	file, err := tv.db.GetFile(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	validation := &types.TypeValidation{
		File:             file,
		IsValid:          true,
		UndefinedSymbols: make([]*types.UndefinedUsage, 0),
		TypeMismatches:   make([]*types.TypeMismatch, 0),
		MissingMethods:   make([]*types.MissingMethod, 0),
		InvalidCalls:     make([]*types.InvalidCall, 0),
		UnusedImports:    make([]*types.Import, 0),
		Suggestions:      make([]string, 0),
	}

	// Get all symbols in this file
	symbols, err := tv.db.GetSymbolsByFile(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols: %w", err)
	}

	// Get all references in this file
	references, err := tv.db.GetReferencesByFile(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %w", err)
	}

	// Build symbol map for quick lookup
	symbolMap := make(map[string]*types.Symbol)
	for _, sym := range symbols {
		symbolMap[sym.Name] = sym
	}

	// Check each reference
	for _, ref := range references {
		// Get the referenced symbol
		refSymbol, err := tv.db.GetSymbol(ref.SymbolID)
		if err != nil {
			// Symbol not found - undefined usage
			context := ref.ReferenceType

			undefined := &types.UndefinedUsage{
				Name:      "unknown",
				File:      file,
				Line:      ref.LineNumber,
				Column:    ref.ColumnNumber,
				Context:   context,
				UsageType: "unknown",
				Severity:  "error",
			}

			// Try to extract symbol name from context
			if context != "" {
				name := tv.extractSymbolNameFromContext(context)
				undefined.Name = name
				undefined.UsageType = tv.inferUsageType(context)

				// Find similar symbols (typo suggestions)
				similar := tv.findSimilarSymbols(name, symbols)
				if len(similar) > 0 {
					undefined.PossibleMatches = similar
					undefined.Suggestion = fmt.Sprintf("Did you mean '%s'?", similar[0].Name)
				}
			}

			validation.UndefinedSymbols = append(validation.UndefinedSymbols, undefined)
			validation.IsValid = false
			continue
		}

		// Validate the reference based on symbol type
		tv.validateReference(ref, refSymbol, file, validation)
	}

	// Check for unused imports
	imports, err := tv.db.GetImportsByFile(fileID)
	if err == nil {
		for _, imp := range imports {
			if !tv.isImportUsed(imp, references) {
				validation.UnusedImports = append(validation.UnusedImports, imp)
				validation.Suggestions = append(validation.Suggestions,
					fmt.Sprintf("Import '%s' is unused and can be removed", imp.Source))
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
func (tv *TypeValidator) FindUndefinedUsages(fileID int64) ([]*types.UndefinedUsage, error) {
	validation, err := tv.ValidateFile(fileID)
	if err != nil {
		return nil, err
	}
	return validation.UndefinedSymbols, nil
}

// CheckMethodExists checks if a method exists on a type
func (tv *TypeValidator) CheckMethodExists(typeName, methodName string, projectID int64) (*types.MissingMethod, error) {
	// Find the type symbol
	typeSymbol, err := tv.db.GetSymbolByName(typeName)
	if err != nil {
		return &types.MissingMethod{
			TypeName:   typeName,
			MethodName: methodName,
			Suggestion: fmt.Sprintf("Type '%s' not found", typeName),
		}, nil
	}

	// Find methods of this type
	methods, err := tv.db.GetMethodsForType(typeSymbol.ID)
	if err != nil {
		return nil, err
	}

	// Check if method exists
	for _, method := range methods {
		if method.Name == methodName {
			return nil, nil // Method exists
		}
	}

	// Method doesn't exist - build response
	missing := &types.MissingMethod{
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
	}

	return missing, nil
}

// ValidateSymbolTypes validates types for a specific symbol
func (tv *TypeValidator) ValidateSymbolTypes(symbolID int64) (*types.TypeValidation, error) {
	symbol, err := tv.db.GetSymbol(symbolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol: %w", err)
	}

	file, err := tv.db.GetFile(symbol.FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	validation := &types.TypeValidation{
		File:             file,
		Symbol:           symbol,
		IsValid:          true,
		UndefinedSymbols: make([]*types.UndefinedUsage, 0),
		TypeMismatches:   make([]*types.TypeMismatch, 0),
		MissingMethods:   make([]*types.MissingMethod, 0),
		InvalidCalls:     make([]*types.InvalidCall, 0),
		Suggestions:      make([]string, 0),
	}

	// Get all calls made by this symbol
	relationships, err := tv.db.GetRelationshipsForSymbol(symbolID)
	if err != nil {
		return validation, nil
	}

	for _, rel := range relationships {
		if rel.Type == types.RelationshipCalls {
			// Check if the called symbol exists
			_, err := tv.db.GetSymbol(rel.ToSymbolID)
			if err != nil {
				// Called symbol doesn't exist
				undefined := &types.UndefinedUsage{
					Name:      "unknown",
					File:      file,
					UsageType: "function",
					Severity:  "error",
				}
				validation.UndefinedSymbols = append(validation.UndefinedSymbols, undefined)
				validation.IsValid = false
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
func (tv *TypeValidator) CalculateTypeSafetyScore(fileID int64) (*types.TypeSafetyScore, error) {
	validation, err := tv.ValidateFile(fileID)
	if err != nil {
		return nil, err
	}

	symbols, err := tv.db.GetSymbolsByFile(fileID)
	if err != nil {
		return nil, err
	}

	score := &types.TypeSafetyScore{
		TotalSymbols:   len(symbols),
		TypedSymbols:   0,
		UntypedSymbols: 0,
		ErrorCount:     len(validation.UndefinedSymbols) + len(validation.TypeMismatches),
		WarningCount:   len(validation.InvalidCalls),
	}

	// Count typed vs untyped symbols
	for _, sym := range symbols {
		if sym.Signature != "" || sym.Type == types.SymbolTypeFunction || sym.Type == types.SymbolTypeMethod {
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

func (tv *TypeValidator) findSimilarSymbols(name string, symbols []*types.Symbol) []*types.Symbol {
	similar := make([]*types.Symbol, 0)

	for _, sym := range symbols {
		if tv.levenshteinDistance(name, sym.Name) <= 2 {
			similar = append(similar, sym)
		}
	}

	return similar
}

func (tv *TypeValidator) findSimilarMethodName(name string, methods []*types.Symbol) *types.Symbol {
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

func (tv *TypeValidator) validateReference(ref *types.Reference, refSymbol *types.Symbol, file *types.File, validation *types.TypeValidation) {
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
		if refSymbol.Type == types.SymbolTypeFunction || refSymbol.Type == types.SymbolTypeMethod {
			// Could validate parameter count here
		}
	}
}

func (tv *TypeValidator) isImportUsed(imp *types.Import, references []*types.Reference) bool {
	// Check if any reference uses this import
	importName := imp.Source
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
