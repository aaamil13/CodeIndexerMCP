package sql

import (
	"regexp"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// SQLParser parses SQL source code
type SQLParser struct {
	*parser.BaseParser
}

// NewParser creates a new SQL parser
func NewParser() *SQLParser {
	return &SQLParser{
		BaseParser: parser.NewBaseParser("sql", []string{".sql"}, 100),
	}
}

// Parse parses SQL source code
func (p *SQLParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
		Metadata:      make(map[string]interface{}),
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
	singleLineRe := regexp.MustCompile(`--[^\n]*`)
	content = singleLineRe.ReplaceAllString(content, "")

	// Remove multi-line comments
	multiLineRe := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	content = multiLineRe.ReplaceAllString(content, "")

	return content
}

func (p *SQLParser) extractTables(content string, result *types.ParseResult) {
	// CREATE TABLE
	tableRe := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:[\w.]+\.)?(\w+)\s*\(`)

	matches := tableRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeClass, // Use class for tables
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "CREATE TABLE " + name,
			Metadata: map[string]interface{}{
				"table": true,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *SQLParser) extractViews(content string, result *types.ParseResult) {
	// CREATE VIEW
	viewRe := regexp.MustCompile(`(?i)CREATE\s+(?:OR\s+REPLACE\s+)?VIEW\s+(?:[\w.]+\.)?(\w+)\s+AS`)

	matches := viewRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeClass,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "CREATE VIEW " + name,
			Metadata: map[string]interface{}{
				"view": true,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *SQLParser) extractProcedures(content string, result *types.ParseResult) {
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

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeFunction,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  sig,
			Metadata: map[string]interface{}{
				"procedure": true,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *SQLParser) extractFunctions(content string, result *types.ParseResult) {
	// CREATE FUNCTION
	funcRe := regexp.MustCompile(`(?i)CREATE\s+(?:OR\s+REPLACE\s+)?FUNCTION\s+(?:[\w.]+\.)?(\w+)\s*\(([^)]*)\)\s*RETURNS?\s+([\w\s()]+)`)

	matches := funcRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		params := content[match[4]:match[5]]
		returnType := strings.TrimSpace(content[match[6]:match[7]])

		lineNum := strings.Count(content[:match[0]], "\n") + 1

		sig := "CREATE FUNCTION " + name + "(" + params + ") RETURNS " + returnType

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeFunction,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  sig,
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}

func (p *SQLParser) extractTriggers(content string, result *types.ParseResult) {
	// CREATE TRIGGER
	triggerRe := regexp.MustCompile(`(?i)CREATE\s+(?:OR\s+REPLACE\s+)?TRIGGER\s+(\w+)`)

	matches := triggerRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		lineNum := strings.Count(content[:match[0]], "\n") + 1

		symbol := &types.Symbol{
			Name:       name,
			Type:       types.SymbolTypeFunction,
			StartLine:  lineNum,
			EndLine:    lineNum,
			Visibility: types.VisibilityPublic,
			Signature:  "CREATE TRIGGER " + name,
			Metadata: map[string]interface{}{
				"trigger": true,
			},
		}

		result.Symbols = append(result.Symbols, symbol)
	}
}
