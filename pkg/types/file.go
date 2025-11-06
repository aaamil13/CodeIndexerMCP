package types

import "time"

// File represents an indexed file
type File struct {
	ID           int64     `json:"id"`
	ProjectID    int64     `json:"project_id"`
	Path         string    `json:"path"`
	RelativePath string    `json:"relative_path"`
	Language     string    `json:"language"`
	Size         int64     `json:"size"`
	LinesOfCode  int       `json:"lines_of_code"`
	Hash         string    `json:"hash"`
	LastModified time.Time `json:"last_modified"`
	LastIndexed  time.Time `json:"last_indexed"`
}

// ImportType represents the type of import
type ImportType string

const (
	ImportTypeLocal    ImportType = "local"
	ImportTypeExternal ImportType = "external"
	ImportTypeStdlib   ImportType = "stdlib"
)

// Import represents an import statement
type Import struct {
	ID             int64      `json:"id"`
	FileID         int64      `json:"file_id"`
	Source         string     `json:"source"`
	ImportedNames  []string   `json:"imported_names,omitempty"`
	ImportType     ImportType `json:"import_type"`
	LineNumber     int        `json:"line_number"`
	ImportedSymbol string     `json:"imported_symbol,omitempty"` // For specific symbol imports
}
