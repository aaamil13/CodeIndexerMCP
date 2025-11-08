package extractors

import (
	"testing"

	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

func TestGoExtractor_ExtractFunctions(t *testing.T) {
	sourceCode := `
package main

// A simple function
func hello() string {
    return "world"
}

/*
A multi-line comment
for a function with parameters
*/
func add(a, b int) int {
    return a + b
}

func (r *MyReceiver) aMethod(a, b int) int {
	return a + b
}
`
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	tree := parser.Parse(nil, []byte(sourceCode))

	parseResult := &parsing.ParseResult{
		Tree:       tree,
		Language:   "go",
		SourceCode: []byte(sourceCode),
		RootNode:   tree.RootNode(),
	}

	grammarManager := parsing.NewGrammarManager()
	queryEngine := parsing.NewQueryEngine(grammarManager)
	extractor := NewGoExtractor(queryEngine)

	functions, err := extractor.ExtractFunctions(parseResult, "test.go")
	if err != nil {
		t.Fatalf("ExtractFunctions failed: %v", err)
	}

	if len(functions) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(functions))
	}

	helloFunc := functions[0]
	if helloFunc.Name != "hello" {
		t.Errorf("expected function name 'hello', got '%s'", helloFunc.Name)
	}
	if helloFunc.ReturnType != "string" {
		t.Errorf("expected return type 'string', got '%s'", helloFunc.ReturnType)
	}

	addFunc := functions[1]
	if addFunc.Name != "add" {
		t.Errorf("expected function name 'add', got '%s'", addFunc.Name)
	}
	if addFunc.ReturnType != "int" {
		t.Errorf("expected return type 'int', got '%s'", addFunc.ReturnType)
	}
	if len(addFunc.Parameters) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(addFunc.Parameters))
	}
}

func TestGoExtractor_ExtractAll(t *testing.T) {
	sourceCode := `
package main

import (
	"fmt"
	"math"
)

// MyStruct is a test struct.
type MyStruct struct {
	// A public field
	PublicField int
	privateField string
}

// MyInterface is a test interface.
type MyInterface interface {
	DoSomething()
}

// NewMyStruct creates a new MyStruct.
func NewMyStruct() *MyStruct {
	return &MyStruct{}
}

// DoSomething is a method on MyStruct.
func (s *MyStruct) DoSomething() {
	fmt.Println("Doing something")
}

// privateFunc is a private function.
func privateFunc() {
	fmt.Println("Private function")
}
`
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())
	tree := parser.Parse(nil, []byte(sourceCode))

	parseResult := &parsing.ParseResult{
		Tree:       tree,
		Language:   "go",
		SourceCode: []byte(sourceCode),
		RootNode:   tree.RootNode(),
	}

	grammarManager := parsing.NewGrammarManager()
	queryEngine := parsing.NewQueryEngine(grammarManager)
	extractor := NewGoExtractor(queryEngine)

	fileSymbols, err := extractor.ExtractAll(parseResult, "test.go")
	if err != nil {
		t.Fatalf("ExtractAll failed: %v", err)
	}

	if fileSymbols.Language != "go" {
		t.Errorf("expected language 'go', got '%s'", fileSymbols.Language)
	}

	if len(fileSymbols.Functions) != 2 {
		t.Errorf("expected 2 functions, got %d", len(fileSymbols.Functions))
	}

	if len(fileSymbols.Methods) != 1 {
		t.Errorf("expected 1 method, got %d", len(fileSymbols.Methods))
	}

	if len(fileSymbols.Classes) != 1 {
		t.Errorf("expected 1 struct, got %d", len(fileSymbols.Classes))
	}

	// if len(fileSymbols.Interfaces) != 1 {
	// 	t.Errorf("expected 1 interface, got %d", len(fileSymbols.Interfaces))
	// }

	// if len(fileSymbols.Imports) != 2 {
	// 	t.Errorf("expected 2 imports, got %d", len(fileSymbols.Imports))
	// }
}
