package types

// Parser is the interface that all language parsers must implement
type Parser interface {
	// Language returns the language name (e.g., "go", "python", "typescript")
	Language() string

	// Extensions returns supported file extensions (e.g., [".go", ".mod"])
	Extensions() []string

	// Parse parses the file content and returns structured information
	Parse(content []byte, filePath string) (*ParseResult, error)

	// CanParse checks if this parser can handle the given file
	CanParse(filePath string) bool
}

// ParseResult contains all information extracted from parsing a file
type ParseResult struct {
	Symbols       []*Symbol              `json:"symbols"`
	Imports       []*Import              `json:"imports"`
	Relationships []*Relationship        `json:"relationships"`
	Frameworks    []*FrameworkInfo       `json:"frameworks,omitempty"`    // Framework-specific info
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Errors        []ParseError           `json:"errors,omitempty"`
}

// ParseError represents an error that occurred during parsing
type ParseError struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
}
