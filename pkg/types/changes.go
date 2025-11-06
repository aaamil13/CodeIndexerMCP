package types

// ChangeType represents the type of change
type ChangeType string

const (
	ChangeTypeAdd    ChangeType = "add"
	ChangeTypeModify ChangeType = "modify"
	ChangeTypeDelete ChangeType = "delete"
	ChangeTypeRename ChangeType = "rename"
	ChangeTypeMove   ChangeType = "move"
)

// Change represents a code change
type Change struct {
	Type          ChangeType `json:"type"`
	Symbol        *Symbol    `json:"symbol,omitempty"`
	OldSymbol     *Symbol    `json:"old_symbol,omitempty"`     // For renames/modifications
	File          *File      `json:"file"`
	LineStart     int        `json:"line_start"`
	LineEnd       int        `json:"line_end"`
	OldCode       string     `json:"old_code,omitempty"`
	NewCode       string     `json:"new_code,omitempty"`
	Timestamp     string     `json:"timestamp"`
	Description   string     `json:"description"`
}

// ChangeImpactResult represents the result of analyzing changes
type ChangeImpactResult struct {
	Changes            []*Change            `json:"changes"`
	AffectedSymbols    []*Symbol            `json:"affected_symbols"`
	AffectedFiles      []*File              `json:"affected_files"`
	BrokenReferences   []*BrokenReference   `json:"broken_references"`
	RequiredUpdates    []*RequiredUpdate    `json:"required_updates"`
	ValidationErrors   []*ValidationError   `json:"validation_errors"`
	AutoFixSuggestions []*AutoFixSuggestion `json:"auto_fix_suggestions"`
	RiskLevel          string               `json:"risk_level"`
	CanAutoFix         bool                 `json:"can_auto_fix"`
}

// BrokenReference represents a broken reference after a change
type BrokenReference struct {
	Reference     *Reference `json:"reference"`
	File          *File      `json:"file"`
	Line          int        `json:"line"`
	MissingSymbol string     `json:"missing_symbol"`
	Reason        string     `json:"reason"`
	Severity      string     `json:"severity"` // error, warning, info
}

// RequiredUpdate represents a required update to fix broken references
type RequiredUpdate struct {
	File        *File  `json:"file"`
	Line        int    `json:"line"`
	OldCode     string `json:"old_code"`
	NewCode     string `json:"new_code"`
	Reason      string `json:"reason"`
	Automatic   bool   `json:"automatic"` // Can be auto-fixed?
}

// ValidationError represents a validation error
type ValidationError struct {
	Type     string `json:"type"`      // syntax, semantic, reference, type
	File     *File  `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Message  string `json:"message"`
	Severity string `json:"severity"`  // error, warning
	Code     string `json:"code,omitempty"`
}

// AutoFixSuggestion represents an automatic fix suggestion
type AutoFixSuggestion struct {
	Type        string   `json:"type"`        // rename, update_import, fix_reference
	File        *File    `json:"file"`
	LineStart   int      `json:"line_start"`
	LineEnd     int      `json:"line_end"`
	OldCode     string   `json:"old_code"`
	NewCode     string   `json:"new_code"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`  // 0.0 to 1.0
	Safe        bool     `json:"safe"`        // Is this safe to apply automatically?
}

// DependencyGraph represents the dependency graph of the project
type DependencyGraph struct {
	Nodes []*DependencyNode `json:"nodes"`
	Edges []*DependencyEdge `json:"edges"`
}

// DependencyNode represents a node in the dependency graph
type DependencyNode struct {
	Symbol       *Symbol `json:"symbol"`
	File         *File   `json:"file"`
	Type         string  `json:"type"` // file, symbol, module
	Level        int     `json:"level"` // Distance from root
}

// DependencyEdge represents an edge in the dependency graph
type DependencyEdge struct {
	From         string `json:"from"`         // Symbol or file ID
	To           string `json:"to"`           // Symbol or file ID
	Type         string `json:"type"`         // imports, calls, uses, extends
	Weight       int    `json:"weight"`       // Usage count
}

// ChangeSet represents a set of related changes
type ChangeSet struct {
	ID          string    `json:"id"`
	Changes     []*Change `json:"changes"`
	Author      string    `json:"author,omitempty"`
	Timestamp   string    `json:"timestamp"`
	Description string    `json:"description"`
	Branch      string    `json:"branch,omitempty"`
}

// ValidationResult represents the result of validating a changeset
type ValidationResult struct {
	ChangeSet       *ChangeSet          `json:"changeset"`
	IsValid         bool                `json:"is_valid"`
	Errors          []*ValidationError  `json:"errors"`
	Warnings        []*ValidationError  `json:"warnings"`
	Impact          *ChangeImpactResult `json:"impact"`
	CanProceed      bool                `json:"can_proceed"`
	Recommendations []string            `json:"recommendations"`
}
