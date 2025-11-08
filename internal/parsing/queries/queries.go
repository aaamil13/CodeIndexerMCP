package queries

import (
	_ "embed"
	"fmt"
)

// Вградени Tree-sitter заявки (компилирани в бинарния файл)

//go:embed go.scm
var Go string

//go:embed python.scm
var Python string

//go:embed typescript.scm
var TypeScript string

//go:embed javascript.scm
var JavaScript string

//go:embed java.scm
var Java string

//go:embed c.scm
var C string

//go:embed cpp.scm
var Cpp string

//go:embed rust.scm
var Rust string

// Добавете всички езици...

// GetQuery връща вградената заявка за език
func GetQuery(language, queryName string) (string, error) {
    // Map на всички заявки
    allQueries := map[string]string{
        "go":         Go,
        "python":     Python,
        "typescript": TypeScript,
        "javascript": JavaScript,
        "java":       Java,
        "c":          C,
        "cpp":        Cpp,
        "rust":       Rust,
    }
    
    query, exists := allQueries[language]
    if !exists {
        return "", fmt.Errorf("no query file for language: %s", language)
    }
    
    return query, nil
}
