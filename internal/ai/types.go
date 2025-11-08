package ai

import (
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

type ChangeType string

const (
	ChangeTypeModify ChangeType = "modify"
	ChangeTypeRename ChangeType = "rename"
	ChangeTypeDelete ChangeType = "delete"
)

type Visibility string

const (
	VisibilityPublic    Visibility = "public"
	VisibilityPrivate   Visibility = "private"
	VisibilityProtected Visibility = "protected"
	VisibilityInternal  Visibility = "internal"
	VisibilityPackage   Visibility = "package" // For Java package-private
)

type Change struct {
	Type        ChangeType
	Symbol      *model.Symbol
	OldSymbol   *model.Symbol
	File        string
	LineStart   int
	LineEnd     int
	Timestamp   string
	Description string
}

type BrokenReference struct {
	Reference     *model.Reference
	File          string
	Line          int
	MissingSymbol string
	Reason        string
	Severity      string
}

type RequiredUpdate struct {
	File      string
	Line      int
	OldCode   string
	NewCode   string
	Reason    string
	Automatic bool
}

type ValidationError struct {
	Type     string
	File     string
	Line     int
	Message  string
	Severity string
}

type AutoFixSuggestion struct {
	Type        string
	File        string
	LineStart   int
	LineEnd     int
	OldCode     string
	NewCode     string
	Description string
	Confidence  float64
	Safe        bool
}

type ChangeImpact struct {
	RiskLevel       float64
	AffectedFiles   []string
	AffectedSymbols []*model.Symbol
}

type ChangeImpactResult struct {
	Changes            []*Change
	AffectedSymbols    []*model.Symbol
	AffectedFiles      []string
	BrokenReferences   []*BrokenReference
	RequiredUpdates    []*RequiredUpdate
	ValidationErrors   []*ValidationError
	AutoFixSuggestions []*AutoFixSuggestion
	RiskLevel          float64
	CanAutoFix         bool
}

type ValidationResult struct {
	ChangeSet       *ChangeSet
	Errors          []*ValidationError
	Warnings        []*ValidationError
	Recommendations []string
	Impact          *ChangeImpactResult
	IsValid         bool
	CanProceed      bool
}

type ChangeSet struct {
	Changes   []*Change
	Timestamp string
}

type CodeContext struct {
	Symbol         *model.Symbol
	File           string
	Code           string
	Dependencies   []string
	RelatedSymbols []*model.Symbol
	Callers        []*model.Symbol
	Callees        []*model.Symbol
	UsageExamples  []*UsageExample
	Documentation  string
	Context        string
}

type UsageExample struct {
	FilePath    string
	LineNumber  int
	Code        string
	Context     string
	Description string
}

