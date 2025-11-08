package sql

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// SQLParser parses SQL source code
type SQLParser struct {
}

// NewParser creates a new SQL parser
func NewParser() *SQLParser {
	return &SQLParser{}
}

// Language returns the language identifier (e.g., "sql")
func (p *SQLParser) Language() string {
	return "sql"
}

// Extensions returns file extensions this parser handles (e.g., [".sql"])
func (p *SQLParser) Extensions() []string {
	return []string{".sql"}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *SQLParser) Priority() int {
	return 100
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *SQLParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses SQL source code
func (p *SQLParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:  make([]*model.Symbol, 0),
		Imports:  make([]*model.Import, 0),
		Metadata: make(map[string]interface{}),
	}

	contentStr := string(content)

	// Normalize content (remove comments)
	contentStr = p.removeComments(contentStr)

	// Extract tables
	p.extractTables(contentStr, result)

	// Extract views
	p.extractViews(contentStr, result)

	// Extract procedures
	p.extractProcedures(contentStr, result)

	// Extract functions
	p.extractFunctions(contentStr, result)

	// Extract triggers
	p.extractTriggers(contentStr, result)

	result.Metadata["language"] = "sql"

	return result, nil
}

func (p *SQLParser) removeComments(content string) string {
	// Remove single-line comments
	singleLineRe := regexp.MustCompile(`--[^
]*`)
	content = singleLineRe.ReplaceAllString(content, "")

	// Remove multi-line comments
	multiLineRe := regexp.MustCompile(`/\[\s\S]*?\*/`)
	content = multiLineRe.ReplaceAllString(content, "")

	return content
}

func (p *SQLParser) extractTables(content string, result *parsing.ParseResult) {
	// CREATE TABLE
	tableRe := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:[\w.]+\.)?(\w+)\s*\(`)

	matches := tableRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindClass, // Use class for tables
			File:       "",                    // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  "CREATE TABLE " + name,
			Metadata: map[string]string{
				"table": "true",
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *SQLParser) extractViews(content string, result *parsing.ParseResult) {
	// CREATE VIEW
	viewRe := regexp.MustCompile(`(?i)CREATE\s+(?:OR\s+REPLACE\s+)?VIEW\s+(?:[\w.]+\.)?(\w+)\s+AS`)

	matches := viewRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindClass,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  "CREATE VIEW " + name,
			Metadata: map[string]string{
				"view": "true",
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *SQLParser) extractProcedures(content string, result *parsing.ParseResult) {
	// CREATE PROCEDURE / CREATE PROC
	procRe := regexp.MustCompile(`(?i)CREATE\s+(?:OR\s+REPLACE\s+)?(?:PROCEDURE|PROC)\s+(?:[\w.]+\.)?(\w+)\s*(?:\(([^)]*)\))?`)

	matches := procRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		params := ""
		if match[4] != -1 {
			params = content[match[4]:match[5]]
		}

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		sig := "CREATE PROCEDURE " + name
		if params != "" {
			sig += "(" + params + ")"
		}

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindFunction,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  sig,
			Metadata: map[string]string{
				"procedure": "true",
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *SQLParser) extractFunctions(content string, result *parsing.ParseResult) {
	// CREATE FUNCTION
	funcRe := regexp.MustCompile(`(?i)CREATE\s+(?:OR\s+REPLACE\s+)?FUNCTION\s+(?:[\w.]+\.)?(\w+)\s*\(([^)]*)\)\s*RETURNS?\s+([\w\s()]+)`)

	matches := funcRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		params := content[match[4]:match[5]]
		returnType := strings.TrimSpace(content[match[6]:match[7]])

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		sig := "CREATE FUNCTION " + name + "(" + params + ") RETURNS " + returnType

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindFunction,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  sig,
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *SQLParser) extractTriggers(content string, result *parsing.ParseResult) {
	// CREATE TRIGGER
	triggerRe := regexp.MustCompile(`(?i)CREATE\s+(?:OR\s+REPLACE\s+)?TRIGGER\s+(\w+)`)

	matches := triggerRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &model.Symbol{
			Name:       name,
			Kind:       model.SymbolKindFunction,
			File:       "", // File path will be set by the caller
			Range:      model.Range{Start: model.Position{Line: lineNum}, End: model.Position{Line: lineNum}},
			Visibility: model.VisibilityPublic,
			Signature:  "CREATE TRIGGER " + name,
			Metadata: map[string]string{
				"trigger": "true",
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}
