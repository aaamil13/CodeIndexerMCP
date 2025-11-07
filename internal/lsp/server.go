package lsp

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/aaamil13/CodeIndexerMCP/internal/ai"
	"github.com/aaamil13/CodeIndexerMCP/internal/core"
	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// Server implements a Language Server Protocol server
type Server struct {
	db           *database.DB
	indexer      *core.Indexer
	analyzer     *ai.SemanticAnalyzer
	capabilities ServerCapabilities
	workspaces   map[string]*Workspace
	mu           sync.RWMutex

	// Communication channels
	reader io.Reader
	writer io.Writer

	// State
	initialized bool
	shutdown    bool
}

// NewServer creates a new LSP server
func NewServer(db *database.DB, indexer *core.Indexer) *Server {
	return &Server{
		db:         db,
		indexer:    indexer,
		analyzer:   ai.NewSemanticAnalyzer(db),
		workspaces: make(map[string]*Workspace),
		capabilities: ServerCapabilities{
			TextDocumentSync:   TextDocumentSyncKindFull,
			CompletionProvider: true,
			HoverProvider:      true,
			DefinitionProvider: true,
			ReferencesProvider: true,
			DocumentSymbolProvider: true,
			WorkspaceSymbolProvider: true,
			CodeActionProvider: true,
			RenameProvider:     true,
			SignatureHelpProvider: true,
		},
	}
}

// Start starts the LSP server
func (s *Server) Start(reader io.Reader, writer io.Writer) error {
	s.reader = reader
	s.writer = writer

	// Start message handling loop
	return s.handleMessages()
}

// handleMessages handles incoming LSP messages
func (s *Server) handleMessages() error {
	decoder := json.NewDecoder(s.reader)
	encoder := json.NewEncoder(s.writer)

	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("decode message: %w", err)
		}

		// Handle message
		response, err := s.handleMessage(&msg)
		if err != nil {
			// Send error response
			errResp := ErrorResponse{
				ID:    msg.ID,
				Error: &ResponseError{
					Code:    InternalError,
					Message: err.Error(),
				},
			}
			if err := encoder.Encode(errResp); err != nil {
				return fmt.Errorf("encode error response: %w", err)
			}
			continue
		}

		// Send response if not a notification
		if msg.ID != nil && response != nil {
			if err := encoder.Encode(response); err != nil {
				return fmt.Errorf("encode response: %w", err)
			}
		}
	}
}

// handleMessage handles a single LSP message
func (s *Server) handleMessage(msg *Message) (interface{}, error) {
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "initialized":
		return s.handleInitialized(msg)
	case "shutdown":
		return s.handleShutdown(msg)
	case "exit":
		return s.handleExit(msg)

	// Text document synchronization
	case "textDocument/didOpen":
		return s.handleTextDocumentDidOpen(msg)
	case "textDocument/didChange":
		return s.handleTextDocumentDidChange(msg)
	case "textDocument/didClose":
		return s.handleTextDocumentDidClose(msg)
	case "textDocument/didSave":
		return s.handleTextDocumentDidSave(msg)

	// Language features
	case "textDocument/completion":
		return s.handleCompletion(msg)
	case "textDocument/hover":
		return s.handleHover(msg)
	case "textDocument/definition":
		return s.handleDefinition(msg)
	case "textDocument/references":
		return s.handleReferences(msg)
	case "textDocument/documentSymbol":
		return s.handleDocumentSymbol(msg)
	case "workspace/symbol":
		return s.handleWorkspaceSymbol(msg)
	case "textDocument/codeAction":
		return s.handleCodeAction(msg)
	case "textDocument/rename":
		return s.handleRename(msg)
	case "textDocument/signatureHelp":
		return s.handleSignatureHelp(msg)

	// Custom AI-powered features
	case "codeindexer/analyze":
		return s.handleAnalyze(msg)
	case "codeindexer/typeCheck":
		return s.handleTypeCheck(msg)
	case "codeindexer/findUnused":
		return s.handleFindUnused(msg)
	case "codeindexer/callGraph":
		return s.handleCallGraph(msg)
	case "codeindexer/dependencies":
		return s.handleDependencies(msg)

	default:
		return nil, fmt.Errorf("unknown method: %s", msg.Method)
	}
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(msg *Message) (interface{}, error) {
	var params InitializeParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, fmt.Errorf("unmarshal initialize params: %w", err)
	}

	// Store workspace information
	if params.RootURI != "" {
		ws := &Workspace{
			URI:  params.RootURI,
			Name: params.RootURI,
		}
		s.mu.Lock()
		s.workspaces[params.RootURI] = ws
		s.mu.Unlock()

		// Index workspace
		go s.indexWorkspace(params.RootURI)
	}

	return InitializeResult{
		Capabilities: s.capabilities,
		ServerInfo: ServerInfo{
			Name:    "CodeIndexer LSP",
			Version: "1.0.0",
		},
	}, nil
}

// handleInitialized handles the initialized notification
func (s *Server) handleInitialized( msg *Message) (interface{}, error) {
	s.initialized = true
	return nil, nil
}

// handleShutdown handles the shutdown request
func (s *Server) handleShutdown(msg *Message) (interface{}, error) {
	s.shutdown = true
	return nil, nil
}

// handleExit handles the exit notification
func (s *Server) handleExit(msg *Message) (interface{}, error) {
	// Exit the server
	return nil, io.EOF
}

// handleTextDocumentDidOpen handles document open notification
func (s *Server) handleTextDocumentDidOpen(msg *Message) (interface{}, error) {
	var params DidOpenTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	// Index the opened document
	go s.indexDocument(params.TextDocument.URI, []byte(params.TextDocument.Text))

	return nil, nil
}

// handleTextDocumentDidChange handles document change notification
func (s *Server) handleTextDocumentDidChange(msg *Message) (interface{}, error) {
	var params DidChangeTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	// Re-index the changed document
	if len(params.ContentChanges) > 0 {
		// For full sync, just use the last change
		lastChange := params.ContentChanges[len(params.ContentChanges)-1]
		go s.indexDocument(params.TextDocument.URI, []byte(lastChange.Text))
	}

	return nil, nil
}

// handleTextDocumentDidClose handles document close notification
func (s *Server) handleTextDocumentDidClose(msg *Message) (interface{}, error) {
	// Nothing to do for now
	return nil, nil
}

// handleTextDocumentDidSave handles document save notification
func (s *Server) handleTextDocumentDidSave(msg *Message) (interface{}, error) {
	var params DidSaveTextDocumentParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	// Re-index on save
	if params.Text != nil {
		go s.indexDocument(params.TextDocument.URI, []byte(*params.Text))
	}

	return nil, nil
}

// handleCompletion handles completion requests
func (s *Server) handleCompletion(msg *Message) (interface{}, error) {
	var params CompletionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	// Get symbols in scope
	symbols, err := s.getSymbolsInScope(params.TextDocument.URI, params.Position)
	if err != nil {
		return nil, err
	}

	// Convert to completion items
	items := make([]CompletionItem, 0, len(symbols))
	for _, symbol := range symbols {
		items = append(items, CompletionItem{
			Label:  symbol.Name,
			Kind:   symbolTypeToCompletionKind(symbol.Type),
			Detail: symbol.Signature,
			Documentation: symbol.Documentation,
		})
	}

	return CompletionList{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

// handleHover handles hover requests
func (s *Server) handleHover(msg *Message) (interface{}, error) {
	var params HoverParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	// Find symbol at position
	symbol, err := s.getSymbolAtPosition(params.TextDocument.URI, params.Position)
	if err != nil || symbol == nil {
		return nil, nil
	}

	// Build hover content
	content := fmt.Sprintf("```%s\n%s\n```", "go", symbol.Signature)
	if symbol.Documentation != "" {
		content += "\n\n" + symbol.Documentation
	}

	return Hover{
		Contents: MarkupContent{
			Kind:  MarkupKindMarkdown,
			Value: content,
		},
	}, nil
}

// handleDefinition handles go-to-definition requests
func (s *Server) handleDefinition(msg *Message) (interface{}, error) {
	var params DefinitionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	// Find symbol at position
	symbol, err := s.getSymbolAtPosition(params.TextDocument.URI, params.Position)
	if err != nil || symbol == nil {
		return nil, nil
	}

	// Get file information
	file, err := s.db.GetFile(symbol.FileID)
	if err != nil {
		return nil, err
	}

	return Location{
		URI: "file://" + file.Path,
		Range: Range{
			Start: Position{Line: symbol.StartLine - 1, Character: symbol.StartColumn},
			End:   Position{Line: symbol.EndLine - 1, Character: symbol.EndColumn},
		},
	}, nil
}

// handleReferences handles find-references requests
func (s *Server) handleReferences(msg *Message) (interface{}, error) {
	var params ReferenceParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	// Find symbol at position
	symbol, err := s.getSymbolAtPosition(params.TextDocument.URI, params.Position)
	if err != nil || symbol == nil {
		return []Location{}, nil
	}

	// Get all references
	refs, err := s.db.GetReferencesBySymbol(symbol.ID)
	if err != nil {
		return nil, err
	}

	// Convert to LSP locations
	locations := make([]Location, 0, len(refs))
	for _, ref := range refs {
		file, err := s.db.GetFile(ref.FileID)
		if err != nil {
			continue
		}

		locations = append(locations, Location{
			URI: "file://" + file.Path,
			Range: Range{
				Start: Position{Line: ref.LineNumber - 1, Character: ref.ColumnNumber},
				End:   Position{Line: ref.LineNumber - 1, Character: ref.ColumnNumber + len(symbol.Name)},
			},
		})
	}

	return locations, nil
}

// handleDocumentSymbol handles document symbol requests
func (s *Server) handleDocumentSymbol(msg *Message) (interface{}, error) {
	var params DocumentSymbolParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	// Get file ID from URI
	fileID, err := s.getFileIDFromURI(params.TextDocument.URI)
	if err != nil {
		return []DocumentSymbol{}, nil
	}

	// Get symbols for file
	symbols, err := s.db.GetSymbolsByFile(fileID)
	if err != nil {
		return nil, err
	}

	// Convert to LSP document symbols
	docSymbols := make([]DocumentSymbol, 0, len(symbols))
	for _, symbol := range symbols {
		docSymbols = append(docSymbols, DocumentSymbol{
			Name:   symbol.Name,
			Detail: symbol.Signature,
			Kind:   symbolTypeToSymbolKind(symbol.Type),
			Range: Range{
				Start: Position{Line: symbol.StartLine - 1, Character: symbol.StartColumn},
				End:   Position{Line: symbol.EndLine - 1, Character: symbol.EndColumn},
			},
			SelectionRange: Range{
				Start: Position{Line: symbol.StartLine - 1, Character: symbol.StartColumn},
				End:   Position{Line: symbol.StartLine - 1, Character: symbol.StartColumn + len(symbol.Name)},
			},
		})
	}

	return docSymbols, nil
}

// handleWorkspaceSymbol handles workspace symbol search
func (s *Server) handleWorkspaceSymbol(msg *Message) (interface{}, error) {
	var params WorkspaceSymbolParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	// Search symbols
	symbols, err := s.db.SearchSymbols(types.SearchOptions{
		Query: params.Query,
		Limit: 50,
	})
	if err != nil {
		return nil, err
	}

	// Convert to LSP symbol information
	symbolInfo := make([]SymbolInformation, 0, len(symbols))
	for _, symbol := range symbols {
		file, err := s.db.GetFile(symbol.FileID)
		if err != nil {
			continue
		}

		symbolInfo = append(symbolInfo, SymbolInformation{
			Name: symbol.Name,
			Kind: symbolTypeToSymbolKind(symbol.Type),
			Location: Location{
				URI: "file://" + file.Path,
				Range: Range{
					Start: Position{Line: symbol.StartLine - 1, Character: symbol.StartColumn},
					End:   Position{Line: symbol.EndLine - 1, Character: symbol.EndColumn},
				},
			},
		})
	}

	return symbolInfo, nil
}

// handleCodeAction handles code action requests
func (s *Server) handleCodeAction(msg *Message) (interface{}, error) {
	var params CodeActionParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	// For now, return empty code actions
	// In the future, this could provide:
	// - Quick fixes for type errors
	// - Refactoring suggestions
	// - Import organization
	return []CodeAction{}, nil
}

// handleRename handles rename requests
func (s *Server) handleRename(msg *Message) (interface{}, error) {
	var params RenameParams
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	// Find symbol at position
	symbol, err := s.getSymbolAtPosition(params.TextDocument.URI, params.Position)
	if err != nil || symbol == nil {
		return nil, fmt.Errorf("symbol not found")
	}

	// Get all references
	refs, err := s.db.GetReferencesBySymbol(symbol.ID)
	if err != nil {
		return nil, err
	}

	// Build workspace edit
	changes := make(map[string][]TextEdit)

	// Add edit for definition
	file, _ := s.db.GetFile(symbol.FileID)
	if file != nil {
		uri := "file://" + file.Path
		changes[uri] = append(changes[uri], TextEdit{
			Range: Range{
				Start: Position{Line: symbol.StartLine - 1, Character: symbol.StartColumn},
				End:   Position{Line: symbol.StartLine - 1, Character: symbol.StartColumn + len(symbol.Name)},
			},
			NewText: params.NewName,
		})
	}

	// Add edits for references
	for _, ref := range refs {
		file, err := s.db.GetFile(ref.FileID)
		if err != nil {
			continue
		}

		uri := "file://" + file.Path
		changes[uri] = append(changes[uri], TextEdit{
			Range: Range{
				Start: Position{Line: ref.LineNumber - 1, Character: ref.ColumnNumber},
				End:   Position{Line: ref.LineNumber - 1, Character: ref.ColumnNumber + len(symbol.Name)},
			},
			NewText: params.NewName,
		})
	}

	return WorkspaceEdit{
		Changes: changes,
	}, nil
}

// handleSignatureHelp handles signature help requests
func (s *Server) handleSignatureHelp(msg *Message) (interface{}, error) {
	// Not implemented yet
	return SignatureHelp{
		Signatures: []SignatureInformation{},
	}, nil
}

// Custom AI-powered handlers

func (s *Server) handleAnalyze(msg *Message) (interface{}, error) {
	var params struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	fileID, err := s.getFileIDFromURI(params.URI)
	if err != nil {
		return nil, err
	}

	file, err := s.db.GetFile(fileID)
	if err != nil {
		return nil, err
	}

	// Run semantic analysis on project
	result, err := s.analyzer.AnalyzeProject(file.ProjectID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Server) handleTypeCheck(msg *Message) (interface{}, error) {
	var params struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	fileID, err := s.getFileIDFromURI(params.URI)
	if err != nil {
		return nil, err
	}

	// Use type validator
	validator := ai.NewTypeValidator(s.db)
	validation, err := validator.ValidateFile(fileID)
	if err != nil {
		return nil, err
	}

	return validation, nil
}

func (s *Server) handleFindUnused(msg *Message) (interface{}, error) {
	// Find unused symbols in project
	return nil, fmt.Errorf("not implemented")
}

func (s *Server) handleCallGraph(msg *Message) (interface{}, error) {
	var params struct {
		ProjectID int64 `json:"projectId"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	callGraph, err := s.analyzer.AnalyzeCallGraph(params.ProjectID)
	if err != nil {
		return nil, err
	}

	return callGraph, nil
}

func (s *Server) handleDependencies(msg *Message) (interface{}, error) {
	var params struct {
		ProjectID int64 `json:"projectId"`
	}
	if err := json.Unmarshal(msg.Params, &params); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("not implemented: project-wide dependency graph")
}

// Helper methods

func (s *Server) indexWorkspace(uri string) {
	// Extract path from URI
	path := uriToPath(uri)

	// Get or create project
	project, err := s.db.GetProject(path)
	if err != nil {
		return
	}
	if project == nil {
		project = &types.Project{
			Path: path,
			Name: path, // Use path as name for now
		}
		if err := s.db.CreateProject(project); err != nil {
			return
		}
	}

	s.indexer.IndexAll()
}

func (s *Server) indexDocument(uri string, content []byte) {
	// Index single document
	path := uriToPath(uri)

	// This is simplified - in production would need to get/create project
	s.indexer.IndexFile(path)
}

func (s *Server) getSymbolAtPosition(uri string, pos Position) (*types.Symbol, error) {
	fileID, err := s.getFileIDFromURI(uri)
	if err != nil {
		return nil, err
	}

	symbols, err := s.db.GetSymbolsByFile(fileID)
	if err != nil {
		return nil, err
	}

	// Find symbol at position (line, column)
	for _, symbol := range symbols {
		if pos.Line+1 >= symbol.StartLine && pos.Line+1 <= symbol.EndLine {
			return symbol, nil
		}
	}

	return nil, nil
}

func (s *Server) getSymbolsInScope(uri string, pos Position) ([]*types.Symbol, error) {
	fileID, err := s.getFileIDFromURI(uri)
	if err != nil {
		return nil, err
	}

	// Get all symbols in file (simplified - would need proper scope analysis)
	return s.db.GetSymbolsByFile(fileID)
}

func (s *Server) getFileIDFromURI(uri string) (int64, error) {
	path := uriToPath(uri)
	projectID, err := s.getProjectIDFromURI(uri)
	if err != nil {
		return 0, err
	}

	file, err := s.db.GetFileByPath(projectID, path)
	if err != nil {
		return 0, err
	}
	if file == nil {
		return 0, fmt.Errorf("file not found: %s", path)
	}
	return file.ID, nil
}

func (s *Server) getProjectIDFromURI(uri string) (int64, error) {
	path := uriToPath(uri)
	project, err := s.db.GetProject(path)
	if err != nil {
		return 0, err
	}
	if project == nil {
		return 0, fmt.Errorf("project not found for uri: %s", uri)
	}
	return project.ID, nil
}

func uriToPath(uri string) string {
	// Simple conversion - would need proper URI parsing
	if len(uri) > 7 && uri[:7] == "file://" {
		return uri[7:]
	}
	return uri
}

// Workspace represents an LSP workspace
type Workspace struct {
	URI  string
	Name string
}
