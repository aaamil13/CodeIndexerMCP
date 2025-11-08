package config

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// JSONParser parses JSON configuration files
type JSONParser struct {
}

// NewJSONParser creates a new JSON parser
func NewJSONParser() *JSONParser {
	return &JSONParser{}
}

// Language returns the language identifier
func (p *JSONParser) Language() string {
	return "json"
}

// Extensions returns file extensions this parser handles
func (p *JSONParser) Extensions() []string {
	return []string{".json"}
}

// Priority returns parser priority
func (p *JSONParser) Priority() int {
	return 50
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *JSONParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses JSON content
func (p *JSONParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:  make([]*model.Symbol, 0),
		Imports:  make([]*model.Import, 0),
		Metadata: make(map[string]interface{}),
	}

	// Parse JSON structure
	var data interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		result.ParseErrors = append(result.ParseErrors, parsing.ParseError{
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

func (p *JSONParser) extractKeys(data interface{}, prefix string, result *parsing.ParseResult, line int) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}

			symbol := &model.Symbol{
				Name:       fullKey,
				Kind:       model.SymbolKindVariable,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: line}, End: model.Position{Line: line}},
				Visibility: model.VisibilityPublic,
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
}

// NewYAMLParser creates a new YAML parser
func NewYAMLParser() *YAMLParser {
	return &YAMLParser{}
}

// Language returns the language identifier
func (p *YAMLParser) Language() string {
	return "yaml"
}

// Extensions returns file extensions this parser handles
func (p *YAMLParser) Extensions() []string {
	return []string{".yaml", ".yml"}
}

// Priority returns parser priority
func (p *YAMLParser) Priority() int {
	return 50
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *YAMLParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses YAML content (simplified - real implementation would use yaml library)
func (p *YAMLParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:  make([]*model.Symbol, 0),
		Imports:  make([]*model.Import, 0),
		Metadata: make(map[string]interface{}),
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

			symbol := &model.Symbol{
				Name:       key,
				Kind:       model.SymbolKindVariable,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: i + 1}, End: model.Position{Line: i + 1}},
				Visibility: model.VisibilityPublic,
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
}

// NewXMLParser creates a new XML parser
func NewXMLParser() *XMLParser {
	return &XMLParser{}
}

// Language returns the language identifier
func (p *XMLParser) Language() string {
	return "xml"
}

// Extensions returns file extensions this parser handles
func (p *XMLParser) Extensions() []string {
	return []string{".xml"}
}

// Priority returns parser priority
func (p *XMLParser) Priority() int {
	return 50
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *XMLParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses XML content
func (p *XMLParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:  make([]*model.Symbol, 0),
		Imports:  make([]*model.Import, 0),
		Metadata: make(map[string]interface{}),
	}

	// Parse XML structure
	type Node struct {
		XMLName xml.Name
		Attrs   []xml.Attr `xml:",any,attr"`
		Content []byte     `xml:",innerxml"`
	}

	var root Node
	if err := xml.Unmarshal(content, &root); err != nil {
		result.ParseErrors = append(result.ParseErrors, parsing.ParseError{
			Line:    1,
			Column:  1,
			Message: fmt.Sprintf("XML parse error: %v", err),
		})
		return result, nil
	}

	// Create symbol for root element
	symbol := &model.Symbol{
		Name:       root.XMLName.Local,
		Kind:       model.SymbolKindVariable, // Using variable kind for XML elements for now
		File:       "",                       // File path will be set by the caller
		Range:      model.Range{Start: model.Position{Line: 1}, End: model.Position{Line: 1}},
		Visibility: model.VisibilityPublic,
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
}

// NewTOMLParser creates a new TOML parser
func NewTOMLParser() *TOMLParser {
	return &TOMLParser{}
}

// Language returns the language identifier
func (p *TOMLParser) Language() string {
	return "toml"
}

// Extensions returns file extensions this parser handles
func (p *TOMLParser) Extensions() []string {
	return []string{".toml"}
}

// Priority returns parser priority
func (p *TOMLParser) Priority() int {
	return 50
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *TOMLParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses TOML content (simplified)
func (p *TOMLParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:  make([]*model.Symbol, 0),
		Imports:  make([]*model.Import, 0),
		Metadata: make(map[string]interface{}),
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

			symbol := &model.Symbol{
				Name:       currentSection,
				Kind:       model.SymbolKindClass, // Use class for sections
				File:       "",                    // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: i + 1}, End: model.Position{Line: i + 1}},
				Visibility: model.VisibilityPublic,
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

			symbol := &model.Symbol{
				Name:       fullKey,
				Kind:       model.SymbolKindVariable,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: i + 1}, End: model.Position{Line: i + 1}},
				Visibility: model.VisibilityPublic,
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
}

// NewMarkdownParser creates a new Markdown parser
func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{}
}

// Language returns the language identifier
func (p *MarkdownParser) Language() string {
	return "markdown"
}

// Extensions returns file extensions this parser handles
func (p *MarkdownParser) Extensions() []string {
	return []string{".md", ".markdown"}
}

// Priority returns parser priority
func (p *MarkdownParser) Priority() int {
	return 50
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *MarkdownParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses Markdown content
func (p *MarkdownParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:  make([]*model.Symbol, 0),
		Imports:  make([]*model.Import, 0),
		Metadata: make(map[string]interface{}),
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

			symbolKind := model.SymbolKindClass
			if level == 1 {
				symbolKind = model.SymbolKindClass
			} else if level == 2 {
				symbolKind = model.SymbolKindFunction
			} else {
				symbolKind = model.SymbolKindVariable
			}

			symbol := &model.Symbol{
				Name:          headerText,
				Kind:          symbolKind,
				File:          "", // File path will be set by the caller
				Range:         model.Range{Start: model.Position{Line: i + 1}, End: model.Position{Line: i + 1}},
				Visibility:    model.VisibilityPublic,
				Signature:     line,
				Documentation: headerText,
			}

			result.Symbols = append(result.Symbols, symbol)
		}

		// Code blocks (```language)
		if strings.HasPrefix(line, "```") {
			lang := strings.TrimPrefix(line, "```")
			if lang != "" {
				// Ensure "code_languages" is initialized as a slice of strings
				if _, ok := result.Metadata["code_languages"]; !ok {
					result.Metadata["code_languages"] = []string{}
				}
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
