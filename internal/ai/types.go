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
	Changes            []*model.Change
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
	Changes   []*model.Change
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

type DependencyGraph struct {
	Nodes []*DependencyNode
	Edges []*DependencyEdge
}

type DependencyNode struct {
	Symbol *model.Symbol
	File   string
	Type   string
	Level  int
}

type DependencyEdge struct {
	From   string
	To     string
	Type   string
	Weight float64
}

type RefactoringOpportunity struct {
	Type        string
	Symbol      *model.Symbol
	Description string
	Reason      string
	Impact      string
	Effort      string
	Benefits    []string
	Risks       []string
}

type CodeMetrics struct {
	FilePath             string
	FunctionName         string
	LinesOfCode          int
	CyclomaticComplexity int
	CognitiveComplexity  int
	MaintainabilityIndex float64
	Parameters           int
	ReturnStatements     int
	MaxNestingDepth      int
	CommentDensity       float64
	HasDocumentation     bool
	Quality              string
}

type SemanticAnalysisResult struct {
	ProjectID           string
	TypeErrors          []*TypeMismatch
	UndefinedReferences []*UndefinedUsage
	UnusedSymbols       []*model.Symbol
	CircularDeps        []*CircularDependency
	Warnings            []string
	Metrics             map[string]interface{}
	QualityScore        float64
}

type TypeMismatch struct {
	Symbol      *model.Symbol
	Expected    string
	Found       string
	FilePath    string
	Line        int
	Description string
}

type UndefinedUsage struct {
	SymbolName  string
	FilePath    string
	Line        int
	Description string
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

type SmartSnippet struct {
	Symbol        *model.Symbol
	Code          string
	Dependencies  []string
	RelatedCode   []string
	Documentation string
	UsageHints    []string
	Complete      bool
}

type TypeValidation struct {
	File             string
	Symbol           *model.Symbol
	IsValid          bool
	UndefinedSymbols []*UndefinedUsage
	TypeMismatches   []*TypeMismatch
	MissingMethods   []*MissingMethod
	InvalidCalls     []*InvalidCall
	UnusedImports    []*model.Import
	Suggestions      []string
}

type MissingMethod struct {
	TypeName         string
	MethodName       string
	AvailableMethods []string
	Suggestion       string
}

type InvalidCall struct {
	Symbol      *model.Symbol
	FilePath    string
	Line        int
	Description string
}

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

type SymbolUsageStats struct {
	Symbol         *model.Symbol
	UsageCount     int
	FileCount      int
	UsageByFile    map[string]int
	CommonPatterns []string
	IsDeprecated   bool
	Alternatives   []string
}








