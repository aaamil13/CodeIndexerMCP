package parsing_test

import (
	"testing"

	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
	"github.com/stretchr/testify/assert"
)

func TestASTProvider_ParseGo(t *testing.T) {
	gm := parsing.NewGrammarManager()
	ap := parsing.NewASTProvider(gm)

	sourceCode := []byte(`
package main

func hello() string {
    return "world"
}
`)

	parseResult, err := ap.Parse("go", sourceCode)
	assert.NoError(t, err)
	assert.NotNil(t, parseResult)
	defer parseResult.Close()

	assert.Equal(t, "go", parseResult.Language)
	assert.NotNil(t, parseResult.Tree)
	assert.NotNil(t, parseResult.RootNode)
	assert.Equal(t, "source_file", parseResult.RootNode.Type())
	assert.Empty(t, parseResult.ParseErrors)
}

func TestASTProvider_ParsePython(t *testing.T) {
	gm := parsing.NewGrammarManager()
	ap := parsing.NewASTProvider(gm)

	sourceCode := []byte(`
def greet(name):
    print(f"Hello, {name}")
`)

	parseResult, err := ap.Parse("python", sourceCode)
	assert.NoError(t, err)
	assert.NotNil(t, parseResult)
	defer parseResult.Close()

	assert.Equal(t, "python", parseResult.Language)
	assert.NotNil(t, parseResult.Tree)
	assert.NotNil(t, parseResult.RootNode)
	assert.Equal(t, "module", parseResult.RootNode.Type())
	assert.Empty(t, parseResult.ParseErrors)
}

func TestASTProvider_ParseError(t *testing.T) {
	gm := parsing.NewGrammarManager()
	ap := parsing.NewASTProvider(gm)

	// Malformed Go code
	sourceCode := []byte(`
package main

func hello() string {
    return "world"
`) // Missing closing brace

	parseResult, err := ap.Parse("go", sourceCode)
	assert.NoError(t, err) // Tree-sitter still produces a tree, but with errors
	assert.NotNil(t, parseResult)
	defer parseResult.Close()

	assert.Equal(t, "go", parseResult.Language)
	assert.NotNil(t, parseResult.Tree)
	assert.NotNil(t, parseResult.RootNode)
	assert.True(t, parseResult.RootNode.HasError())
	assert.NotEmpty(t, parseResult.ParseErrors)
	assert.Contains(t, parseResult.ParseErrors[0].Message, "Syntax error")
}

func TestASTProvider_UnsupportedLanguage(t *testing.T) {
	gm := parsing.NewGrammarManager()
	ap := parsing.NewASTProvider(gm)

	sourceCode := []byte(`some unsupported code`)

	parseResult, err := ap.Parse("unsupported_lang", sourceCode)
	assert.Error(t, err)
	assert.Nil(t, parseResult)
	assert.Contains(t, err.Error(), "language not supported")
}
