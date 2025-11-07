package types

// SymbolType represents the type of code symbol
type SymbolType string

const (
	SymbolTypeFunction  SymbolType = "function"
	SymbolTypeClass     SymbolType = "class"
	SymbolTypeMethod    SymbolType = "method"
	SymbolTypeVariable  SymbolType = "variable"
	SymbolTypeInterface SymbolType = "interface"
	SymbolTypeType      SymbolType = "type"
	SymbolTypeEnum      SymbolType = "enum"
	SymbolTypeStruct    SymbolType = "struct"
	SymbolTypeConstant  SymbolType = "constant"
	SymbolTypePackage   SymbolType = "package"
	SymbolTypeModule      SymbolType = "module"
	SymbolTypeNamespace   SymbolType = "namespace"   // For C#, PHP, C++ namespaces
	SymbolTypeProperty    SymbolType = "property"    // For properties in languages like C#, Kotlin, Swift
	SymbolTypeField       SymbolType = "field"       // For fields/member variables in classes/structs
	SymbolTypeConstructor SymbolType = "constructor" // For class constructors
	SymbolTypeDecorator   SymbolType = "decorator"   // For Python decorators
)

// Visibility represents symbol visibility
type Visibility string

const (
	VisibilityPublic    Visibility = "public"
	VisibilityPrivate   Visibility = "private"
	VisibilityProtected Visibility = "protected"
	VisibilityInternal  Visibility = "internal"
	VisibilityPackage   Visibility = "package" // For Java package-private
)

// Symbol represents a code symbol (function, class, variable, etc.)
type Symbol struct {
	ID            int64                  `json:"id"`
	FileID        int64                  `json:"file_id"`
	Name          string                 `json:"name"`
	Type          SymbolType             `json:"type"`
	Signature     string                 `json:"signature"`
	ParentID      *int64                 `json:"parent_id,omitempty"`
	StartLine     int                    `json:"start_line"`
	EndLine       int                    `json:"end_line"`
	StartColumn   int                    `json:"start_column"`
	EndColumn     int                    `json:"end_column"`
	Visibility    Visibility             `json:"visibility"`
	IsExported    bool                   `json:"is_exported"`
	IsAsync       bool                   `json:"is_async"`
	IsStatic      bool                   `json:"is_static"`
	IsAbstract    bool                   `json:"is_abstract"`
	Documentation string                 `json:"documentation,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// RelationshipType represents the type of relationship between symbols
type RelationshipType string

const (
	RelationshipExtends    RelationshipType = "extends"
	RelationshipImplements RelationshipType = "implements"
	RelationshipCalls      RelationshipType = "calls"
	RelationshipUses       RelationshipType = "uses"
	RelationshipImports    RelationshipType = "imports"
	RelationshipContains   RelationshipType = "contains"
)

// Relationship represents a relationship between two symbols (using names, resolved to IDs later)
type Relationship struct {
	ID           int64            `json:"id"`
	FromSymbolID int64            `json:"from_symbol_id,omitempty"` // Resolved ID from SourceName
	ToSymbolID   int64            `json:"to_symbol_id,omitempty"`   // Resolved ID from TargetName
	SourceName   string           `json:"source_name"`              // Original name of the source symbol
	TargetName   string           `json:"target_name"`              // Original name of the target symbol
	Type         RelationshipType `json:"type"`
}

// Reference represents a reference to a symbol
type Reference struct {
	ID            int64  `json:"id"`
	SymbolID      int64  `json:"symbol_id"`
	FileID        int64  `json:"file_id"`
	LineNumber    int    `json:"line_number"`
	ColumnNumber  int    `json:"column_number"`
	ReferenceType string `json:"reference_type"` // 'call', 'assignment', 'type_reference'
}
