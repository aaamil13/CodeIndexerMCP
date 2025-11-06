package ai

import (
	"fmt"
	"strings"
	"time"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// ChangeTracker tracks code changes and their impact
type ChangeTracker struct {
	db             *database.DB
	impactAnalyzer *ImpactAnalyzer
}

// NewChangeTracker creates a new change tracker
func NewChangeTracker(db *database.DB) *ChangeTracker {
	return &ChangeTracker{
		db:             db,
		impactAnalyzer: NewImpactAnalyzer(db),
	}
}

// AnalyzeSymbolChange analyzes the impact of changing a symbol
func (ct *ChangeTracker) AnalyzeSymbolChange(change *types.Change) (*types.ChangeImpactResult, error) {
	result := &types.ChangeImpactResult{
		Changes:            []*types.Change{change},
		AffectedSymbols:    []*types.Symbol{},
		AffectedFiles:      []*types.File{},
		BrokenReferences:   []*types.BrokenReference{},
		RequiredUpdates:    []*types.RequiredUpdate{},
		ValidationErrors:   []*types.ValidationError{},
		AutoFixSuggestions: []*types.AutoFixSuggestion{},
	}

	if change.Symbol == nil {
		return result, fmt.Errorf("change must have a symbol")
	}

	// Get impact analysis
	impact, err := ct.impactAnalyzer.AnalyzeChangeImpact(change.Symbol.Name)
	if err != nil {
		return nil, err
	}

	result.RiskLevel = impact.RiskLevel
	result.AffectedFiles = impact.AffectedFiles
	result.AffectedSymbols = impact.AffectedSymbols

	// Analyze based on change type
	switch change.Type {
	case types.ChangeTypeDelete:
		ct.analyzeDelete(change, impact, result)
	case types.ChangeTypeRename:
		ct.analyzeRename(change, impact, result)
	case types.ChangeTypeModify:
		ct.analyzeModify(change, impact, result)
	}

	// Determine if can auto-fix
	result.CanAutoFix = ct.canAutoFix(result)

	return result, nil
}

// analyzeDelete analyzes symbol deletion impact
func (ct *ChangeTracker) analyzeDelete(change *types.Change, impact *types.ChangeImpact, result *types.ChangeImpactResult) {
	// Get all references
	references, err := ct.db.GetReferencesBySymbol(change.Symbol.ID)
	if err != nil {
		return
	}

	// Each reference becomes a broken reference
	for _, ref := range references {
		file, _ := ct.db.GetFile(ref.FileID)

		broken := &types.BrokenReference{
			Reference:     ref,
			File:          file,
			Line:          ref.LineNumber,
			MissingSymbol: change.Symbol.Name,
			Reason:        fmt.Sprintf("Symbol '%s' was deleted", change.Symbol.Name),
			Severity:      "error",
		}
		result.BrokenReferences = append(result.BrokenReferences, broken)

		// Required update: remove or replace the reference
		update := &types.RequiredUpdate{
			File:      file,
			Line:      ref.LineNumber,
			Reason:    fmt.Sprintf("Remove usage of deleted symbol '%s'", change.Symbol.Name),
			Automatic: false, // Deletion requires manual intervention
		}
		result.RequiredUpdates = append(result.RequiredUpdates, update)

		// Validation error
		valError := &types.ValidationError{
			Type:     "reference",
			File:     file,
			Line:     ref.LineNumber,
			Message:  fmt.Sprintf("Reference to deleted symbol '%s'", change.Symbol.Name),
			Severity: "error",
		}
		result.ValidationErrors = append(result.ValidationErrors, valError)
	}
}

// analyzeRename analyzes symbol rename impact
func (ct *ChangeTracker) analyzeRename(change *types.Change, impact *types.ChangeImpact, result *types.ChangeImpactResult) {
	if change.OldSymbol == nil {
		return
	}

	oldName := change.OldSymbol.Name
	newName := change.Symbol.Name

	// Get all references
	references, err := ct.db.GetReferencesBySymbol(change.OldSymbol.ID)
	if err != nil {
		return
	}

	// Each reference needs to be updated
	for _, ref := range references {
		file, _ := ct.db.GetFile(ref.FileID)

		// Required update
		update := &types.RequiredUpdate{
			File:      file,
			Line:      ref.LineNumber,
			OldCode:   oldName,
			NewCode:   newName,
			Reason:    fmt.Sprintf("Update reference after rename from '%s' to '%s'", oldName, newName),
			Automatic: true, // Rename can be automatic
		}
		result.RequiredUpdates = append(result.RequiredUpdates, update)

		// Auto-fix suggestion
		suggestion := &types.AutoFixSuggestion{
			Type:        "rename",
			File:        file,
			LineStart:   ref.LineNumber,
			LineEnd:     ref.LineNumber,
			OldCode:     oldName,
			NewCode:     newName,
			Description: fmt.Sprintf("Rename '%s' to '%s'", oldName, newName),
			Confidence:  0.95,
			Safe:        true,
		}
		result.AutoFixSuggestions = append(result.AutoFixSuggestions, suggestion)
	}

	// Check if rename might cause conflicts
	if existingSymbol, _ := ct.db.GetSymbolByName(newName); existingSymbol != nil {
		valError := &types.ValidationError{
			Type:     "semantic",
			File:     change.File,
			Line:     change.LineStart,
			Message:  fmt.Sprintf("Symbol '%s' already exists - rename would cause conflict", newName),
			Severity: "error",
		}
		result.ValidationErrors = append(result.ValidationErrors, valError)
	}
}

// analyzeModify analyzes symbol modification impact
func (ct *ChangeTracker) analyzeModify(change *types.Change, impact *types.ChangeImpact, result *types.ChangeImpactResult) {
	// Check if signature changed
	if change.OldSymbol != nil && change.Symbol.Signature != change.OldSymbol.Signature {
		ct.analyzeSignatureChange(change, result)
	}

	// Check if visibility changed
	if change.OldSymbol != nil && change.Symbol.Visibility != change.OldSymbol.Visibility {
		ct.analyzeVisibilityChange(change, result)
	}

	// Check if exported status changed
	if change.OldSymbol != nil && change.Symbol.IsExported != change.OldSymbol.IsExported {
		ct.analyzeExportChange(change, result)
	}
}

// analyzeSignatureChange analyzes signature changes
func (ct *ChangeTracker) analyzeSignatureChange(change *types.Change, result *types.ChangeImpactResult) {
	// Get all call references
	references, err := ct.db.GetReferencesBySymbol(change.Symbol.ID)
	if err != nil {
		return
	}

	callRefs := []*types.Reference{}
	for _, ref := range references {
		if ref.ReferenceType == "call" {
			callRefs = append(callRefs, ref)
		}
	}

	if len(callRefs) > 0 {
		// Signature change affects all callers
		for _, ref := range callRefs {
			file, _ := ct.db.GetFile(ref.FileID)

			valError := &types.ValidationError{
				Type:     "semantic",
				File:     file,
				Line:     ref.LineNumber,
				Message:  fmt.Sprintf("Function signature changed for '%s' - caller may need updates", change.Symbol.Name),
				Severity: "warning",
			}
			result.ValidationErrors = append(result.ValidationErrors, valError)

			update := &types.RequiredUpdate{
				File:      file,
				Line:      ref.LineNumber,
				Reason:    fmt.Sprintf("Update call to '%s' to match new signature", change.Symbol.Name),
				Automatic: false,
			}
			result.RequiredUpdates = append(result.RequiredUpdates, update)
		}
	}
}

// analyzeVisibilityChange analyzes visibility changes
func (ct *ChangeTracker) analyzeVisibilityChange(change *types.Change, result *types.ChangeImpactResult) {
	oldVis := change.OldSymbol.Visibility
	newVis := change.Symbol.Visibility

	// If making more restrictive (public -> private)
	if oldVis == types.VisibilityPublic && newVis != types.VisibilityPublic {
		references, _ := ct.db.GetReferencesBySymbol(change.Symbol.ID)

		for _, ref := range references {
			file, _ := ct.db.GetFile(ref.FileID)

			// Check if reference is from outside the package/file
			if file.ID != change.File.ID {
				valError := &types.ValidationError{
					Type:     "semantic",
					File:     file,
					Line:     ref.LineNumber,
					Message:  fmt.Sprintf("'%s' is no longer accessible (visibility changed to %s)", change.Symbol.Name, newVis),
					Severity: "error",
				}
				result.ValidationErrors = append(result.ValidationErrors, valError)
			}
		}
	}
}

// analyzeExportChange analyzes export status changes
func (ct *ChangeTracker) analyzeExportChange(change *types.Change, result *types.ChangeImpactResult) {
	if change.OldSymbol.IsExported && !change.Symbol.IsExported {
		// Making unexported - breaking change for external users
		valError := &types.ValidationError{
			Type:     "semantic",
			File:     change.File,
			Line:     change.LineStart,
			Message:  fmt.Sprintf("Making '%s' unexported is a breaking API change", change.Symbol.Name),
			Severity: "error",
		}
		result.ValidationErrors = append(result.ValidationErrors, valError)
	}
}

// canAutoFix determines if changes can be automatically fixed
func (ct *ChangeTracker) canAutoFix(result *types.ChangeImpactResult) bool {
	// Can auto-fix only if:
	// 1. There are auto-fix suggestions
	// 2. No validation errors (only warnings are OK)
	// 3. All suggestions are marked as safe

	if len(result.AutoFixSuggestions) == 0 {
		return false
	}

	for _, err := range result.ValidationErrors {
		if err.Severity == "error" {
			return false
		}
	}

	for _, suggestion := range result.AutoFixSuggestions {
		if !suggestion.Safe {
			return false
		}
	}

	return true
}

// ValidateChanges validates a set of changes
func (ct *ChangeTracker) ValidateChanges(changes []*types.Change) (*types.ValidationResult, error) {
	result := &types.ValidationResult{
		ChangeSet: &types.ChangeSet{
			Changes:   changes,
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Errors:          []*types.ValidationError{},
		Warnings:        []*types.ValidationError{},
		Recommendations: []string{},
	}

	// Analyze each change
	for _, change := range changes {
		impactResult, err := ct.AnalyzeSymbolChange(change)
		if err != nil {
			continue
		}

		// Collect errors and warnings
		for _, valErr := range impactResult.ValidationErrors {
			if valErr.Severity == "error" {
				result.Errors = append(result.Errors, valErr)
			} else {
				result.Warnings = append(result.Warnings, valErr)
			}
		}

		// Store first impact result
		if result.Impact == nil {
			result.Impact = impactResult
		}
	}

	// Determine if valid
	result.IsValid = len(result.Errors) == 0
	result.CanProceed = result.IsValid || len(result.Errors) < 5 // Allow some errors

	// Generate recommendations
	if len(result.Errors) > 0 {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Fix %d error(s) before proceeding", len(result.Errors)))
	}
	if len(result.Warnings) > 0 {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Review %d warning(s)", len(result.Warnings)))
	}
	if result.Impact != nil && len(result.Impact.AutoFixSuggestions) > 0 {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Consider applying %d auto-fix suggestion(s)", len(result.Impact.AutoFixSuggestions)))
	}

	return result, nil
}

// GenerateAutoFixes generates automatic fixes for a change
func (ct *ChangeTracker) GenerateAutoFixes(change *types.Change) ([]*types.AutoFixSuggestion, error) {
	impactResult, err := ct.AnalyzeSymbolChange(change)
	if err != nil {
		return nil, err
	}

	return impactResult.AutoFixSuggestions, nil
}

// SimulateChange simulates a change without applying it
func (ct *ChangeTracker) SimulateChange(symbolName string, changeType types.ChangeType, newValue string) (*types.ChangeImpactResult, error) {
	symbol, err := ct.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	file, err := ct.db.GetFile(symbol.FileID)
	if err != nil {
		return nil, err
	}

	change := &types.Change{
		Type:      changeType,
		Symbol:    symbol,
		File:      file,
		LineStart: symbol.StartLine,
		LineEnd:   symbol.EndLine,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if changeType == types.ChangeTypeRename {
		oldSymbol := *symbol
		change.OldSymbol = &oldSymbol
		change.Symbol = &types.Symbol{
			ID:         symbol.ID,
			Name:       newValue,
			Type:       symbol.Type,
			Visibility: symbol.Visibility,
			IsExported: symbol.IsExported,
		}
		change.Description = fmt.Sprintf("Rename '%s' to '%s'", symbolName, newValue)
	}

	return ct.AnalyzeSymbolChange(change)
}
