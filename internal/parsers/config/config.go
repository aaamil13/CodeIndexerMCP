package config

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// JSONParser parses JSON configuration files
type JSONParser struct {
	*parser.BaseParser
}

// NewJSONParser creates a new JSON parser
func NewJSONParser() *JSONParser {
	return &JSONParser{
		BaseParser: parser.NewBaseParser("json", []string{".json"}, 50),
	}
}

// Parse parses JSON content
func (p *JSONParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	// Parse JSON structure
	var data interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		result.Errors = append(result.Errors, types.ParseError{
			Line:    1,
			Column:  1,
			Message: fmt.Sprintf("JSON parse error: %v", err),
		})
		return result, nil // Return partial result
	}

	// Extract configuration keys as "symbols"
	p.extractKeys(data, "", result, 1)

	result.Metadata["type"] = "configuration"
	result.Metadata["format"] = "json"

	return result, nil
}

func (p *JSONParser) extractKeys(data interface{}, prefix string, result *types.ParseResult, line int) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}

			symbol := &types.Symbol{
				Name:       fullKey,
				Type:       types.SymbolTypeVariable,
				StartLine:  line,
				EndLine:    line,
				Visibility: types.VisibilityPublic,
			}

			// Set signature based on value type
			switch value.(type) {
			case map[string]interface{}:
				symbol.Signature = fmt.Sprintf("%s: object", fullKey)
				p.extractKeys(value, fullKey, result, line+1)
			case []interface{}:
				symbol.Signature = fmt.Sprintf("%s: array", fullKey)
			case string:
				symbol.Signature = fmt.Sprintf("%s: string", fullKey)
			case float64:
				symbol.Signature = fmt.Sprintf("%s: number", fullKey)
			case bool:
				symbol.Signature = fmt.Sprintf("%s: boolean", fullKey)
			default:
				symbol.Signature = fmt.Sprintf("%s: unknown", fullKey)
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	case []interface{}:
		for i, item := range v {
			indexKey := fmt.Sprintf("%s[%d]", prefix, i)
			p.extractKeys(item, indexKey, result, line+i)
		}
	}
}

// YAMLParser parses YAML configuration files
type YAMLParser struct {
	*parser.BaseParser
}

// NewYAMLParser creates a new YAML parser
func NewYAMLParser() *YAMLParser {
	return &YAMLParser{
		BaseParser: parser.NewBaseParser("yaml", []string{".yaml", ".yml"}, 50),
	}
}

// Parse parses YAML content (simplified - real implementation would use yaml library)
func (p *YAMLParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	// Basic YAML parsing (would use gopkg.in/yaml.v3 in real implementation)
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Simple key detection
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key := strings.TrimSpace(parts[0])

			symbol := &types.Symbol{
				Name:       key,
				Type:       types.SymbolTypeVariable,
				StartLine:  i + 1,
				EndLine:    i + 1,
				Visibility: types.VisibilityPublic,
				Signature:  line,
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}

	result.Metadata["type"] = "configuration"
	result.Metadata["format"] = "yaml"

	return result, nil
}

// XMLParser parses XML configuration files
type XMLParser struct {
	*parser.BaseParser
}

// NewXMLParser creates a new XML parser
func NewXMLParser() *XMLParser {
	return &XMLParser{
		BaseParser: parser.NewBaseParser("xml", []string{".xml"}, 50),
	}
}

// Parse parses XML content
func (p *XMLParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	// Parse XML structure
	type Node struct {
		XMLName xml.Name
		Attrs   []xml.Attr `xml:",any,attr"`
		Content []byte     `xml:",innerxml"`
	}

	var root Node
	if err := xml.Unmarshal(content, &root); err != nil {
		result.Errors = append(result.Errors, types.ParseError{
			Line:    1,
			Column:  1,
			Message: fmt.Sprintf("XML parse error: %v", err),
		})
		return result, nil
	}

	// Create symbol for root element
	symbol := &types.Symbol{
		Name:       root.XMLName.Local,
		Type:       types.SymbolTypeVariable,
		StartLine:  1,
		EndLine:    1,
		Visibility: types.VisibilityPublic,
		Signature:  fmt.Sprintf("<%s>", root.XMLName.Local),
	}

	result.Symbols = append(result.Symbols, symbol)
	result.Metadata["type"] = "configuration"
	result.Metadata["format"] = "xml"
	result.Metadata["root_element"] = root.XMLName.Local

	return result, nil
}

// TOMLParser parses TOML configuration files
type TOMLParser struct {
	*parser.BaseParser
}

// NewTOMLParser creates a new TOML parser
func NewTOMLParser() *TOMLParser {
	return &TOMLParser{
		BaseParser: parser.NewBaseParser("toml", []string{".toml"}, 50),
	}
}

// Parse parses TOML content (simplified)
func (p *TOMLParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	// Basic TOML parsing (would use github.com/BurntSushi/toml in real implementation)
	lines := strings.Split(string(content), "\n")
	var currentSection string

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")

			symbol := &types.Symbol{
				Name:       currentSection,
				Type:       types.SymbolTypeClass, // Use class for sections
				StartLine:  i + 1,
				EndLine:    i + 1,
				Visibility: types.VisibilityPublic,
				Signature:  line,
			}

			result.Symbols = append(result.Symbols, symbol)
			continue
		}

		// Key-value pairs
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])

			fullKey := key
			if currentSection != "" {
				fullKey = currentSection + "." + key
			}

			symbol := &types.Symbol{
				Name:       fullKey,
				Type:       types.SymbolTypeVariable,
				StartLine:  i + 1,
				EndLine:    i + 1,
				Visibility: types.VisibilityPublic,
				Signature:  line,
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}

	result.Metadata["type"] = "configuration"
	result.Metadata["format"] = "toml"

	return result, nil
}

// MarkdownParser parses Markdown documentation files
type MarkdownParser struct {
	*parser.BaseParser
}

// NewMarkdownParser creates a new Markdown parser
func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{
		BaseParser: parser.NewBaseParser("markdown", []string{".md", ".markdown"}, 50),
	}
}

// Parse parses Markdown content
func (p *MarkdownParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
	result := &types.ParseResult{
		Symbols:       make([]*types.Symbol, 0),
		Imports:       make([]*types.Import, 0),
		Relationships: make([]*types.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Headers as symbols
		if strings.HasPrefix(line, "#") {
			level := 0
			for j := 0; j < len(line) && line[j] == '#'; j++ {
				level++
			}

			headerText := strings.TrimSpace(line[level:])

			symbolType := types.SymbolTypeClass
			if level == 1 {
				symbolType = types.SymbolTypeClass
			} else if level == 2 {
				symbolType = types.SymbolTypeFunction
			} else {
				symbolType = types.SymbolTypeVariable
			}

			symbol := &types.Symbol{
				Name:          headerText,
				Type:          symbolType,
				StartLine:     i + 1,
				EndLine:       i + 1,
				Visibility:    types.VisibilityPublic,
				Signature:     line,
				Documentation: headerText,
			}

			result.Symbols = append(result.Symbols, symbol)
		}

		// Code blocks (```language)
		if strings.HasPrefix(line, "```") {
			lang := strings.TrimPrefix(line, "```")
			if lang != "" {
				result.Metadata["code_languages"] = append(
					result.Metadata["code_languages"].([]string),
					lang,
				)
			}
		}
	}

	result.Metadata["type"] = "documentation"
	result.Metadata["format"] = "markdown"

	return result, nil
}
