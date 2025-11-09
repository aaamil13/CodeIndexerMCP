package ai

import (
	"fmt"
	"strconv"
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

	file, err := tv.db.GetFileByPath(0, filePath) // projectID is 0 for now as it's not available here
	if err != nil || file == nil {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	// Get all symbols in this file
	symbols, err := tv.db.GetSymbolsByFile(file.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols for file %s: %w", filePath, err)
	}

	// Collect all references for symbols in this file
	var references []*model.Reference
	for _, sym := range symbols {
		refs, err := tv.db.GetReferencesBySymbol(sym.ID)
		if err != nil {
			fmt.Printf("Warning: Failed to get references for symbol %d in file %s: %v\n", sym.ID, filePath, err)
			continue
		}
		references = append(references, refs...)
	}

	// Build symbol map for quick lookup
	symbolMap := make(map[string]*model.Symbol) // Map symbol name to symbol for faster lookup
	symbolIDMap := make(map[string]*model.Symbol) // Map symbol ID to symbol
	for _, sym := range symbols {
		symbolMap[sym.Name] = sym
		symbolIDMap[strconv.Itoa(sym.ID)] = sym // Convert int to string for map key
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
	imports, err := tv.db.GetImportsByFile(file.ID) // Use file.ID
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
	// For now, this is a placeholder. A proper implementation would need to:
	// 1. Get the project ID (currently passed as string) and convert to int.
	// 2. Search for the type symbol within that project.
	// 3. Find symbols whose parent is the typeName and whose kind is a method.

	// Placeholder: Assume typeSymbol is found for now
	// To avoid compilation errors, we'll return a placeholder error
	return &model.MissingMethod{
		TypeName:   typeName,
		MethodName: methodName,
		Suggestion: "Method existence check is not fully implemented yet.",
	}, nil
}

// ValidateSymbolTypes validates types for a specific symbol
func (tv *TypeValidator) ValidateSymbolTypes(symbolID string) (*model.TypeValidation, error) {
	symID, err := strconv.Atoi(symbolID)
	if err != nil {
		return nil, fmt.Errorf("invalid symbol ID: %w", err)
	}
	symbol, err := tv.db.GetSymbolByID(symID)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol %s: %w", symbolID, err)
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolID)
	}

	file, err := tv.db.GetFileByID(symbol.FileID)
	if err != nil || file == nil {
		return nil, fmt.Errorf("file not found for symbol %d: %w", symbol.ID, err)
	}

	validation := &model.TypeValidation{
		File:             file.Path, // Use file.Path from the retrieved file
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
		if ref.ReferenceType == model.ReferenceTypeCall { // Changed to model.ReferenceTypeCall
			// We need a way to get the target symbol by its name and the file it's in, or by its ID if it's stored
			// For now, GetSymbolByName is a simplified approach, but in reality, it needs more context.
			calledSymbol, err := tv.db.GetSymbolByName(ref.TargetSymbolName) // This needs projectID and potentially fileID for uniqueness
			if err != nil || calledSymbol == nil {
				// Called symbol doesn't exist - undefined usage
				undefined := &model.UndefinedUsage{
					SymbolName:  ref.TargetSymbolName,
					FilePath:    file.Path, // Use file.Path from the retrieved file
					Line:        ref.Line,
					Description: fmt.Sprintf("Function/method '%s' called by '%s' is undefined", ref.TargetSymbolName, symbol.Name),
				}
				validation.UndefinedSymbols = append(validation.UndefinedSymbols, undefined)
				validation.IsValid = false
			} else {
				// Basic type validation for calls
				if calledSymbol.Kind != model.SymbolKindFunction && calledSymbol.Kind != model.SymbolKindMethod {
					invalidCall := &model.InvalidCall{
						CallerSymbol: symbol,
						CalledSymbol: calledSymbol,
						FilePath:     file.Path, // Use file.Path from the retrieved file
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

	file, err := tv.db.GetFileByPath(0, filePath) // projectID is 0 for now as it's not available here
	if err != nil || file == nil {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	symbols, err := tv.db.GetSymbolsByFile(file.ID)
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
		// Simplified heuristic: A symbol is "typed" if it has a non-empty Signature,
		// or if it's a function/method (which inherently have return types/parameters that can be typed).
		if sym.Signature != "" || sym.Kind == model.SymbolKindFunction || sym.Kind == model.SymbolKindMethod {
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

func (tv *TypeValidator) inferUsageType(refType model.ReferenceType) string {
	// Convert model.ReferenceType to string for string operations
	refTypeStr := string(refType)
	if strings.Contains(refTypeStr, "call") { // Using "call" as a keyword for function/method calls
		return "function_or_method_call"
	}
	if strings.Contains(refTypeStr, "usage") {
		return "variable_usage"
	}
	if strings.Contains(refTypeStr, "import") {
		return "import_usage"
	}
	return "unknown_usage"
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
	// Convert model.ReferenceType to string for string operations
	refTypeStr := string(ref.ReferenceType)

	// Check if this is a method call on an object
	if strings.Contains(refTypeStr, ".") {
		parts := strings.Split(refTypeStr, ".")
		if len(parts) >= 2 {
			// This is potentially a method call
			// We'd need to check if the method exists on the type
			// For now, just note it
		}
	}

	// Check if this is a function call with wrong number of arguments
	if strings.Contains(refTypeStr, "(") && strings.Contains(refTypeStr, ")") {
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
		if strings.Contains(string(ref.ReferenceType), importName) { // Explicitly convert to string
			return true
		}
	}

	// If no explicit usage found, assume it's used (to avoid false positives)
	return len(references) > 0
}
