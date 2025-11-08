package parser

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
)

// ParserPlugin represents a language parser plugin
type ParserPlugin interface {
	// Language returns the language identifier (e.g., "go", "python", "typescript")
	Language() string

	// Extensions returns file extensions this parser handles (e.g., [".go", ".mod"])
	Extensions() []string

	// Parse parses file content and returns symbols
	Parse(content []byte, filePath string) (*parsing.ParseResult, error)

	// SupportsFramework checks if parser supports specific framework analysis
	SupportsFramework(framework string) bool

	// Priority returns parser priority (higher = preferred when multiple parsers match)
	Priority() int
}

// FrameworkAnalyzer represents a framework-specific analyzer
type FrameworkAnalyzer interface {
	// Framework returns the framework identifier (e.g., "react", "django", "rails")
	Framework() string

	// Language returns the target language
	Language() string

	// Analyze analyzes code for framework-specific patterns
	Analyze(result *parsing.ParseResult, content []byte) (*model.FrameworkInfo, error)

	// DetectFramework detects if file uses this framework
	DetectFramework(content []byte, filePath string) bool
}

// Registry manages parser plugins and framework analyzers
type Registry struct {
	parsers    map[string][]ParserPlugin      // language -> parsers
	extensions map[string][]ParserPlugin      // extension -> parsers
	analyzers  map[string][]FrameworkAnalyzer // language -> analyzers
	mu         sync.RWMutex
}

// NewRegistry creates a new parser registry
func NewRegistry() *Registry {
	return &Registry{
		parsers:    make(map[string][]ParserPlugin),
		extensions: make(map[string][]ParserPlugin),
		analyzers:  make(map[string][]FrameworkAnalyzer),
	}
}

// CanParse checks if any registered parser can handle the given file
func (r *Registry) CanParse(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))

	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.extensions[ext]
	return ok
}

// RegisterParser registers a parser plugin
func (r *Registry) RegisterParser(parser ParserPlugin) {
	r.mu.Lock()
	defer r.mu.Unlock()

	lang := parser.Language()
	r.parsers[lang] = append(r.parsers[lang], parser)

	// Register by extensions
	for _, ext := range parser.Extensions() {
		r.extensions[ext] = append(r.extensions[ext], parser)
	}
}

// RegisterFrameworkAnalyzer registers a framework analyzer
func (r *Registry) RegisterFrameworkAnalyzer(analyzer FrameworkAnalyzer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	lang := analyzer.Language()
	r.analyzers[lang] = append(r.analyzers[lang], analyzer)
}

// GetParserForFile returns the best parser for a file
func (r *Registry) GetParserForFile(filePath string) (ParserPlugin, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	r.mu.RLock()
	parsers, ok := r.extensions[ext]
	r.mu.RUnlock()

	if !ok || len(parsers) == 0 {
		return nil, fmt.Errorf("no parser registered for extension: %s", ext)
	}

	// Return parser with highest priority
	bestParser := parsers[0]
	for _, p := range parsers {
		if p.Priority() > bestParser.Priority() {
			bestParser = p
		}
	}

	return bestParser, nil
}

// GetParserForLanguage returns a parser for a specific language
func (r *Registry) GetParserForLanguage(language string) (ParserPlugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	parsers, ok := r.parsers[language]
	if !ok || len(parsers) == 0 {
		return nil, fmt.Errorf("no parser registered for language: %s", language)
	}

	return parsers[0], nil
}

// GetFrameworkAnalyzers returns all framework analyzers for a language
func (r *Registry) GetFrameworkAnalyzers(language string) []FrameworkAnalyzer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.analyzers[language]
}

// ParseFile parses a file and applies framework analyzers
func (r *Registry) ParseFile(filePath string, content []byte) (*parsing.ParseResult, error) {
	parser, err := r.GetParserForFile(filePath)
	if err != nil {
		return nil, err
	}

	result, err := parser.Parse(content, filePath)
	if err != nil {
		return nil, err
	}

	// Apply framework analyzers
	analyzers := r.GetFrameworkAnalyzers(parser.Language())
	for _, analyzer := range analyzers {
		if analyzer.DetectFramework(content, filePath) {
			frameworkInfo, err := analyzer.Analyze(result, content)
			if err == nil && frameworkInfo != nil {
				// result.Frameworks = append(result.Frameworks, frameworkInfo) // parsing.ParseResult does not have Frameworks
			}
		}
	}

	return result, nil
}

// ListLanguages returns all registered languages
func (r *Registry) ListLanguages() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	languages := make([]string, 0, len(r.parsers))
	for lang := range r.parsers {
		languages = append(languages, lang)
	}
	return languages
}

// ListFrameworks returns all registered frameworks
func (r *Registry) ListFrameworks() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	frameworks := make(map[string]bool)
	for _, analyzers := range r.analyzers {
		for _, analyzer := range analyzers {
			frameworks[analyzer.Framework()] = true
		}
	}

	result := make([]string, 0, len(frameworks))
	for fw := range frameworks {
		result = append(result, fw)
	}
	return result
}