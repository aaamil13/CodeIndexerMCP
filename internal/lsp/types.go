package lsp

import "encoding/json"

// Message represents an LSP message
type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

// ResponseError represents an LSP error response
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      *int           `json:"id"`
	Error   *ResponseError `json:"error"`
}

// InitializeParams represents initialize request parameters
type InitializeParams struct {
	ProcessID             *int                `json:"processId"`
	RootPath              string              `json:"rootPath,omitempty"`
	RootURI               string              `json:"rootUri"`
	InitializationOptions interface{}         `json:"initializationOptions,omitempty"`
	Capabilities          ClientCapabilities  `json:"capabilities"`
	Trace                 string              `json:"trace,omitempty"`
	WorkspaceFolders      []WorkspaceFolder   `json:"workspaceFolders,omitempty"`
}

// ClientCapabilities represents client capabilities
type ClientCapabilities struct {
	Workspace    WorkspaceClientCapabilities    `json:"workspace,omitempty"`
	TextDocument TextDocumentClientCapabilities `json:"textDocument,omitempty"`
}

// WorkspaceClientCapabilities represents workspace-specific client capabilities
type WorkspaceClientCapabilities struct {
	ApplyEdit              bool                    `json:"applyEdit,omitempty"`
	WorkspaceEdit          WorkspaceEditCapabilities `json:"workspaceEdit,omitempty"`
	DidChangeConfiguration bool                    `json:"didChangeConfiguration,omitempty"`
	DidChangeWatchedFiles  bool                    `json:"didChangeWatchedFiles,omitempty"`
	Symbol                 bool                    `json:"symbol,omitempty"`
	ExecuteCommand         bool                    `json:"executeCommand,omitempty"`
}

// WorkspaceEditCapabilities represents workspace edit capabilities
type WorkspaceEditCapabilities struct {
	DocumentChanges bool `json:"documentChanges,omitempty"`
}

// TextDocumentClientCapabilities represents text document-specific client capabilities
type TextDocumentClientCapabilities struct {
	Synchronization    TextDocumentSyncClientCapabilities `json:"synchronization,omitempty"`
	Completion         CompletionClientCapabilities       `json:"completion,omitempty"`
	Hover              bool                               `json:"hover,omitempty"`
	SignatureHelp      bool                               `json:"signatureHelp,omitempty"`
	References         bool                               `json:"references,omitempty"`
	DocumentHighlight  bool                               `json:"documentHighlight,omitempty"`
	DocumentSymbol     bool                               `json:"documentSymbol,omitempty"`
	Formatting         bool                               `json:"formatting,omitempty"`
	RangeFormatting    bool                               `json:"rangeFormatting,omitempty"`
	OnTypeFormatting   bool                               `json:"onTypeFormatting,omitempty"`
	Definition         bool                               `json:"definition,omitempty"`
	CodeAction         bool                               `json:"codeAction,omitempty"`
	CodeLens           bool                               `json:"codeLens,omitempty"`
	DocumentLink       bool                               `json:"documentLink,omitempty"`
	Rename             bool                               `json:"rename,omitempty"`
}

// TextDocumentSyncClientCapabilities represents text document sync capabilities
type TextDocumentSyncClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	WillSave            bool `json:"willSave,omitempty"`
	WillSaveWaitUntil   bool `json:"willSaveWaitUntil,omitempty"`
	DidSave             bool `json:"didSave,omitempty"`
}

// CompletionClientCapabilities represents completion capabilities
type CompletionClientCapabilities struct {
	DynamicRegistration bool                     `json:"dynamicRegistration,omitempty"`
	CompletionItem      CompletionItemCapabilities `json:"completionItem,omitempty"`
}

// CompletionItemCapabilities represents completion item capabilities
type CompletionItemCapabilities struct {
	SnippetSupport bool `json:"snippetSupport,omitempty"`
}

// WorkspaceFolder represents a workspace folder
type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// InitializeResult represents initialize response
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo,omitempty"`
}

// ServerInfo represents server information
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// ServerCapabilities represents server capabilities
type ServerCapabilities struct {
	TextDocumentSync           TextDocumentSyncKind `json:"textDocumentSync"`
	CompletionProvider         bool                 `json:"completionProvider,omitempty"`
	HoverProvider              bool                 `json:"hoverProvider,omitempty"`
	SignatureHelpProvider      bool                 `json:"signatureHelpProvider,omitempty"`
	DefinitionProvider         bool                 `json:"definitionProvider,omitempty"`
	TypeDefinitionProvider     bool                 `json:"typeDefinitionProvider,omitempty"`
	ImplementationProvider     bool                 `json:"implementationProvider,omitempty"`
	ReferencesProvider         bool                 `json:"referencesProvider,omitempty"`
	DocumentHighlightProvider  bool                 `json:"documentHighlightProvider,omitempty"`
	DocumentSymbolProvider     bool                 `json:"documentSymbolProvider,omitempty"`
	WorkspaceSymbolProvider    bool                 `json:"workspaceSymbolProvider,omitempty"`
	CodeActionProvider         bool                 `json:"codeActionProvider,omitempty"`
	CodeLensProvider           bool                 `json:"codeLensProvider,omitempty"`
	DocumentFormattingProvider bool                 `json:"documentFormattingProvider,omitempty"`
	RenameProvider             bool                 `json:"renameProvider,omitempty"`
	DocumentLinkProvider       bool                 `json:"documentLinkProvider,omitempty"`
}

// TextDocumentSyncKind represents text document sync options
type TextDocumentSyncKind int

const (
	TextDocumentSyncKindNone        TextDocumentSyncKind = 0
	TextDocumentSyncKindFull        TextDocumentSyncKind = 1
	TextDocumentSyncKindIncremental TextDocumentSyncKind = 2
)

// Position represents a position in a text document
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a range in a text document
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location in a text document
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// TextDocumentIdentifier identifies a text document
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// VersionedTextDocumentIdentifier identifies a versioned text document
type VersionedTextDocumentIdentifier struct {
	TextDocumentIdentifier
	Version int `json:"version"`
}

// TextDocumentItem represents a text document
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// DidOpenTextDocumentParams represents didOpen parameters
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// DidChangeTextDocumentParams represents didChange parameters
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier   `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// TextDocumentContentChangeEvent represents a content change event
type TextDocumentContentChangeEvent struct {
	Range       *Range `json:"range,omitempty"`
	RangeLength int    `json:"rangeLength,omitempty"`
	Text        string `json:"text"`
}

// DidCloseTextDocumentParams represents didClose parameters
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DidSaveTextDocumentParams represents didSave parameters
type DidSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Text         *string                `json:"text,omitempty"`
}

// TextDocumentPositionParams represents position parameters
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// CompletionParams represents completion request parameters
type CompletionParams struct {
	TextDocumentPositionParams
	Context *CompletionContext `json:"context,omitempty"`
}

// CompletionContext represents completion context
type CompletionContext struct {
	TriggerKind      CompletionTriggerKind `json:"triggerKind"`
	TriggerCharacter string                `json:"triggerCharacter,omitempty"`
}

// CompletionTriggerKind represents how completion was triggered
type CompletionTriggerKind int

const (
	CompletionTriggerKindInvoked           CompletionTriggerKind = 1
	CompletionTriggerKindTriggerCharacter  CompletionTriggerKind = 2
	CompletionTriggerKindTriggerForIncompleteCompletions CompletionTriggerKind = 3
)

// CompletionList represents a list of completion items
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// CompletionItem represents a completion item
type CompletionItem struct {
	Label         string             `json:"label"`
	Kind          CompletionItemKind `json:"kind,omitempty"`
	Detail        string             `json:"detail,omitempty"`
	Documentation interface{}        `json:"documentation,omitempty"`
	SortText      string             `json:"sortText,omitempty"`
	FilterText    string             `json:"filterText,omitempty"`
	InsertText    string             `json:"insertText,omitempty"`
	TextEdit      *TextEdit          `json:"textEdit,omitempty"`
}

// CompletionItemKind represents the kind of completion item
type CompletionItemKind int

const (
	CompletionItemKindText          CompletionItemKind = 1
	CompletionItemKindMethod        CompletionItemKind = 2
	CompletionItemKindFunction      CompletionItemKind = 3
	CompletionItemKindConstructor   CompletionItemKind = 4
	CompletionItemKindField         CompletionItemKind = 5
	CompletionItemKindVariable      CompletionItemKind = 6
	CompletionItemKindClass         CompletionItemKind = 7
	CompletionItemKindInterface     CompletionItemKind = 8
	CompletionItemKindModule        CompletionItemKind = 9
	CompletionItemKindProperty      CompletionItemKind = 10
	CompletionItemKindUnit          CompletionItemKind = 11
	CompletionItemKindValue         CompletionItemKind = 12
	CompletionItemKindEnum          CompletionItemKind = 13
	CompletionItemKindKeyword       CompletionItemKind = 14
	CompletionItemKindSnippet       CompletionItemKind = 15
	CompletionItemKindColor         CompletionItemKind = 16
	CompletionItemKindFile          CompletionItemKind = 17
	CompletionItemKindReference     CompletionItemKind = 18
	CompletionItemKindFolder        CompletionItemKind = 19
	CompletionItemKindEnumMember    CompletionItemKind = 20
	CompletionItemKindConstant      CompletionItemKind = 21
	CompletionItemKindStruct        CompletionItemKind = 22
	CompletionItemKindEvent         CompletionItemKind = 23
	CompletionItemKindOperator      CompletionItemKind = 24
	CompletionItemKindTypeParameter CompletionItemKind = 25
)

// HoverParams represents hover request parameters
type HoverParams struct {
	TextDocumentPositionParams
}

// Hover represents hover information
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// MarkupContent represents formatted content
type MarkupContent struct {
	Kind  MarkupKind `json:"kind"`
	Value string     `json:"value"`
}

// MarkupKind represents the kind of markup
type MarkupKind string

const (
	MarkupKindPlainText MarkupKind = "plaintext"
	MarkupKindMarkdown  MarkupKind = "markdown"
)

// DefinitionParams represents definition request parameters
type DefinitionParams struct {
	TextDocumentPositionParams
}

// ReferenceParams represents references request parameters
type ReferenceParams struct {
	TextDocumentPositionParams
	Context ReferenceContext `json:"context"`
}

// ReferenceContext represents reference context
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// DocumentSymbolParams represents document symbol request parameters
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DocumentSymbol represents a document symbol
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           SymbolKind       `json:"kind"`
	Deprecated     bool             `json:"deprecated,omitempty"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// SymbolKind represents the kind of symbol
type SymbolKind int

const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)

// WorkspaceSymbolParams represents workspace symbol search parameters
type WorkspaceSymbolParams struct {
	Query string `json:"query"`
}

// SymbolInformation represents symbol information
type SymbolInformation struct {
	Name          string     `json:"name"`
	Kind          SymbolKind `json:"kind"`
	Location      Location   `json:"location"`
	ContainerName string     `json:"containerName,omitempty"`
	Deprecated    bool       `json:"deprecated,omitempty"`
}

// CodeActionParams represents code action request parameters
type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

// CodeActionContext represents code action context
type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Diagnostic represents a diagnostic (error, warning, etc.)
type Diagnostic struct {
	Range    Range              `json:"range"`
	Severity DiagnosticSeverity `json:"severity,omitempty"`
	Code     interface{}        `json:"code,omitempty"`
	Source   string             `json:"source,omitempty"`
	Message  string             `json:"message"`
}

// DiagnosticSeverity represents the severity of a diagnostic
type DiagnosticSeverity int

const (
	DiagnosticSeverityError       DiagnosticSeverity = 1
	DiagnosticSeverityWarning     DiagnosticSeverity = 2
	DiagnosticSeverityInformation DiagnosticSeverity = 3
	DiagnosticSeverityHint        DiagnosticSeverity = 4
)

// CodeAction represents a code action
type CodeAction struct {
	Title       string         `json:"title"`
	Kind        string         `json:"kind,omitempty"`
	Diagnostics []Diagnostic   `json:"diagnostics,omitempty"`
	Edit        *WorkspaceEdit `json:"edit,omitempty"`
	Command     *Command       `json:"command,omitempty"`
}

// Command represents a command
type Command struct {
	Title     string        `json:"title"`
	Command   string        `json:"command"`
	Arguments []interface{} `json:"arguments,omitempty"`
}

// RenameParams represents rename request parameters
type RenameParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	NewName      string                 `json:"newName"`
}

// WorkspaceEdit represents a workspace edit
type WorkspaceEdit struct {
	Changes         map[string][]TextEdit `json:"changes,omitempty"`
	DocumentChanges []TextDocumentEdit    `json:"documentChanges,omitempty"`
}

// TextEdit represents a text edit
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// TextDocumentEdit represents a text document edit
type TextDocumentEdit struct {
	TextDocument VersionedTextDocumentIdentifier `json:"textDocument"`
	Edits        []TextEdit                      `json:"edits"`
}

// SignatureHelpParams represents signature help request parameters
type SignatureHelpParams struct {
	TextDocumentPositionParams
	Context *SignatureHelpContext `json:"context,omitempty"`
}

// SignatureHelpContext represents signature help context
type SignatureHelpContext struct {
	TriggerKind         SignatureHelpTriggerKind `json:"triggerKind"`
	TriggerCharacter    string                   `json:"triggerCharacter,omitempty"`
	IsRetrigger         bool                     `json:"isRetrigger"`
	ActiveSignatureHelp *SignatureHelp           `json:"activeSignatureHelp,omitempty"`
}

// SignatureHelpTriggerKind represents how signature help was triggered
type SignatureHelpTriggerKind int

const (
	SignatureHelpTriggerKindInvoked          SignatureHelpTriggerKind = 1
	SignatureHelpTriggerKindTriggerCharacter SignatureHelpTriggerKind = 2
	SignatureHelpTriggerKindContentChange    SignatureHelpTriggerKind = 3
)

// SignatureHelp represents signature help information
type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature int                    `json:"activeSignature,omitempty"`
	ActiveParameter int                    `json:"activeParameter,omitempty"`
}

// SignatureInformation represents a function signature
type SignatureInformation struct {
	Label         string                 `json:"label"`
	Documentation interface{}            `json:"documentation,omitempty"`
	Parameters    []ParameterInformation `json:"parameters,omitempty"`
}

// ParameterInformation represents a function parameter
type ParameterInformation struct {
	Label         string      `json:"label"`
	Documentation interface{} `json:"documentation,omitempty"`
}
