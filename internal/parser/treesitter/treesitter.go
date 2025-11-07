package treesitter

import (
	"fmt"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// TreeSitterParser is a generic parser using Tree-sitter
// Note: This is a placeholder/wrapper. Actual Tree-sitter bindings would be added per language
type TreeSitterParser struct {
	language   string
	extensions []string
	priority   int
	// grammar would be the tree-sitter grammar for this language
	// grammar *sitter.Language
}

// NewTreeSitterParser creates a new Tree-sitter based parser
func NewTreeSitterParser(language string, extensions []string) *TreeSitterParser {
	return &TreeSitterParser{
		language:   language,
		extensions: extensions,
		priority:   100, // Tree-sitter parsers have high priority
	}
}

// Language returns the language identifier
func (p *TreeSitterParser) Language() string {
	return p.language
}

// Extensions returns supported file extensions
func (p *TreeSitterParser) Extensions() []string {
	return p.extensions
}

// Priority returns parser priority
func (p *TreeSitterParser) Priority() int {
	return p.priority
}

// SupportsFramework checks framework support
func (p *TreeSitterParser) SupportsFramework(framework string) bool {
	// Framework support is handled by framework analyzers
	return false
}

// Parse parses file content using Tree-sitter
func (p *TreeSitterParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	// This is a placeholder for actual Tree-sitter integration
	// Real implementation would:
	// 1. Parse with tree-sitter
	// 2. Walk the syntax tree
	// 3. Extract symbols, imports, relationships

	return nil, fmt.Errorf("tree-sitter parser not yet implemented for %s", p.language)
}

// LanguageConfig holds configuration for a Tree-sitter language
type LanguageConfig struct {
	Name        string
	Extensions  []string
	Grammar     string   // Path to grammar or grammar name
	QueryFiles  []string // Paths to query files for symbol extraction
}

// Common Tree-sitter language configurations
var LanguageConfigs = map[string]*LanguageConfig{
	"typescript": {
		Name:       "typescript",
		Extensions: []string{".ts", ".tsx"},
		Grammar:    "tree-sitter-typescript",
		QueryFiles: []string{"queries/typescript/symbols.scm"},
	},
	"javascript": {
		Name:       "javascript",
		Extensions: []string{".js", ".jsx", ".mjs"},
		Grammar:    "tree-sitter-javascript",
		QueryFiles: []string{"queries/javascript/symbols.scm"},
	},
	"java": {
		Name:       "java",
		Extensions: []string{".java"},
		Grammar:    "tree-sitter-java",
		QueryFiles: []string{"queries/java/symbols.scm"},
	},
	"c_sharp": {
		Name:       "c_sharp",
		Extensions: []string{".cs"},
		Grammar:    "tree-sitter-c-sharp",
		QueryFiles: []string{"queries/c_sharp/symbols.scm"},
	},
	"c": {
		Name:       "c",
		Extensions: []string{".c", ".h"},
		Grammar:    "tree-sitter-c",
		QueryFiles: []string{"queries/c/symbols.scm"},
	},
	"cpp": {
		Name:       "cpp",
		Extensions: []string{".cpp", ".cc", ".cxx", ".hpp", ".h", ".hxx"},
		Grammar:    "tree-sitter-cpp",
		QueryFiles: []string{"queries/cpp/symbols.scm"},
	},
	"php": {
		Name:       "php",
		Extensions: []string{".php"},
		Grammar:    "tree-sitter-php",
		QueryFiles: []string{"queries/php/symbols.scm"},
	},
	"ruby": {
		Name:       "ruby",
		Extensions: []string{".rb"},
		Grammar:    "tree-sitter-ruby",
		QueryFiles: []string{"queries/ruby/symbols.scm"},
	},
	"rust": {
		Name:       "rust",
		Extensions: []string{".rs"},
		Grammar:    "tree-sitter-rust",
		QueryFiles: []string{"queries/rust/symbols.scm"},
	},
	"kotlin": {
		Name:       "kotlin",
		Extensions: []string{".kt", ".kts"},
		Grammar:    "tree-sitter-kotlin",
		QueryFiles: []string{"queries/kotlin/symbols.scm"},
	},
	"swift": {
		Name:       "swift",
		Extensions: []string{".swift"},
		Grammar:    "tree-sitter-swift",
		QueryFiles: []string{"queries/swift/symbols.scm"},
	},
	"bash": {
		Name:       "bash",
		Extensions: []string{".sh", ".bash"},
		Grammar:    "tree-sitter-bash",
		QueryFiles: []string{"queries/bash/symbols.scm"},
	},
	"sql": {
		Name:       "sql",
		Extensions: []string{".sql"},
		Grammar:    "tree-sitter-sql",
		QueryFiles: []string{"queries/sql/symbols.scm"},
	},
	"html": {
		Name:       "html",
		Extensions: []string{".html", ".htm"},
		Grammar:    "tree-sitter-html",
		QueryFiles: []string{"queries/html/symbols.scm"},
	},
	"css": {
		Name:       "css",
		Extensions: []string{".css", ".scss", ".sass", ".less"},
		Grammar:    "tree-sitter-css",
		QueryFiles: []string{"queries/css/symbols.scm"},
	},
	"powershell": {
		Name:       "powershell",
		Extensions: []string{".ps1", ".psm1", ".psd1"},
		Grammar:    "tree-sitter-powershell",
		QueryFiles: []string{"queries/powershell/symbols.scm"},
	},
}

// SymbolExtractor extracts symbols from Tree-sitter parse tree
// This would use Tree-sitter queries to find function/class/etc definitions
type SymbolExtractor struct {
	language string
	queries  []string // Tree-sitter query strings
}

// NewSymbolExtractor creates a symbol extractor for a language
func NewSymbolExtractor(language string) *SymbolExtractor {
	return &SymbolExtractor{
		language: language,
		queries:  getDefaultQueries(language),
	}
}

// getDefaultQueries returns default queries for common symbol types
func getDefaultQueries(language string) []string {
	// These would be actual Tree-sitter query strings
	// Different for each language syntax
	return []string{
		// Function definitions
		"(function_declaration name: (identifier) @name)",
		// Class definitions
		"(class_declaration name: (identifier) @name)",
		// Method definitions
		"(method_declaration name: (identifier) @name)",
	}
}

// Extract extracts symbols from parse tree (placeholder)
func (se *SymbolExtractor) Extract(tree interface{}) ([]*types.Symbol, error) {
	// This would walk the tree-sitter parse tree and extract symbols
	// using the configured queries
	return nil, fmt.Errorf("not implemented")
}

// Helper functions for language detection

// IsScriptLanguage checks if language is a scripting language
func IsScriptLanguage(lang string) bool {
	scriptLangs := map[string]bool{
		"python": true, "javascript": true, "ruby": true,
		"php": true, "bash": true, "powershell": true,
	}
	return scriptLangs[strings.ToLower(lang)]
}

// IsCompiledLanguage checks if language is compiled
func IsCompiledLanguage(lang string) bool {
	compiledLangs := map[string]bool{
		"go": true, "java": true, "c": true, "cpp": true,
		"c_sharp": true, "rust": true, "kotlin": true, "swift": true,
	}
	return compiledLangs[strings.ToLower(lang)]
}

// IsStaticallyTyped checks if language is statically typed
func IsStaticallyTyped(lang string) bool {
	staticLangs := map[string]bool{
		"go": true, "java": true, "c": true, "cpp": true,
		"c_sharp": true, "rust": true, "kotlin": true, "swift": true,
		"typescript": true,
	}
	return staticLangs[strings.ToLower(lang)]
}
