package rst

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// ReStructuredTextParser parses reStructuredText documentation files
type ReStructuredTextParser struct {
}

// NewParser creates a new reStructuredText parser
func NewParser() *ReStructuredTextParser {
	return &ReStructuredTextParser{}
}

// Language returns the language identifier (e.g., "rst")
func (p *ReStructuredTextParser) Language() string {
	return "rst"
}

// Extensions returns file extensions this parser handles (e.g., [".rst", ".rest"])
func (p *ReStructuredTextParser) Extensions() []string {
	return []string{".rst", ".rest"}
}

// Priority returns parser priority (higher = preferred when multiple parsers match)
func (p *ReStructuredTextParser) Priority() int {
	return 50
}

// SupportsFramework checks if parser supports specific framework analysis
func (p *ReStructuredTextParser) SupportsFramework(framework string) bool {
	return false
}

// Parse parses reStructuredText content
func (p *ReStructuredTextParser) Parse(content []byte, filePath string) (*parsing.ParseResult, error) {
	result := &parsing.ParseResult{
		Symbols:       make([]*model.Symbol, 0),
		Imports:       make([]*model.Import, 0),
		Relationships: make([]*model.Relationship, 0),
		Metadata:      make(map[string]interface{}),
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Extract sections (headers)
	p.extractSections(lines, result)

	// Extract directives
	p.extractDirectives(lines, result)

	// Extract references
	p.extractReferences(lines, result)

	result.Metadata["language"] = "rst"
	result.Metadata["type"] = "documentation"

	return result, nil
}

func (p *ReStructuredTextParser) extractSections(lines []string, result *parsing.ParseResult) {
	// RST headers: underline with =, -, ~, etc.
	// Title
	// =====
	// or
	// ======
	// Title
	// ======

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			continue
		}

		// Check if next line is an underline
		if i+1 < len(lines) {
			nextLine := lines[i+1]
			if p.isUnderline(nextLine) && len(nextLine) >= len(strings.TrimSpace(line)) {
				title := strings.TrimSpace(line)
				if title != "" {
					level := p.getHeaderLevel(nextLine[0])

					symbol := &model.Symbol{
						Name:       title,
						Kind:       model.SymbolKindVariable,
						File:       "", // File path will be set by the caller
						Range:      model.Range{Start: model.Position{Line: i + 1}, End: model.Position{Line: i + 1}},
						Visibility: model.VisibilityPublic,
						Signature:  title,
						Metadata: map[string]string{
							"header":       "true",
							"header_level": strconv.Itoa(level),
						},
					}

					result.Symbols = append(result.Symbols, symbol)
					i++ // Skip the underline
					continue
				}
			}
		}

		// Check if this is an overline (previous line is also a line of symbols)
		if i > 0 {
			prevLine := lines[i-1]
			if i+1 < len(lines) {
				nextLine := lines[i+1]
				if p.isUnderline(prevLine) && p.isUnderline(nextLine) &&
					len(prevLine) >= len(strings.TrimSpace(line)) &&
					len(nextLine) >= len(strings.TrimSpace(line)) {

					title := strings.TrimSpace(line)
					if title != "" {
						level := p.getHeaderLevel(prevLine[0])

						symbol := &model.Symbol{
							Name:       title,
							Kind:       model.SymbolKindVariable,
							File:       "", // File path will be set by the caller
							Range:      model.Range{Start: model.Position{Line: i + 1}, End: model.Position{Line: i + 1}},
							Visibility: model.VisibilityPublic,
							Signature:  title,
							Metadata: map[string]string{
								"header":       "true",
								"header_level": strconv.Itoa(level),
								"overline":     "true",
							},
						}

						result.Symbols = append(result.Symbols, symbol)
					}
				}
			}
		}
	}
}

func (p *ReStructuredTextParser) isUnderline(line string) bool {
	if len(line) == 0 {
		return false
	}

	// Check if line consists of a single repeated character
	firstChar := line[0]
	underlineChars := "=-~`:#\"'^_*+"

	if !strings.ContainsRune(underlineChars, rune(firstChar)) {
		return false
	}

	for _, ch := range line {
		if ch != rune(firstChar) && ch != ' ' {
			return false
		}
	}

	return true
}

func (p *ReStructuredTextParser) getHeaderLevel(char byte) int {
	// Common convention for header levels in RST
	levels := map[byte]int{
		'#': 1, // With overline
		'*': 1, // With overline
		'=': 2,
		'-': 3,
		'^': 4,
		'"': 5,
		'~': 6,
	}

	if level, ok := levels[char]; ok {
		return level
	}
	return 3 // Default
}

func (p *ReStructuredTextParser) extractDirectives(lines []string, result *parsing.ParseResult) {
	// Directives: .. directive:: content
	directiveRe := regexp.MustCompile(`^\.\.\s+([\w-]+)::\s*(.*)`)

	for i, line := range lines {
		if matches := directiveRe.FindStringSubmatch(line); matches != nil {
			directive := matches[1]
			content := matches[2]

			// Skip common directives that are not interesting
			if directive == "note" || directive == "warning" || directive == "code-block" {
				continue
			}

			name := directive
			if content != "" {
				name = directive + ": " + content
			}

			symbol := &model.Symbol{
				Name:       name,
				Kind:       model.SymbolKindFunction,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: i + 1}, End: model.Position{Line: i + 1}},
				Visibility: model.VisibilityPublic,
				Signature:  ".. " + directive + "::",
				Metadata: map[string]string{
					"directive": directive,
				},
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}
}

func (p *ReStructuredTextParser) extractReferences(lines []string, result *parsing.ParseResult) {
	// Reference definitions: .. _label: or .. _label:
	refRe := regexp.MustCompile(`^\.\.\s+_([^:]+):\s*(.*)`)

	for i, line := range lines {
		if matches := refRe.FindStringSubmatch(line); matches != nil {
			label := matches[1]
			target := matches[2]

			sig := "_" + label
			if target != "" {
				sig += ": " + target
			}

			symbol := &model.Symbol{
				Name:       label,
				Kind:       model.SymbolKindConstant,
				File:       "", // File path will be set by the caller
				Range:      model.Range{Start: model.Position{Line: i + 1}, End: model.Position{Line: i + 1}},
				Visibility: model.VisibilityPublic,
				Signature:  sig,
				Metadata: map[string]string{
					"reference": "true",
				},
			}

			result.Symbols = append(result.Symbols, symbol)
		}
	}
}
