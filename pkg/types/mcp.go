package types

// MCPTool represents an MCP tool definition
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// SearchOptions contains options for symbol search
type SearchOptions struct {
	Query       string      `json:"query"`
	ProjectID   int64       `json:"project_id"` // Added ProjectID
	Type        *SymbolType `json:"type,omitempty"`
	Language    string      `json:"language,omitempty"`
	FilePattern string      `json:"file_pattern,omitempty"`
	Limit       int         `json:"limit,omitempty"`
}

// FileStructure represents the structure of a file
type FileStructure struct {
	FilePath string    `json:"file_path"`
	Language string    `json:"language"`
	Symbols  []*Symbol `json:"symbols"`
	Imports  []*Import `json:"imports"`
}

// SymbolDetails contains detailed information about a symbol
type SymbolDetails struct {
	Symbol        *Symbol         `json:"symbol"`
	File          *File           `json:"file"`
	References    []*Reference    `json:"references"`
	Relationships []*Relationship `json:"relationships"`
	Documentation string          `json:"documentation,omitempty"`
}

// ComplexityMetrics contains code complexity metrics
type ComplexityMetrics struct {
	FilePath             string  `json:"file_path"`
	FunctionName         string  `json:"function_name,omitempty"`
	CyclomaticComplexity int     `json:"cyclomatic_complexity"`
	CognitiveComplexity  int     `json:"cognitive_complexity"`
	LinesOfCode          int     `json:"lines_of_code"`
	MaintainabilityIndex float64 `json:"maintainability_index"`
}

// ProjectOverview contains high-level project information
type ProjectOverview struct {
	Project          *Project       `json:"project"`
	TotalFiles       int            `json:"total_files"`
	TotalSymbols     int            `json:"total_symbols"`
	LanguageStats    map[string]int `json:"language_stats"`
	TopLevelSymbols  []*Symbol      `json:"top_level_symbols"`
	RecentlyModified []*File        `json:"recently_modified"`
}
