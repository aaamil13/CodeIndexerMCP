package ai

import (
	"fmt"
	"time"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

// ChangeTracker tracks code changes and their impact
type ChangeTracker struct {
	db             *database.Manager
	impactAnalyzer *ImpactAnalyzer
}

// NewChangeTracker creates a new change tracker
func NewChangeTracker(db *database.Manager) *ChangeTracker {
	return &ChangeTracker{
		db:             db,
		impactAnalyzer: NewImpactAnalyzer(db),
	}
}

// AnalyzeSymbolChange analyzes the impact of changing a symbol
func (ct *ChangeTracker) AnalyzeSymbolChange(change *model.Change) (*model.ChangeImpact, error) {
	result := &model.ChangeImpact{
		Changes:            []*model.Change{change},
		AffectedSymbols:    []*model.Symbol{},
		AffectedFiles:      []string{},
		BrokenReferences:   []*model.BrokenReference{},
		RequiredUpdates:    []*model.RequiredUpdate{},
		ValidationErrors:   []*model.ValidationError{},
		AutoFixSuggestions: []*model.AutoFixSuggestion{},
	}

	if change.Symbol == nil {
		return result, fmt.Errorf("change must contain a symbol")
	}

	// Get the full symbol details
	existingSymbol, err := ct.db.GetSymbolByID(change.Symbol.ID)
	if err != nil {
		return result, fmt.Errorf("failed to get existing symbol details for %d: %w", change.Symbol.ID, err)
	}
	if existingSymbol == nil {
		// If symbol not found in DB, it might be a new symbol being added.
		// For now, assume the change.Symbol itself is the "full" symbol.
		existingSymbol = change.Symbol
	}

	// For now, we'll use the existingSymbol directly as the "fullSymbol" for impact analysis,
	// as GetFunctionDetails and GetClassDetails are no longer available.
	// Future improvements might involve reconstructing model.Function or model.Class
	// from the base model.Symbol and its metadata.
	_ = existingSymbol // This variable will be used in a later step if needed.


	// Perform impact analysis based on change type
	switch change.Type {
	case model.ChangeTypeDelete:
		ct.analyzeDelete(change, result)
	case model.ChangeTypeRename:
		ct.analyzeRename(change, result)
	case model.ChangeTypeModify:
		ct.analyzeModify(change, result)
	}

	// Determine if can auto-fix
	result.CanAutoFix = ct.canAutoFix(result)

	return result, nil
}

// analyzeDelete analyzes symbol deletion impact
func (ct *ChangeTracker) analyzeDelete(change *model.Change, result *model.ChangeImpact) {
	references, err := ct.db.GetReferencesBySymbol(change.Symbol.ID)
	if err != nil {
		result.ValidationErrors = append(result.ValidationErrors, &model.ValidationError{
			Type:     "internal_error",
			File:     change.File,
			Line:     change.LineStart,
			Message:  fmt.Sprintf("Failed to get references for deletion analysis: %v", err),
			Severity: "error",
		})
		return
	}

	for _, ref := range references {
		result.BrokenReferences = append(result.BrokenReferences, &model.BrokenReference{
			ReferencingFile: ref.FilePath,
			ReferencingLine: ref.Line,
			SymbolName:      ref.TargetSymbolName,
			Problem:         fmt.Sprintf("Reference to deleted symbol '%s'", change.Symbol.Name),
		})
		result.ValidationErrors = append(result.ValidationErrors, &model.ValidationError{
			Type:     "breaking_change",
			File:     ref.FilePath,
			Line:     ref.Line,
			Message:  fmt.Sprintf("Symbol '%s' is being deleted, affecting reference in %s:%d", change.Symbol.Name, ref.FilePath, ref.Line),
			Severity: "error",
		})
		result.AffectedFiles = append(result.AffectedFiles, ref.FilePath)
	}
}

// analyzeRename analyzes symbol rename impact
func (ct *ChangeTracker) analyzeRename(change *model.Change, result *model.ChangeImpact) {
	if change.OldSymbol == nil || change.NewSymbol == nil {
		return // Not a valid rename change
	}

	references, err := ct.db.GetReferencesBySymbol(change.OldSymbol.ID)
	if err != nil {
		result.ValidationErrors = append(result.ValidationErrors, &model.ValidationError{
			Type:     "internal_error",
			File:     change.File,
			Line:     change.LineStart,
			Message:  fmt.Sprintf("Failed to get references for rename analysis: %v", err),
			Severity: "error",
		})
		return
	}

	for _, ref := range references {
		// Suggest changing the reference to the new symbol name
		result.AutoFixSuggestions = append(result.AutoFixSuggestions, &model.AutoFixSuggestion{
			FilePath: ref.FilePath,
			Line:     ref.Line,
			Column:   ref.Column,
			OldText:  change.OldSymbol.Name,
			NewText:  change.NewSymbol.Name,
			Message:  fmt.Sprintf("Rename reference to '%s' to '%s'", change.OldSymbol.Name, change.NewSymbol.Name),
			Safe:     true, // Renames are generally safe if all references are updated
		})
		result.AffectedFiles = append(result.AffectedFiles, ref.FilePath)
	}
}

// analyzeModify analyzes symbol modification impact
func (ct *ChangeTracker) analyzeModify(change *model.Change, result *model.ChangeImpact) {
	if change.OldSymbol == nil || change.NewSymbol == nil {
		return // Not a valid modify change
	}

	// Check if signature changed
	if change.NewSymbol.Signature != change.OldSymbol.Signature {
		ct.analyzeSignatureChange(change, result)
	}

	// Check if visibility changed
	if change.NewSymbol.Visibility != change.OldSymbol.Visibility {
		ct.analyzeVisibilityChange(change, result)
	}

	// Check if exported status changed
	isExportedOld := strings.ToUpper(change.OldSymbol.Name[0:1]) == change.OldSymbol.Name[0:1]
	isExportedNew := strings.ToUpper(change.NewSymbol.Name[0:1]) == change.NewSymbol.Name[0:1]
	if isExportedOld != isExportedNew {
		ct.analyzeExportChange(change, result)
	}
}

// analyzeSignatureChange analyzes signature changes
func (ct *ChangeTracker) analyzeSignatureChange(change *model.Change, result *model.ChangeImpact) {
	references, err := ct.db.GetReferencesBySymbol(change.NewSymbol.ID)
	if err != nil {
		result.ValidationErrors = append(result.ValidationErrors, &model.ValidationError{
			Type:     "internal_error",
			File:     change.File,
			Line:     change.LineStart,
			Message:  fmt.Sprintf("Failed to get references for signature change analysis: %v", err),
			Severity: "error",
		})
		return
	}

	for _, ref := range references {
		// This is a simplified check. A full analysis would compare old and new parameters, return types, etc.
		result.ValidationErrors = append(result.ValidationErrors, &model.ValidationError{
			Type:     "breaking_change",
			File:     ref.FilePath,
			Line:     ref.Line,
			Message:  fmt.Sprintf("Signature of '%s' changed from '%s' to '%s', affecting call in %s:%d", change.NewSymbol.Name, change.OldSymbol.Signature, change.NewSymbol.Signature, ref.FilePath, ref.Line),
			Severity: "warning", // Could be error depending on severity of change
		})
		result.AffectedFiles = append(result.AffectedFiles, ref.FilePath)
	}
}

// analyzeVisibilityChange analyzes visibility changes
func (ct *ChangeTracker) analyzeVisibilityChange(change *model.Change, result *model.ChangeImpact) {
	if change.NewSymbol.Visibility == model.VisibilityPrivate && change.OldSymbol.Visibility != model.VisibilityPrivate {
		// If a symbol becomes private, any external references will be broken
		references, err := ct.db.GetReferencesBySymbol(change.NewSymbol.ID)
		if err != nil {
			result.ValidationErrors = append(result.ValidationErrors, &model.ValidationError{
				Type:     "internal_error",
				File:     change.File,
				Line:     change.LineStart,
				Message:  fmt.Sprintf("Failed to get references for visibility change analysis: %v", err),
				Severity: "error",
			})
			return
		}

		for _, ref := range references {
			result.BrokenReferences = append(result.BrokenReferences, &model.BrokenReference{
				ReferencingFile: ref.FilePath,
				ReferencingLine: ref.Line,
				SymbolName:      ref.TargetSymbolName,
				Problem:         fmt.Sprintf("Reference to symbol '%s' which became private", change.NewSymbol.Name),
			})
			result.ValidationErrors = append(result.ValidationErrors, &model.ValidationError{
				Type:     "breaking_change",
				File:     ref.FilePath,
				Line:     ref.Line,
				Message:  fmt.Sprintf("Symbol '%s' changed to private, affecting reference in %s:%d", change.NewSymbol.Name, ref.FilePath, ref.Line),
				Severity: "error",
			})
			result.AffectedFiles = append(result.AffectedFiles, ref.FilePath)
		}
	}
}

// analyzeExportChange analyzes export status changes
func (ct *ChangeTracker) analyzeExportChange(change *model.Change, result *model.ChangeImpact) {
	isExportedOld := change.OldSymbol != nil && strings.ToUpper(change.OldSymbol.Name[0:1]) == change.OldSymbol.Name[0:1]
	isExportedNew := strings.ToUpper(change.Symbol.Name[0:1]) == change.Symbol.Name[0:1]

	if isExportedOld && !isExportedNew {
		// Making unexported - breaking change for external users
		valError := &model.ValidationError{
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
func (ct *ChangeTracker) canAutoFix(result *model.ChangeImpact) bool {
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
func (ct *ChangeTracker) ValidateChanges(changes []*model.Change) (*model.ValidationResult, error) {
	result := &model.ValidationResult{
		ChangeSet: &model.ChangeSet{
			Changes:   changes,
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Errors:          []*model.ValidationError{},
		Warnings:        []*model.ValidationError{},
		Recommendations: []string{},
	}

	// TODO: Implement after DB methods are available
	// // Analyze each change
	// for _, change := range changes {
	// 	impactResult, err := ct.AnalyzeSymbolChange(change)
	// 	if err != nil {
	// 		continue
	// 	}

	// 	// Collect errors and warnings
	// 	for _, valErr := range impactResult.ValidationErrors {
	// 		if valErr.Severity == "error" {
	// 			result.Errors = append(result.Errors, valErr)
	// 		} else {
	// 			result.Warnings = append(result.Warnings, valErr)
	// 		}
	// 	}

	// 	// Store first impact result
	// 	if result.Impact == nil {
	// 		result.Impact = impactResult
	// 	}
	// }

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
func (ct *ChangeTracker) GenerateAutoFixes(change *model.Change) ([]*model.AutoFixSuggestion, error) {
	// TODO: Implement after DB methods are available
	// impactResult, err := ct.AnalyzeSymbolChange(change)
	// if err != nil {
	// 	return nil, err
	// }

	// return impactResult.AutoFixSuggestions, nil
	return nil, nil
}

// SimulateChange simulates a change without applying it
func (ct *ChangeTracker) SimulateChange(symbolName string, changeType model.ChangeType, newValue string) (*model.ChangeImpact, error) {
	// TODO: Implement after DB methods are available
	return nil, fmt.Errorf("not implemented")
}
