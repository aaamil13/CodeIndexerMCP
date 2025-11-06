package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aaamil13/CodeIndexerMCP/internal/core"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// Server is the MCP server
type Server struct {
	indexer *core.Indexer
	tools   map[string]*Tool
	stdin   io.Reader
	stdout  io.Writer
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Handler     func(params json.RawMessage) (interface{}, error)
}

// NewServer creates a new MCP server
func NewServer(indexer *core.Indexer) *Server {
	server := &Server{
		indexer: indexer,
		tools:   make(map[string]*Tool),
		stdin:   os.Stdin,
		stdout:  os.Stdout,
	}

	server.registerTools()
	return server
}

// Start starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	decoder := json.NewDecoder(s.stdin)
	encoder := json.NewEncoder(s.stdout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var request MCPRequest
			if err := decoder.Decode(&request); err != nil {
				if err == io.EOF {
					return nil
				}
				s.sendError(encoder, "", fmt.Errorf("failed to decode request: %w", err))
				continue
			}

			response := s.handleRequest(&request)
			if err := encoder.Encode(response); err != nil {
				return fmt.Errorf("failed to encode response: %w", err)
			}
		}
	}
}

// MCPRequest represents an MCP request
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// handleRequest handles an MCP request
func (s *Server) handleRequest(req *MCPRequest) *MCPResponse {
	resp := &MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = s.handleInitialize(req.Params)

	case "tools/list":
		resp.Result = s.handleToolsList()

	case "tools/call":
		result, err := s.handleToolCall(req.Params)
		if err != nil {
			resp.Error = &MCPError{
				Code:    -32603,
				Message: err.Error(),
			}
		} else {
			resp.Result = result
		}

	default:
		resp.Error = &MCPError{
			Code:    -32601,
			Message: fmt.Sprintf("method not found: %s", req.Method),
		}
	}

	return resp
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(params json.RawMessage) interface{} {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]bool{
				"listChanged": false,
			},
		},
		"serverInfo": map[string]string{
			"name":    "code-indexer-mcp",
			"version": "0.1.0",
		},
	}
}

// handleToolsList returns the list of available tools
func (s *Server) handleToolsList() interface{} {
	tools := make([]map[string]interface{}, 0, len(s.tools))

	for _, tool := range s.tools {
		tools = append(tools, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}

	return map[string]interface{}{
		"tools": tools,
	}
}

// handleToolCall handles a tool call
func (s *Server) handleToolCall(params json.RawMessage) (interface{}, error) {
	var req struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("failed to parse tool call: %w", err)
	}

	tool, ok := s.tools[req.Name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", req.Name)
	}

	result, err := tool.Handler(req.Arguments)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": formatResult(result),
			},
		},
	}, nil
}

// sendError sends an error response
func (s *Server) sendError(encoder *json.Encoder, id interface{}, err error) {
	response := &MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    -32603,
			Message: err.Error(),
		},
	}
	encoder.Encode(response)
}

// formatResult formats a result as a string
func formatResult(result interface{}) string {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", result)
	}
	return string(data)
}

// registerTools registers all available tools
func (s *Server) registerTools() {
	s.registerTool(&Tool{
		Name:        "search_symbols",
		Description: "Search for symbols (functions, classes, methods, variables) in the codebase",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query (symbol name)",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Symbol type filter (function, class, method, variable, etc.)",
				},
				"language": map[string]interface{}{
					"type":        "string",
					"description": "Language filter (go, python, typescript, etc.)",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Maximum number of results",
				},
			},
			"required": []string{"query"},
		},
		Handler: s.handleSearchSymbols,
	})

	s.registerTool(&Tool{
		Name:        "get_file_structure",
		Description: "Get the structure of a specific file (all symbols and imports)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file (relative or absolute)",
				},
			},
			"required": []string{"file_path"},
		},
		Handler: s.handleGetFileStructure,
	})

	s.registerTool(&Tool{
		Name:        "get_project_overview",
		Description: "Get an overview of the entire project (statistics, languages, etc.)",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: s.handleGetProjectOverview,
	})

	s.registerTool(&Tool{
		Name:        "index_project",
		Description: "Trigger a full re-index of the project",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: s.handleIndexProject,
	})

	s.registerTool(&Tool{
		Name:        "get_symbol_details",
		Description: "Get detailed information about a specific symbol (including references and relationships)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"symbol_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the symbol",
				},
			},
			"required": []string{"symbol_name"},
		},
		Handler: s.handleGetSymbolDetails,
	})

	s.registerTool(&Tool{
		Name:        "find_references",
		Description: "Find all references to a symbol in the codebase",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"symbol_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the symbol to find references for",
				},
			},
			"required": []string{"symbol_name"},
		},
		Handler: s.handleFindReferences,
	})

	s.registerTool(&Tool{
		Name:        "get_dependencies",
		Description: "Get dependencies for a specific file",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file (relative or absolute)",
				},
			},
			"required": []string{"file_path"},
		},
		Handler: s.handleGetDependencies,
	})

	s.registerTool(&Tool{
		Name:        "list_files",
		Description: "List all indexed files in the project",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"language": map[string]interface{}{
					"type":        "string",
					"description": "Filter by language (optional)",
				},
			},
		},
		Handler: s.handleListFiles,
	})
}

// registerTool registers a tool
func (s *Server) registerTool(tool *Tool) {
	s.tools[tool.Name] = tool
}

// Tool handlers

func (s *Server) handleSearchSymbols(params json.RawMessage) (interface{}, error) {
	var opts types.SearchOptions
	if err := json.Unmarshal(params, &opts); err != nil {
		return nil, err
	}

	symbols, err := s.indexer.SearchSymbols(opts)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"symbols": symbols,
		"count":   len(symbols),
	}, nil
}

func (s *Server) handleGetFileStructure(params json.RawMessage) (interface{}, error) {
	var req struct {
		FilePath string `json:"file_path"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	structure, err := s.indexer.GetFileStructure(req.FilePath)
	if err != nil {
		return nil, err
	}

	return structure, nil
}

func (s *Server) handleGetProjectOverview(params json.RawMessage) (interface{}, error) {
	overview, err := s.indexer.GetProjectOverview()
	if err != nil {
		return nil, err
	}

	return overview, nil
}

func (s *Server) handleIndexProject(params json.RawMessage) (interface{}, error) {
	if err := s.indexer.IndexAll(); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":  "success",
		"message": "Project indexed successfully",
	}, nil
}

func (s *Server) handleGetSymbolDetails(params json.RawMessage) (interface{}, error) {
	var req struct {
		SymbolName string `json:"symbol_name"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	details, err := s.indexer.GetSymbolDetails(req.SymbolName)
	if err != nil {
		return nil, err
	}

	return details, nil
}

func (s *Server) handleFindReferences(params json.RawMessage) (interface{}, error) {
	var req struct {
		SymbolName string `json:"symbol_name"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	references, err := s.indexer.FindReferences(req.SymbolName)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"symbol":     req.SymbolName,
		"references": references,
		"count":      len(references),
	}, nil
}

func (s *Server) handleGetDependencies(params json.RawMessage) (interface{}, error) {
	var req struct {
		FilePath string `json:"file_path"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	deps, err := s.indexer.GetDependencies(req.FilePath)
	if err != nil {
		return nil, err
	}

	return deps, nil
}

func (s *Server) handleListFiles(params json.RawMessage) (interface{}, error) {
	var req struct {
		Language string `json:"language"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	files, err := s.indexer.GetAllFiles()
	if err != nil {
		return nil, err
	}

	// Filter by language if specified
	if req.Language != "" {
		filtered := []*types.File{}
		for _, file := range files {
			if file.Language == req.Language {
				filtered = append(filtered, file)
			}
		}
		files = filtered
	}

	return map[string]interface{}{
		"files": files,
		"count": len(files),
	}, nil
}
