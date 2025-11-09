package extractors

import (
	"testing"
	"encoding/json" // Added for JSON unmarshalling

	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
	"github.com/aaamil13/CodeIndexerMCP/internal/model" // Explicitly import model
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

	// Unmarshal the SymbolsJSON to verify content
	var extractedSymbols []*model.Symbol
	err = json.Unmarshal(fileSymbols.SymbolsJSON, &extractedSymbols)
	if err != nil {
		t.Fatalf("failed to unmarshal symbolsJSON: %v", err)
	}

	var functions []*model.Function
	var methods []*model.Method
	var classes []*model.Class

	for _, sym := range extractedSymbols {
		switch sym.Kind {
		case "function":
			var f model.Function
			if metaStr, ok := sym.Metadata["function"]; ok {
				if err := json.Unmarshal([]byte(metaStr), &f); err != nil { // Assuming Metadata["function"] stores marshaled function data
					t.Fatalf("failed to unmarshal function metadata: %v", err)
				}
			}
			f.Symbol = *sym // Copy common symbol fields
			functions = append(functions, &f)
		case "method":
			var m model.Method
			if metaStr, ok := sym.Metadata["method"]; ok {
				if err := json.Unmarshal([]byte(metaStr), &m); err != nil { // Assuming Metadata["method"] stores marshaled method data
					t.Fatalf("failed to unmarshal method metadata: %v", err)
				}
			}
			m.Symbol = *sym // Copy common symbol fields
			methods = append(methods, &m)
		case "struct": // Classes are represented as structs in Go
			var c model.Class
			if metaStr, ok := sym.Metadata["class"]; ok {
				if err := json.Unmarshal([]byte(metaStr), &c); err != nil { // Assuming Metadata["class"] stores marshaled class data
					t.Fatalf("failed to unmarshal class metadata: %v", err)
				}
			}
			c.Symbol = *sym // Copy common symbol fields
			classes = append(classes, &c)
		}
	}

	if len(functions) != 2 {
		t.Errorf("expected 2 functions, got %d", len(functions))
	}

	if len(methods) != 1 {
		t.Errorf("expected 1 method, got %d", len(methods))
	}

	if len(classes) != 1 {
		t.Errorf("expected 1 struct (class), got %d", len(classes))
	}
}
