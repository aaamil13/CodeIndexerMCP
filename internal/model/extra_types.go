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

type BrokenReference struct {
	Reference     *Reference
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

// ChangeImpact analyzes the impact of changing a symbol
type ChangeImpact struct {
	Changes            []*Change
	AffectedSymbols    []*Symbol
	AffectedFiles      []string
	BrokenReferences   []*BrokenReference
	RequiredUpdates    []*RequiredUpdate
	ValidationErrors   []*ValidationError
	AutoFixSuggestions []*AutoFixSuggestion
	RiskLevel          float64
	CanAutoFix         bool
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
	Symbol        *Symbol
	Code          string
	Dependencies  []string
	RelatedCode   []string
	Documentation string
	UsageHints    []string
	Complete      bool
}

// SymbolUsageStats provides usage statistics for a symbol
type SymbolUsageStats struct {
	Symbol         *Symbol
	UsageCount     int
	FileCount      int
	UsageByFile    map[string]int
	CommonPatterns []string
	IsDeprecated   bool
	Alternatives   []string
}

// RefactoringOpportunity represents a potential refactoring
type RefactoringOpportunity struct {
	Type        string
	Symbol      *Symbol
	Description string
	Reason      string
	Impact      string
	Effort      string
	Benefits    []string
	Risks       []string
}

// ChangeType defines the type of a change
type ChangeType string

const (
	ChangeTypeModify ChangeType = "modify"
	ChangeTypeAdd    ChangeType = "add"
	ChangeTypeDelete ChangeType = "delete"
	ChangeTypeRename ChangeType = "rename"
	ChangeTypeMove   ChangeType = "move"
)

// Change represents a single change to a symbol
type Change struct {
	Type         ChangeType
	File         string
	LineStart    int
	LineEnd      int
	Symbol       *Symbol
	OldSymbol    *Symbol
	NewContent   string
	Relationship string
}

// ValidationResult represents the result of validating changes
type ValidationResult struct {
	ChangeSet       *ChangeSet
	Errors          []*ValidationError
	Warnings        []*ValidationError
	Recommendations []string
	Impact          *ChangeImpact
	IsValid         bool
	CanProceed      bool
}

// ChangeSet represents a set of changes
type ChangeSet struct {
	Changes   []*Change
	Timestamp string
}

// DependencyGraph represents a dependency graph for a symbol
type DependencyGraph struct {
	Nodes              []*DependencyNode
	Edges              []*DependencyEdge
	DirectDependencies int
	DirectDependents   int
	CouplingScore      float64
}

// DependencyNode represents a node in a dependency graph
type DependencyNode struct {
	Symbol *Symbol
	File   string
	Type   string
	Level  int
}

// DependencyEdge represents an edge in a dependency graph
type DependencyEdge struct {
	From   string
	To     string
	Type   string
	Weight float64
}

// TypeValidation represents the result of type validation for a file
type TypeValidation struct {
	File             string
	Symbol           *Symbol
	IsValid          bool
	UndefinedSymbols []*UndefinedUsage
	TypeMismatches   []*TypeMismatch
	MissingMethods   []*MissingMethod
	InvalidCalls     []*InvalidCall
	UnusedImports    []*Import
	Suggestions      []string
}

// TypeMismatch represents a type error
type TypeMismatch struct {
	Symbol      *Symbol
	Expected    string
	Found       string
	FilePath    string
	Line        int
	Description string
}

// UndefinedUsage represents an undefined symbol usage
type UndefinedUsage struct {
	SymbolName  string
	FilePath    string
	Line        int
	Description string
}

// MissingMethod represents a missing method on a type
type MissingMethod struct {
	TypeName         string
	MethodName       string
	AvailableMethods []string
	Suggestion       string
}

// InvalidCall represents an invalid function call
type InvalidCall struct {
	Symbol      *Symbol
	FilePath    string
	Line        int
	Description string
}

// TypeSafetyScore represents the type safety score for a file
type TypeSafetyScore struct {
	TotalSymbols   int
	TypedSymbols   int
	UntypedSymbols int
	ErrorCount     int
	WarningCount   int
	Score          float64
	Rating         string
	Recommendation string
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

type UsageExample struct {
	FilePath    string
	LineNumber  int
	Code        string
	Context     string
	Description string
}

type SemanticAnalysisResult struct {
	ProjectID           string
	TypeErrors          []*TypeMismatch
	UndefinedReferences []*UndefinedUsage
	UnusedSymbols       []*Symbol
	CircularDeps        []*CircularDependency
	Warnings            []string
	Metrics             map[string]interface{}
	QualityScore        float64
}

type TypeInference struct {
	SymbolName   string
	InferredType string
	Confidence   float64
	Reasoning    string
}

type CircularDependency struct {
	Files       []string
	Description string
	Severity    string
}

type CallGraph struct {
	Nodes []*CallGraphNode
	Edges []*CallGraphEdge
}

type CallGraphNode struct {
	SymbolID   string
	SymbolName string
	FilePath   string
	CallCount  int
}

type CallGraphEdge struct {
	FromSymbolID string
	ToSymbolID   string
	CallSites    int
}

type CodeContext struct {
	Symbol         *Symbol
	File           string
	Code           string
	Dependencies   []string
	RelatedSymbols []*Symbol
	Callers        []*Symbol
	Callees        []*Symbol
	UsageExamples  []*UsageExample
	Documentation  string
	Context        string
}
