package types

// SemanticAnalysisResult contains comprehensive semantic analysis for a project
type SemanticAnalysisResult struct {
	ProjectID           int64
	TypeErrors          []*TypeMismatch
	UndefinedReferences []*UndefinedUsage
	UnusedSymbols       []*Symbol
	CircularDeps        []*CircularDependency
	Warnings            []string
	Metrics             map[string]interface{}
	QualityScore        float64 // 0-100
}

// SymbolTypeInference represents inferred type information for a symbol based on other symbols
type SymbolTypeInference struct {
	SymbolName   string
	InferredType string
	Confidence   float64 // 0.0 - 1.0
	Reasoning    string
	Sources      []*Symbol // Symbols used to infer the type
}

// CircularDependency represents a circular dependency in the codebase
type CircularDependency struct {
	Files       []string
	Description string
	Severity    string // "error", "warning", "info"
	Impact      string
	Suggestion  string
}

// CallGraph represents the function/method call graph for a project
type CallGraph struct {
	Nodes []*CallGraphNode
	Edges []*CallGraphEdge
}

// CallGraphNode represents a function/method in the call graph
type CallGraphNode struct {
	SymbolID   int64
	SymbolName string
	FilePath   string
	CallCount  int // Number of times this function is called
	IsExternal bool
	IsRecursive bool
}

// CallGraphEdge represents a call relationship between two functions
type CallGraphEdge struct {
	FromSymbolID int64
	ToSymbolID   int64
	CallSites    int // Number of call sites
	IsRecursive  bool
	IsConditional bool // Called within if/switch
}

// SemanticContext provides context for semantic operations
type SemanticContext struct {
	ProjectID     int64
	CurrentFile   *File
	CurrentSymbol *Symbol
	Scope         *Scope
	TypeCache     map[string]string // symbol name -> inferred type
}

// Scope represents a lexical scope in code
type Scope struct {
	ID          int64
	ParentScope *Scope
	Symbols     []*Symbol
	Type        string // "global", "class", "function", "block"
}

// FileDependencyEdge represents a dependency relationship between files or modules
type FileDependencyEdge struct {
	FromFile   *File
	ToFile     *File
	ImportPath string
	Type       string // "import", "require", "include", "using"
	IsCircular bool
}

// CodeQualityMetrics contains various code quality metrics
type CodeQualityMetrics struct {
	ProjectID            int64
	TotalFiles           int
	TotalLines           int
	TotalSymbols         int
	AverageComplexity    float64
	TypeSafetyScore      float64
	SemanticQualityScore float64
	TestCoverage         float64
	DuplicationRate      float64
	TechnicalDebtScore   float64
	Maintainability      string // "excellent", "good", "fair", "poor"
	Recommendations      []string
}

// CrossFileReference represents a reference that spans files
type CrossFileReference struct {
	Symbol         *Symbol
	ReferencedFrom []*Reference
	DefinedIn      *File
	ReferencedIn   []*File
	IsPublic       bool
	UsageCount     int
}
