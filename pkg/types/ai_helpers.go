package types

// CodeContext represents extracted code context for AI analysis
type CodeContext struct {
	Symbol            *Symbol            `json:"symbol"`
	File              *File              `json:"file"`
	Code              string             `json:"code"`               // The actual code
	Dependencies      []string           `json:"dependencies"`       // Required imports
	RelatedSymbols    []*Symbol          `json:"related_symbols"`    // Related symbols in same file
	Callers           []*Symbol          `json:"callers"`            // Who calls this
	Callees           []*Symbol          `json:"callees"`            // What this calls
	UsageExamples     []*UsageExample    `json:"usage_examples"`     // Examples of usage
	Documentation     string             `json:"documentation"`      // Full documentation
	Complexity        *ComplexityMetrics `json:"complexity"`         // Complexity metrics
	Context           string             `json:"context"`            // Surrounding context
}

// UsageExample represents an example of how a symbol is used
type UsageExample struct {
	FilePath    string `json:"file_path"`
	LineNumber  int    `json:"line_number"`
	Code        string `json:"code"`         // Code snippet showing usage
	Context     string `json:"context"`      // Surrounding context
	Description string `json:"description"`  // What this example shows
}

// ChangeImpact represents the impact of changing a symbol
type ChangeImpact struct {
	Symbol              *Symbol   `json:"symbol"`
	DirectReferences    int       `json:"direct_references"`     // Direct usages
	IndirectReferences  int       `json:"indirect_references"`   // Transitive usages
	AffectedFiles       []*File   `json:"affected_files"`        // Files that would be affected
	AffectedSymbols     []*Symbol `json:"affected_symbols"`      // Symbols that would be affected
	RiskLevel           string    `json:"risk_level"`            // low, medium, high
	Suggestions         []string  `json:"suggestions"`           // Refactoring suggestions
	BreakingChanges     bool      `json:"breaking_changes"`      // Would this break the API?
}

// CodeMetrics represents various code quality metrics
type CodeMetrics struct {
	FilePath              string  `json:"file_path"`
	FunctionName          string  `json:"function_name,omitempty"`
	LinesOfCode           int     `json:"lines_of_code"`
	CyclomaticComplexity  int     `json:"cyclomatic_complexity"`
	CognitiveComplexity   int     `json:"cognitive_complexity"`
	MaintainabilityIndex  float64 `json:"maintainability_index"`
	Parameters            int     `json:"parameters"`
	ReturnStatements      int     `json:"return_statements"`
	MaxNestingDepth       int     `json:"max_nesting_depth"`
	CommentDensity        float64 `json:"comment_density"`
	HasDocumentation      bool    `json:"has_documentation"`
	Quality               string  `json:"quality"` // excellent, good, fair, poor
}

// SmartSnippet represents a code snippet with all its dependencies
type SmartSnippet struct {
	Symbol        *Symbol  `json:"symbol"`
	Code          string   `json:"code"`           // Main code
	Dependencies  []string `json:"dependencies"`   // Required imports
	RelatedCode   []string `json:"related_code"`   // Related code snippets
	Documentation string   `json:"documentation"`
	UsageHints    []string `json:"usage_hints"`    // How to use this
	Complete      bool     `json:"complete"`       // Is this a complete, runnable snippet?
}

// SymbolUsageStats represents usage statistics for a symbol
type SymbolUsageStats struct {
	Symbol            *Symbol           `json:"symbol"`
	UsageCount        int               `json:"usage_count"`        // Total times used
	FileCount         int               `json:"file_count"`         // In how many files
	UsageByFile       map[string]int    `json:"usage_by_file"`      // Per-file usage
	CommonPatterns    []string          `json:"common_patterns"`    // Common usage patterns
	FirstUsed         string            `json:"first_used"`         // When first used
	LastUsed          string            `json:"last_used"`          // When last used
	IsDeprecated      bool              `json:"is_deprecated"`      // Is it marked deprecated?
	Alternatives      []string          `json:"alternatives"`       // Alternative symbols
}

// CodePattern represents a detected code pattern
type CodePattern struct {
	Name        string   `json:"name"`         // Pattern name (e.g., "Singleton", "Factory")
	Description string   `json:"description"`  // What this pattern does
	Symbols     []*Symbol `json:"symbols"`     // Symbols implementing this pattern
	Files       []*File  `json:"files"`        // Files containing this pattern
	Quality     string   `json:"quality"`      // How well it's implemented
	Suggestions []string `json:"suggestions"`  // Improvement suggestions
}

// RefactoringOpportunity represents a refactoring suggestion
type RefactoringOpportunity struct {
	Type        string   `json:"type"`         // Type of refactoring
	Symbol      *Symbol  `json:"symbol"`
	File        *File    `json:"file"`
	Description string   `json:"description"`  // What to refactor
	Reason      string   `json:"reason"`       // Why refactor
	Impact      string   `json:"impact"`       // Impact level
	Effort      string   `json:"effort"`       // Effort level (low, medium, high)
	Benefits    []string `json:"benefits"`     // Benefits of refactoring
	Risks       []string `json:"risks"`        // Potential risks
}

// NavigationHint represents a quick navigation suggestion for AI
type NavigationHint struct {
	From        string   `json:"from"`          // Starting point
	To          string   `json:"to"`            // Destination
	Path        []string `json:"path"`          // Navigation path
	Reason      string   `json:"reason"`        // Why navigate here
	Related     []string `json:"related"`       // Related locations
}

// CodeSimilarity represents similarity between code segments
type CodeSimilarity struct {
	Symbol1        *Symbol  `json:"symbol1"`
	Symbol2        *Symbol  `json:"symbol2"`
	SimilarityScore float64 `json:"similarity_score"` // 0.0 to 1.0
	CommonPatterns  []string `json:"common_patterns"`
	Differences     []string `json:"differences"`
	SuggestRefactor bool     `json:"suggest_refactor"`
}
