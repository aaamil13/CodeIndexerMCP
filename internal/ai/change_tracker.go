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
		return result, fmt.Errorf("must have a symbol")
	}

	// Get impact analysis
	impact, err := ct.impactAnalyzer.AnalyzeChangeImpact(change.Symbol.Name)
	if err != nil {
		return nil, err
	}

	result.RiskLevel = impact.RiskLevel
	result.AffectedFiles = impact.AffectedFiles
	result.AffectedSymbols = impact.AffectedSymbols

	// TODO: Implement analysis logic
	// // Analyze based on change type
	// switch change.Type {
	// case ChangeTypeDelete:
	// 	ct.analyzeDelete(change, impact, result)
	// case ChangeTypeRename:
	// 	ct.analyzeRename(change, impact, result)
	// case ChangeTypeModify:
	// 	ct.analyzeModify(change, impact, result)
	// }

	// Determine if can auto-fix
	result.CanAutoFix = ct.canAutoFix(result)

	return result, nil
}

// analyzeDelete analyzes symbol deletion impact
func (ct *ChangeTracker) analyzeDelete(change *model.Change, impact *model.ChangeImpact, result *model.ChangeImpact) {
	// TODO: Implement after DB methods are available
}

// analyzeRename analyzes symbol rename impact
func (ct *ChangeTracker) analyzeRename(change *model.Change, impact *model.ChangeImpact, result *model.ChangeImpact) {
	// TODO: Implement after DB methods are available
}

// analyzeModify analyzes symbol modification impact
func (ct *ChangeTracker) analyzeModify(change *model.Change, impact *model.ChangeImpact, result *model.ChangeImpact) {
	// Check if signature changed
	if change.OldSymbol != nil && change.Symbol.Signature != change.OldSymbol.Signature {
		ct.analyzeSignatureChange(change, result)
	}

	// Check if visibility changed
	if change.OldSymbol != nil && change.Symbol.Visibility != change.OldSymbol.Visibility {
		ct.analyzeVisibilityChange(change, result)
	}

	// Check if exported status changed
	isExportedOld := change.OldSymbol != nil && strings.ToUpper(change.OldSymbol.Name[0:1]) == change.OldSymbol.Name[0:1]
	isExportedNew := strings.ToUpper(change.Symbol.Name[0:1]) == change.Symbol.Name[0:1]
	if change.OldSymbol != nil && isExportedOld != isExportedNew {
		ct.analyzeExportChange(change, result)
	}
}

// analyzeSignatureChange analyzes signature changes
func (ct *ChangeTracker) analyzeSignatureChange(change *model.Change, result *model.ChangeImpact) {
	// TODO: Implement after DB methods are available
}

// analyzeVisibilityChange analyzes visibility changes
func (ct *ChangeTracker) analyzeVisibilityChange(change *model.Change, result *model.ChangeImpact) {
	// TODO: Implement after DB methods are available
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
	return nil, nil
}