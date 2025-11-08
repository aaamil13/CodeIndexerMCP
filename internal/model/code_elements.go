package model

import "time"

// Position представя позиция в кода
type Position struct {
    Line   int `json:"line"`
    Column int `json:"column"`
    Byte   int `json:"byte"`
}

// Range представя диапазон в кода
type Range struct {
    Start Position `json:"start"`
    End   Position `json:"end"`
}

// Status за AI-driven development
type DevelopmentStatus string

const (
    StatusPlanned    DevelopmentStatus = "planned"
    StatusInProgress DevelopmentStatus = "in_progress"
    StatusCompleted  DevelopmentStatus = "completed"
    StatusTesting    DevelopmentStatus = "testing"
    StatusVerified   DevelopmentStatus = "verified"
    StatusFailed     DevelopmentStatus = "failed"
)

// CodeElement е базовия интерфейс за всички елементи
type CodeElement interface {
    GetName() string
    GetKind() string
    GetRange() Range
    GetFile() string
}

// SymbolKind defines the type of a symbol
type SymbolKind string

const (
	SymbolKindFile        SymbolKind = "file"
	SymbolKindModule      SymbolKind = "module"
	SymbolKindNamespace   SymbolKind = "namespace"
	SymbolKindPackage     SymbolKind = "package"
	SymbolKindClass       SymbolKind = "class"
	SymbolKindMethod      SymbolKind = "method"
	SymbolKindFunction    SymbolKind = "function"
	SymbolKindConstructor SymbolKind = "constructor"
	SymbolKindVariable    SymbolKind = "variable"
	SymbolKindConstant    SymbolKind = "constant"
	SymbolKindField       SymbolKind = "field"
	SymbolKindProperty    SymbolKind = "property"
	SymbolKindEnum        SymbolKind = "enum"
	SymbolKindInterface   SymbolKind = "interface"
	SymbolKindStruct      SymbolKind = "struct"
	SymbolKindTypeAlias   SymbolKind = "type_alias"
	SymbolKindDecorator   SymbolKind = "decorator"
	SymbolKindUnknown     SymbolKind = "unknown"
)

// Visibility defines the visibility of a symbol
type Visibility string

const (
	VisibilityPublic    Visibility = "public"
	VisibilityPrivate   Visibility = "private"
	VisibilityProtected Visibility = "protected"
	VisibilityInternal  Visibility = "internal"
	VisibilityPackage   Visibility = "package"
	VisibilityUnknown   Visibility = "unknown"
)

// ImportKind defines the type of an import
type ImportKind string

const (
	ImportKindStdlib   ImportKind = "stdlib"
	ImportKindLocal    ImportKind = "local"
	ImportKindExternal ImportKind = "external"
	ImportKindUnknown  ImportKind = "unknown"
)

// Symbol представя универсален символ
type Symbol struct {
    ID            string            `json:"id"`
    Name          string            `json:"name"`
    Kind          SymbolKind        `json:"kind"` // "function", "class", "method", etc.
    File          string            `json:"file"`
    Range         Range             `json:"range"`
    Signature     string            `json:"signature"`
    Documentation string            `json:"documentation"`
    Visibility    Visibility        `json:"visibility"` // "public", "private", "protected"
    Language      string            `json:"language"`
    
    // 💡 ПОДОБРЕНИЕ #5: Content Hash за детекция на промени
    ContentHash   string            `json:"content_hash"`
    
    // AI-driven development metadata
    Status        DevelopmentStatus `json:"status,omitempty"`
    Priority      int               `json:"priority,omitempty"`
    AssignedAgent string            `json:"assigned_agent,omitempty"`
    TestIDs       []string          `json:"test_ids,omitempty"`
    Dependencies  []string          `json:"dependencies,omitempty"`
    
    CreatedAt     time.Time         `json:"created_at"`
    UpdatedAt     time.Time         `json:"updated_at"`
    
    Metadata      map[string]string `json:"metadata,omitempty"`
}

// Function представя функция
type Function struct {
    Symbol
    Parameters   []Parameter `json:"parameters"`
    ReturnType   string      `json:"return_type,omitempty"`
    Body         string      `json:"body,omitempty"`
    IsAsync      bool        `json:"is_async,omitempty"`
    IsGenerator  bool        `json:"is_generator,omitempty"`
    Decorators   []string    `json:"decorators,omitempty"`
}

// Parameter представя параметър на функция
type Parameter struct {
    Name         string `json:"name"`
    Type         string `json:"type,omitempty"`
    DefaultValue string `json:"default_value,omitempty"`
    IsOptional   bool   `json:"is_optional"`
    IsVariadic   bool   `json:"is_variadic"`
}

// Class представя клас
type Class struct {
    Symbol
    BaseClasses  []string   `json:"base_classes,omitempty"`
    Interfaces   []string   `json:"interfaces,omitempty"`
    Methods      []Method   `json:"methods"`
    Fields       []Field    `json:"fields"`
    IsAbstract   bool       `json:"is_abstract,omitempty"`
    IsInterface  bool       `json:"is_interface,omitempty"`
}

// Method представя метод на клас
type Method struct {
    Function
    ReceiverType string `json:"receiver_type,omitempty"`
    IsStatic     bool   `json:"is_static,omitempty"`
    IsVirtual    bool   `json:"is_virtual,omitempty"`
    IsOverride   bool   `json:"is_override,omitempty"`
}

// Field представя поле на клас
type Field struct {
    Name         string `json:"name"`
    Type         string `json:"type"`
    DefaultValue string `json:"default_value,omitempty"`
    Visibility   string `json:"visibility"`
    IsStatic     bool   `json:"is_static"`
    IsConstant   bool   `json:"is_const"`
}

// Import представя импорт
type Import struct {
    FilePath   string   `json:"file_path"` // Path of the file where this import was found
    Path       string   `json:"path"`
    Alias      string   `json:"alias,omitempty"`
    Members    []string `json:"members,omitempty"`
    IsWildcard bool     `json:"is_wildcard"`
    Range      Range    `json:"range"`
}

// Variable представя променлива
type Variable struct {
    Symbol
    Type          string `json:"type,omitempty"`
    InitialValue  string `json:"initial_value,omitempty"`
    IsConstant    bool   `json:"is_constant"`
    Scope         string `json:"scope"` // "global", "local", "module"
}

// Interface представя интерфейс
type Interface struct {
    Symbol
    Methods      []Method `json:"methods"`
    BaseTypes    []string `json:"base_types,omitempty"`
}

// TestDefinition за AI test generation
type TestDefinition struct {
    ID               string   `json:"id"`
    TargetSymbolID   string   `json:"target_symbol_id"`
    TestName         string   `json:"test_name"`
    Description      string   `json:"description"`
    ExpectedBehavior string   `json:"expected_behavior"`
    Preconditions    []string `json:"preconditions"`
    Assertions       []string `json:"assertions"`
    Status           DevelopmentStatus `json:"status"`
    Priority         int      `json:"priority"`
}

// BuildTask за AI-driven scaffold
type BuildTask struct {
    ID            string            `json:"id"`
    Type          string            `json:"type"` // "create_function", "implement_method", etc.
    TargetSymbol  string            `json:"target_symbol"`
    Description   string            `json:"description"`
    Status        DevelopmentStatus `json:"status"`
    Priority      int               `json:"priority"`
    Dependencies  []string          `json:"dependencies"` // IDs на други tasks
    AssignedAgent string            `json:"assigned_agent,omitempty"`
    CreatedAt     time.Time         `json:"created_at"`
    UpdatedAt     time.Time         `json:"updated_at"`
    CompletedAt   *time.Time        `json:"completed_at,omitempty"`
}

// Имплементации на интерфейса
func (s *Symbol) GetName() string   { return s.Name }
func (s *Symbol) GetKind() string   { return string(s.Kind) }
func (s *Symbol) GetRange() Range   { return s.Range }
func (s *Symbol) GetFile() string   { return s.File }

func (f *Function) GetName() string { return f.Symbol.Name }
func (f *Function) GetKind() string { return "function" }
func (f *Function) GetRange() Range { return f.Symbol.Range }
func (f *Function) GetFile() string { return f.Symbol.File }

func (c *Class) GetName() string { return c.Symbol.Name }
func (c *Class) GetKind() string { return "class" }
func (c *Class) GetRange() Range { return c.Symbol.Range }
func (c *Class) GetFile() string { return c.Symbol.File }

// Reference represents a reference between symbols
type Reference struct {
	SourceSymbolID   string `json:"source_symbol_id"`
	TargetSymbolName string `json:"target_symbol_name"`
	ReferenceType    string `json:"reference_type"`
	FilePath         string `json:"file_path"`
	Line             int    `json:"line"`
	Column           int    `json:"column"`
}

// RelationshipType defines the type of relationship
type RelationshipType string

const (
	RelationshipKindCalls      RelationshipType = "calls"
	RelationshipKindUses       RelationshipType = "uses"
	RelationshipKindExtends    RelationshipType = "extends"
	RelationshipKindImplements RelationshipType = "implements"
	RelationshipKindComposes   RelationshipType = "composes"
)

// Relationship represents a relationship between two symbols
type Relationship struct {
	Type         RelationshipType `json:"type"`
	SourceSymbol string           `json:"source_symbol"` // Name or ID of the source symbol
	TargetSymbol string           `json:"target_symbol"` // Name or ID of the target symbol
	FilePath     string           `json:"file_path"`     // File where the relationship was found
	Line         int              `json:"line"`          // Line number where the relationship was found
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// FrameworkInfo holds information about detected frameworks
type FrameworkInfo struct {
	Name    string            `json:"name"`
	Version string            `json:"version,omitempty"`
	Config  map[string]string `json:"config,omitempty"`
	Entries []string          `json:"entry_points,omitempty"` // e.g., for Django, settings.py
}


