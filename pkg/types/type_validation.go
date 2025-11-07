package types

// TypeInfo represents type information for a symbol
type TypeInfo struct {
	ID           int64  `json:"id"`
	SymbolID     int64  `json:"symbol_id"`
	TypeName     string `json:"type_name"`      // e.g., "int", "string", "User"
	IsPointer    bool   `json:"is_pointer"`     // e.g., *User
	IsArray      bool   `json:"is_array"`       // e.g., []string
	IsMap        bool   `json:"is_map"`         // e.g., map[string]int
	KeyType      string `json:"key_type"`       // For maps
	ValueType    string `json:"value_type"`     // For arrays/maps
	IsInterface  bool   `json:"is_interface"`   // Interface type
	IsFunction   bool   `json:"is_function"`    // Function type
	GenericTypes []string `json:"generic_types"` // Generic type parameters
}

// MethodSignature represents a method signature with type information
type MethodSignature struct {
	SymbolID    int64             `json:"symbol_id"`
	Parameters  []*ParameterInfo  `json:"parameters"`
	ReturnTypes []*TypeInfo       `json:"return_types"`
	ReceiverType *TypeInfo        `json:"receiver_type,omitempty"` // For methods
}

// ParameterInfo represents a function/method parameter
type ParameterInfo struct {
	Name     string    `json:"name"`
	Type     *TypeInfo `json:"type"`
	Optional bool      `json:"optional"`
	Variadic bool      `json:"variadic"` // ... parameter
}

// TypeValidation represents validation results for types
type TypeValidation struct {
	File              *File                  `json:"file"`
	Symbol            *Symbol                `json:"symbol,omitempty"`
	IsValid           bool                   `json:"is_valid"`
	UndefinedSymbols  []*UndefinedUsage      `json:"undefined_symbols"`
	TypeMismatches    []*TypeMismatch        `json:"type_mismatches"`
	MissingMethods    []*MissingMethod       `json:"missing_methods"`
	InvalidCalls      []*InvalidCall         `json:"invalid_calls"`
	UnusedImports     []*Import              `json:"unused_imports"`
	Suggestions       []string               `json:"suggestions"`
}

// UndefinedUsage represents usage of an undefined symbol
type UndefinedUsage struct {
	Name           string    `json:"name"`
	File           *File     `json:"file"`
	Line           int       `json:"line"`
	Column         int       `json:"column"`
	Context        string    `json:"context"`
	UsageType      string    `json:"usage_type"` // "function", "variable", "method", "type"
	Severity       string    `json:"severity"`   // "error", "warning"
	Suggestion     string    `json:"suggestion,omitempty"`
	PossibleMatches []*Symbol `json:"possible_matches,omitempty"` // Similar symbols that exist
}

// TypeMismatch represents a type mismatch error
type TypeMismatch struct {
	File          *File     `json:"file"`
	Line          int       `json:"line"`
	Column        int       `json:"column"`
	Context       string    `json:"context"`
	ExpectedType  string    `json:"expected_type"`
	ActualType    string    `json:"actual_type"`
	Severity      string    `json:"severity"` // "error", "warning"
	Description   string    `json:"description"`
	Suggestion    string    `json:"suggestion,omitempty"`
}

// MissingMethod represents a method that doesn't exist on a type
type MissingMethod struct {
	TypeName       string    `json:"type_name"`
	MethodName     string    `json:"method_name"`
	File           *File     `json:"file"`
	Line           int       `json:"line"`
	Column         int       `json:"column"`
	Context        string    `json:"context"`
	AvailableMethods []string `json:"available_methods,omitempty"`
	Suggestion     string    `json:"suggestion,omitempty"`
}

// InvalidCall represents an invalid function/method call
type InvalidCall struct {
	SymbolName     string    `json:"symbol_name"`
	File           *File     `json:"file"`
	Line           int       `json:"line"`
	Column         int       `json:"column"`
	Context        string    `json:"context"`
	Issue          string    `json:"issue"` // "wrong_param_count", "wrong_param_type", "not_callable"
	Expected       string    `json:"expected"`
	Actual         string    `json:"actual"`
	Severity       string    `json:"severity"`
	Suggestion     string    `json:"suggestion,omitempty"`
}

// TypeCheckResult represents comprehensive type checking results for a file or project
type TypeCheckResult struct {
	Files             []*File            `json:"files"`
	TotalErrors       int                `json:"total_errors"`
	TotalWarnings     int                `json:"total_warnings"`
	UndefinedCount    int                `json:"undefined_count"`
	TypeMismatchCount int                `json:"type_mismatch_count"`
	InvalidCallCount  int                `json:"invalid_call_count"`
	Validations       []*TypeValidation  `json:"validations"`
	Summary           string             `json:"summary"`
	IsTypeSafe        bool               `json:"is_type_safe"`
}

// TypeInference represents inferred type information
type TypeInference struct {
	SymbolName    string    `json:"symbol_name"`
	InferredType  string    `json:"inferred_type"`
	Confidence    float64   `json:"confidence"` // 0.0 - 1.0
	Reasoning     string    `json:"reasoning"`
	File          *File     `json:"file"`
	Line          int       `json:"line"`
}

// VariableUsage represents how a variable is used
type VariableUsage struct {
	Variable      *Symbol         `json:"variable"`
	File          *File           `json:"file"`
	UsageType     string          `json:"usage_type"` // "read", "write", "call"
	Line          int             `json:"line"`
	Context       string          `json:"context"`
	InferredType  string          `json:"inferred_type,omitempty"`
}

// CallSiteInfo represents information about a function/method call site
type CallSiteInfo struct {
	CallerSymbol  *Symbol         `json:"caller_symbol"`
	CalleeSymbol  *Symbol         `json:"callee_symbol,omitempty"` // nil if undefined
	File          *File           `json:"file"`
	Line          int             `json:"line"`
	Column        int             `json:"column"`
	Arguments     []*ArgumentInfo `json:"arguments"`
	IsValid       bool            `json:"is_valid"`
	Errors        []string        `json:"errors,omitempty"`
}

// ArgumentInfo represents an argument passed to a function/method
type ArgumentInfo struct {
	Position      int       `json:"position"`
	Value         string    `json:"value"`
	InferredType  string    `json:"inferred_type"`
	ExpectedType  string    `json:"expected_type,omitempty"`
	IsValid       bool      `json:"is_valid"`
}

// TypeSafetyScore represents the type safety score for code
type TypeSafetyScore struct {
	Score           float64   `json:"score"` // 0.0 - 100.0
	Rating          string    `json:"rating"` // "excellent", "good", "fair", "poor"
	TotalSymbols    int       `json:"total_symbols"`
	TypedSymbols    int       `json:"typed_symbols"`
	UntypedSymbols  int       `json:"untyped_symbols"`
	ErrorCount      int       `json:"error_count"`
	WarningCount    int       `json:"warning_count"`
	Recommendation  string    `json:"recommendation"`
}
