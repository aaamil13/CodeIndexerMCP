package model

import "time"

// SearchOptions defines options for searching symbols
type SearchOptions struct {
	Query    string
	Language string
	Kind     string
	Limit    int
}

// FileStructure represents the structure of a file
type FileStructure struct {
	FilePath string
	Language string
	Symbols  []*Symbol
	Imports  []*Import
}

// ProjectOverview provides a summary of the project
type ProjectOverview struct {
	Project       *Project
	TotalFiles    int
	TotalSymbols  int
	LanguageStats map[string]int
}

// SymbolDetails provides detailed information about a symbol
type SymbolDetails struct {
	Symbol        *Symbol
	File          *File
	References    []*Reference
	Relationships []*Relationship
	Documentation string
}

// CodeContext provides comprehensive context for a symbol
type CodeContext struct {
	Symbol      *Symbol
	Snippet     string
	Usages      []*Reference
	Callees     []*Symbol
	Callers     []*Symbol
	Related     []*Symbol
	TestFiles   []*File
	Doc         string
	Definition  string
	Breadcrumbs []*Symbol
}

// ChangeImpact analyzes the impact of changing a symbol
type ChangeImpact struct {
	Symbol         *Symbol
	Dependents     []*Symbol
	References     []*Reference
	TestImpact     []*TestFile
	RiskScore      float64
	Recommendations []string
}

// TestFile represents a test file
type TestFile struct {
	Path    string
	Content string
}

// CodeMetrics calculates code quality metrics
type CodeMetrics struct {
	SymbolName      string
	Cyclomatic      int
	Maintainability float64
	Halstead        *HalsteadMetrics
	Cohesion        float64
	Coupling        float64
}

// HalsteadMetrics represents Halstead complexity metrics
type HalsteadMetrics struct {
	NumOperands  int
	NumOperators int
	NumUniqueOps int
	NumUniqueAnds int
	Volume       float64
	Difficulty   float64
	Effort       float64
}

// SmartSnippet represents a self-contained code snippet
type SmartSnippet struct {
	SymbolName   string
	Signature    string
	Body         string
	Dependencies []*Import
	Language     string
}

// SymbolUsageStats provides usage statistics for a symbol
type SymbolUsageStats struct {
	Symbol      *Symbol
	UsageCount  int
	FileCount   int
	LastUsed    time.Time
	SampleUsage *Reference
}

// RefactoringOpportunity represents a potential refactoring
type RefactoringOpportunity struct {
	Symbol      *Symbol
	Type        string // e.g., "extract_method", "rename"
	Description string
	Confidence  float64
}

// ChangeType defines the type of a change
type ChangeType string

const (
	ChangeTypeModify ChangeType = "modify"
	ChangeTypeAdd    ChangeType = "add"
	ChangeTypeDelete ChangeType = "delete"
)

// Change represents a single change to a symbol
type Change struct {
	SymbolName string
	Type       ChangeType
	OldValue   string
	NewValue   string
}

// ValidationResult represents the result of validating changes
type ValidationResult struct {
	IsValid bool
	Errors  []string
}

// ChangeImpactResult represents the result of a change impact analysis
type ChangeImpactResult struct {
	Change     *Change
	Impact     *ChangeImpact
	Validation *ValidationResult
}

// AutoFixSuggestion represents a suggestion for automatically fixing a change
type AutoFixSuggestion struct {
	Description string
	Action      string // e.g., "replace", "insert"
	FilePath    string
	Line        int
	Column      int
	OldCode     string
	NewCode     string
}

// DependencyGraph represents a dependency graph for a symbol
type DependencyGraph struct {
	Symbol       *Symbol
	Dependencies []*DependencyNode
	Dependents   []*DependencyNode
}

// DependencyNode represents a node in a dependency graph
type DependencyNode struct {
	Symbol *Symbol
	Depth  int
}

// TypeValidation represents the result of type validation for a file
type TypeValidation struct {
	FileID        int
	IsValid       bool
	Errors        []*TypeError
	SafetyScore   float64
	UndefinedRate float64
}

// TypeError represents a type error
type TypeError struct {
	FilePath string
	Line     int
	Column   int
	Message  string
}

// UndefinedUsage represents an undefined symbol usage
type UndefinedUsage struct {
	SymbolName string
	FilePath   string
	Line       int
	Column     int
}

// MissingMethod represents a missing method on a type
type MissingMethod struct {
	TypeName   string
	MethodName string
	FilePath   string
	Line       int
	Column     int
}

// TypeSafetyScore represents the type safety score for a file
type TypeSafetyScore struct {
	FileID      int
	Score       float64
	TotalChecks int
	FailedChecks int
}

// Project represents a project
type Project struct {
	ID            int
	Path          string
	Name          string
	LanguageStats map[string]int
	LastIndexed   time.Time
	CreatedAt     time.Time
}

// File represents a file in the project
type File struct {
	ID           int
	ProjectID    int
	Path         string
	RelativePath string
	Language     string
	Size         int64
	LinesOfCode  int
	Hash         string
	LastModified time.Time
	LastIndexed  time.Time
}
