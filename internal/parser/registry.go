package parser

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// Registry manages language parsers
type Registry struct {
	parsers map[string]types.Parser // language -> parser
	extMap  map[string]string       // extension -> language
	mu      sync.RWMutex
}

// NewRegistry creates a new parser registry
func NewRegistry() *Registry {
	return &Registry{
		parsers: make(map[string]types.Parser),
		extMap:  make(map[string]string),
	}
}

// Register registers a parser for a language
func (r *Registry) Register(parser types.Parser) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	lang := parser.Language()
	if _, exists := r.parsers[lang]; exists {
		return fmt.Errorf("parser for language %s already registered", lang)
	}

	r.parsers[lang] = parser

	// Map extensions to language
	for _, ext := range parser.Extensions() {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		r.extMap[ext] = lang
	}

	return nil
}

// GetParser retrieves a parser for a language
func (r *Registry) GetParser(language string) (types.Parser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	parser, ok := r.parsers[language]
	if !ok {
		return nil, fmt.Errorf("no parser found for language: %s", language)
	}

	return parser, nil
}

// GetParserForFile retrieves a parser based on file extension
func (r *Registry) GetParserForFile(filePath string) (types.Parser, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	r.mu.RLock()
	lang, ok := r.extMap[ext]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no parser found for extension: %s", ext)
	}

	return r.GetParser(lang)
}

// CanParse checks if a file can be parsed
func (r *Registry) CanParse(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))

	r.mu.RLock()
	_, ok := r.extMap[ext]
	r.mu.RUnlock()

	return ok
}

// SupportedLanguages returns a list of supported languages
func (r *Registry) SupportedLanguages() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	langs := make([]string, 0, len(r.parsers))
	for lang := range r.parsers {
		langs = append(langs, lang)
	}

	return langs
}

// SupportedExtensions returns a list of supported file extensions
func (r *Registry) SupportedExtensions() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	exts := make([]string, 0, len(r.extMap))
	for ext := range r.extMap {
		exts = append(exts, ext)
	}

	return exts
}
