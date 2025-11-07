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
					"description": "Language filter (go, python, typescript, etc.)",
				},
			},
		},
		Handler: s.handleListFiles,
	})

	// AI-powered tools
	s.registerTool(&Tool{
		Name:        "get_code_context",
		Description: "Get comprehensive context for a symbol including usage examples, dependencies, and relationships",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"symbol_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the symbol",
				},
				"depth": map[string]interface{}{
					"type":        "number",
					"description": "Context depth (number of usage examples, default: 5)",
				},
			},
			"required": []string{"symbol_name"},
		},
		Handler: s.handleGetCodeContext,
	})

	s.registerTool(&Tool{
		Name:        "analyze_change_impact",
		Description: "Analyze the impact of changing or refactoring a symbol (risk level, affected files, suggestions)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"symbol_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the symbol to analyze",
				},
			},
			"required": []string{"symbol_name"},
		},
		Handler: s.handleAnalyzeChangeImpact,
	})

	s.registerTool(&Tool{
		Name:        "get_code_metrics",
		Description: "Calculate code quality metrics (complexity, maintainability, quality rating)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"symbol_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the symbol (function/method)",
				},
			},
			"required": []string{"symbol_name"},
		},
		Handler: s.handleGetCodeMetrics,
	})

	s.registerTool(&Tool{
		Name:        "extract_smart_snippet",
		Description: "Extract a self-contained code snippet with all dependencies and usage hints",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"symbol_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the symbol to extract",
				},
			},
			"required": []string{"symbol_name"},
		},
		Handler: s.handleExtractSmartSnippet,
	})

	s.registerTool(&Tool{
		Name:        "get_usage_statistics",
		Description: "Get detailed usage statistics for a symbol (usage count, patterns, files)",
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
		Handler: s.handleGetUsageStatistics,
	})

	s.registerTool(&Tool{
		Name:        "suggest_refactorings",
		Description: "Get AI-powered refactoring suggestions for a symbol",
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
		Handler: s.handleSuggestRefactorings,
	})

	s.registerTool(&Tool{
		Name:        "find_unused_symbols",
		Description: "Find unused/dead code in the project",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: s.handleFindUnusedSymbols,
	})

	// Change tracking tools
	s.registerTool(&Tool{
		Name:        "simulate_change",
		Description: "Simulate a code change and see its impact before applying it (shows affected files, broken references, auto-fix suggestions)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"symbol_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the symbol to change",
				},
				"change_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of change: add, modify, delete, rename, move",
				},
				"new_value": map[string]interface{}{
					"type":        "string",
					"description": "New value (e.g., new name for rename, new signature for modify)",
				},
			},
			"required": []string{"symbol_name", "change_type"},
		},
		Handler: s.handleSimulateChange,
	})

	s.registerTool(&Tool{
		Name:        "build_dependency_graph",
		Description: "Build a dependency graph for a symbol showing what it depends on and what depends on it",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"symbol_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the symbol",
				},
				"max_depth": map[string]interface{}{
					"type":        "number",
					"description": "Maximum depth to traverse (default: 3)",
				},
			},
			"required": []string{"symbol_name"},
		},
		Handler: s.handleBuildDependencyGraph,
	})

	s.registerTool(&Tool{
		Name:        "get_symbol_dependencies",
		Description: "Get all symbols that a given symbol depends on",
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
		Handler: s.handleGetSymbolDependencies,
	})

	s.registerTool(&Tool{
		Name:        "get_symbol_dependents",
		Description: "Get all symbols that depend on a given symbol",
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
		Handler: s.handleGetSymbolDependents,
	})

	// Type validation tools
	s.registerTool(&Tool{
		Name:        "validate_file_types",
		Description: "Validate all types in a file and find undefined usages, type mismatches, and missing methods",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to validate",
				},
			},
			"required": []string{"file_path"},
		},
		Handler: s.handleValidateFileTypes,
	})

	s.registerTool(&Tool{
		Name:        "find_undefined_usages",
		Description: "Find all undefined symbol usages in a file (methods, functions, variables that don't exist)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to check",
				},
			},
			"required": []string{"file_path"},
		},
		Handler: s.handleFindUndefinedUsages,
	})

	s.registerTool(&Tool{
		Name:        "check_method_exists",
		Description: "Check if a specific method exists on a type/class",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"type_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the type/class",
				},
				"method_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the method to check",
				},
			},
			"required": []string{"type_name", "method_name"},
		},
		Handler: s.handleCheckMethodExists,
	})

	s.registerTool(&Tool{
		Name:        "calculate_type_safety_score",
		Description: "Calculate type safety score for a file (0-100, with type coverage and error metrics)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file",
				},
			},
			"required": []string{"file_path"},
		},
		Handler: s.handleCalculateTypeSafetyScore,
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

// AI-powered tool handlers

func (s *Server) handleGetCodeContext(params json.RawMessage) (interface{}, error) {
	var req struct {
		SymbolName string `json:"symbol_name"`
		Depth      int    `json:"depth"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	if req.Depth == 0 {
		req.Depth = 5 // Default depth
	}

	context, err := s.indexer.GetCodeContext(req.SymbolName, req.Depth)
	if err != nil {
		return nil, err
	}

	return context, nil
}

func (s *Server) handleAnalyzeChangeImpact(params json.RawMessage) (interface{}, error) {
	var req struct {
		SymbolName string `json:"symbol_name"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	impact, err := s.indexer.AnalyzeChangeImpact(req.SymbolName)
	if err != nil {
		return nil, err
	}

	return impact, nil
}

func (s *Server) handleGetCodeMetrics(params json.RawMessage) (interface{}, error) {
	var req struct {
		SymbolName string `json:"symbol_name"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	metrics, err := s.indexer.GetCodeMetrics(req.SymbolName)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (s *Server) handleExtractSmartSnippet(params json.RawMessage) (interface{}, error) {
	var req struct {
		SymbolName string `json:"symbol_name"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	snippet, err := s.indexer.ExtractSmartSnippet(req.SymbolName)
	if err != nil {
		return nil, err
	}

	return snippet, nil
}

func (s *Server) handleGetUsageStatistics(params json.RawMessage) (interface{}, error) {
	var req struct {
		SymbolName string `json:"symbol_name"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	stats, err := s.indexer.GetUsageStatistics(req.SymbolName)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *Server) handleSuggestRefactorings(params json.RawMessage) (interface{}, error) {
	var req struct {
		SymbolName string `json:"symbol_name"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	suggestions, err := s.indexer.SuggestRefactorings(req.SymbolName)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"symbol":        req.SymbolName,
		"opportunities": suggestions,
		"count":         len(suggestions),
	}, nil
}

func (s *Server) handleFindUnusedSymbols(params json.RawMessage) (interface{}, error) {
	unused, err := s.indexer.FindUnusedSymbols()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"unused_symbols": unused,
		"count":          len(unused),
		"suggestion":     "These symbols may be candidates for removal or refactoring",
	}, nil
}

// Change tracking tool handlers

func (s *Server) handleSimulateChange(params json.RawMessage) (interface{}, error) {
	var req struct {
		SymbolName string `json:"symbol_name"`
		ChangeType string `json:"change_type"`
		NewValue   string `json:"new_value"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	// Parse change type
	var changeType types.ChangeType
	switch req.ChangeType {
	case "add":
		changeType = types.ChangeTypeAdd
	case "modify":
		changeType = types.ChangeTypeModify
	case "delete":
		changeType = types.ChangeTypeDelete
	case "rename":
		changeType = types.ChangeTypeRename
	case "move":
		changeType = types.ChangeTypeMove
	default:
		return nil, fmt.Errorf("invalid change type: %s (must be: add, modify, delete, rename, move)", req.ChangeType)
	}

	impact, err := s.indexer.SimulateSymbolChange(req.SymbolName, changeType, req.NewValue)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"symbol":               req.SymbolName,
		"change_type":          req.ChangeType,
		"new_value":            req.NewValue,
		"affected_symbols":     len(impact.AffectedSymbols),
		"broken_references":    len(impact.BrokenReferences),
		"required_updates":     len(impact.RequiredUpdates),
		"validation_errors":    len(impact.ValidationErrors),
		"auto_fix_suggestions": len(impact.AutoFixSuggestions),
		"can_auto_fix":         impact.CanAutoFix,
		"details":              impact,
	}, nil
}

func (s *Server) handleBuildDependencyGraph(params json.RawMessage) (interface{}, error) {
	var req struct {
		SymbolName string `json:"symbol_name"`
		MaxDepth   int    `json:"max_depth"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	if req.MaxDepth == 0 {
		req.MaxDepth = 3 // Default depth
	}

	graph, err := s.indexer.BuildDependencyGraph(req.SymbolName, req.MaxDepth)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"symbol":            req.SymbolName,
		"total_nodes":       len(graph.Nodes),
		"total_edges":       len(graph.Edges),
		"direct_dependencies": graph.DirectDependencies,
		"direct_dependents": graph.DirectDependents,
		"coupling_score":    graph.CouplingScore,
		"graph":             graph,
	}, nil
}

func (s *Server) handleGetSymbolDependencies(params json.RawMessage) (interface{}, error) {
	var req struct {
		SymbolName string `json:"symbol_name"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	dependencies, err := s.indexer.GetDependencies(req.SymbolName)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"symbol":       req.SymbolName,
		"dependencies": dependencies,
		"count":        len(dependencies),
	}, nil
}

func (s *Server) handleGetSymbolDependents(params json.RawMessage) (interface{}, error) {
	var req struct {
		SymbolName string `json:"symbol_name"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	dependents, err := s.indexer.GetDependents(req.SymbolName)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"symbol":     req.SymbolName,
		"dependents": dependents,
		"count":      len(dependents),
	}, nil
}

// Type validation tool handlers

func (s *Server) handleValidateFileTypes(params json.RawMessage) (interface{}, error) {
	var req struct {
		FilePath string `json:"file_path"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	validation, err := s.indexer.ValidateFileTypes(req.FilePath)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"file":                req.FilePath,
		"is_valid":            validation.IsValid,
		"undefined_count":     len(validation.UndefinedSymbols),
		"type_mismatch_count": len(validation.TypeMismatches),
		"missing_method_count": len(validation.MissingMethods),
		"invalid_call_count":  len(validation.InvalidCalls),
		"unused_import_count": len(validation.UnusedImports),
		"validation":          validation,
	}, nil
}

func (s *Server) handleFindUndefinedUsages(params json.RawMessage) (interface{}, error) {
	var req struct {
		FilePath string `json:"file_path"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	undefined, err := s.indexer.FindUndefinedUsages(req.FilePath)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"file":              req.FilePath,
		"undefined_usages":  undefined,
		"count":             len(undefined),
		"has_errors":        len(undefined) > 0,
	}, nil
}

func (s *Server) handleCheckMethodExists(params json.RawMessage) (interface{}, error) {
	var req struct {
		TypeName   string `json:"type_name"`
		MethodName string `json:"method_name"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	result, err := s.indexer.CheckMethodExists(req.TypeName, req.MethodName)
	if err != nil {
		return nil, err
	}

	if result == nil {
		// Method exists
		return map[string]interface{}{
			"type_name":    req.TypeName,
			"method_name":  req.MethodName,
			"exists":       true,
			"message":      fmt.Sprintf("Method '%s' exists on type '%s'", req.MethodName, req.TypeName),
		}, nil
	}

	// Method doesn't exist
	return map[string]interface{}{
		"type_name":         req.TypeName,
		"method_name":       req.MethodName,
		"exists":            false,
		"available_methods": result.AvailableMethods,
		"suggestion":        result.Suggestion,
		"message":           fmt.Sprintf("Method '%s' not found on type '%s'", req.MethodName, req.TypeName),
	}, nil
}

func (s *Server) handleCalculateTypeSafetyScore(params json.RawMessage) (interface{}, error) {
	var req struct {
		FilePath string `json:"file_path"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return nil, err
	}

	score, err := s.indexer.CalculateTypeSafetyScore(req.FilePath)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"file":              req.FilePath,
		"score":             score.Score,
		"rating":            score.Rating,
		"typed_symbols":     score.TypedSymbols,
		"untyped_symbols":   score.UntypedSymbols,
		"error_count":       score.ErrorCount,
		"warning_count":     score.WarningCount,
		"recommendation":    score.Recommendation,
		"details":           score,
	}, nil
}
