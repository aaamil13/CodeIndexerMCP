package extractors

import (
	"testing"
	"encoding/json" // Added for JSON unmarshalling

	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
	"github.com/aaamil13/CodeIndexerMCP/internal/model" // Explicitly import model
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

func TestPythonExtractor_ExtractClasses(t *testing.T) {
	sourceCode := `
class MyClass:
    pass

class AnotherClass(MyClass):
    def __init__(self):
        pass
`
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree := parser.Parse(nil, []byte(sourceCode))

	parseResult := &parsing.ParseResult{
		Tree:       tree,
		Language:   "python",
		SourceCode: []byte(sourceCode),
		RootNode:   tree.RootNode(),
	}

	grammarManager := parsing.NewGrammarManager()
	queryEngine := parsing.NewQueryEngine(grammarManager)
	extractor := NewPythonExtractor(queryEngine)

	classes, err := extractor.ExtractClasses(parseResult, "test.py")
	if err != nil {
		t.Fatalf("ExtractClasses failed: %v", err)
	}

	if len(classes) != 2 {
		t.Fatalf("expected 2 classes, got %d", len(classes))
	}

	myClass := classes[0]
	if myClass.Name != "MyClass" {
		t.Errorf("expected class name 'MyClass', got '%s'", myClass.Name)
	}

	anotherClass := classes[1]
	if anotherClass.Name != "AnotherClass" {
		t.Errorf("expected class name 'AnotherClass', got '%s'", anotherClass.Name)
	}
	if len(anotherClass.BaseClasses) != 1 {
		t.Errorf("expected 1 base class, got %d", len(anotherClass.BaseClasses))
	}
	if anotherClass.BaseClasses[0] != "MyClass" {
		t.Errorf("expected base class 'MyClass', got '%s'", anotherClass.BaseClasses[0])
	}
}

func TestPythonExtractor_ExtractAll(t *testing.T) {
	sourceCode := `
import os
import sys

def my_function(arg1, arg2):
    """Docstring for my_function."""
    return arg1 + arg2

class MyClass(object):
    def __init__(self, name):
        self.name = name

    def get_name(self):
        return self.name
`
	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())
	tree := parser.Parse(nil, []byte(sourceCode))

	parseResult := &parsing.ParseResult{
		Tree:       tree,
		Language:   "python",
		SourceCode: []byte(sourceCode),
		RootNode:   tree.RootNode(),
	}

	grammarManager := parsing.NewGrammarManager()
	queryEngine := parsing.NewQueryEngine(grammarManager)
	extractor := NewPythonExtractor(queryEngine)

	fileSymbols, err := extractor.ExtractAll(parseResult, "test.py")
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
				if err := json.Unmarshal([]byte(metaStr), &f); err != nil {
					t.Fatalf("failed to unmarshal function metadata: %v", err)
				}
			}
			f.Symbol = *sym // Copy common symbol fields
			functions = append(functions, &f)
		case "method":
			var m model.Method
			if metaStr, ok := sym.Metadata["method"]; ok {
				if err := json.Unmarshal([]byte(metaStr), &m); err != nil {
					t.Fatalf("failed to unmarshal method metadata: %v", err)
				}
			}
			m.Symbol = *sym // Copy common symbol fields
			methods = append(methods, &m)
		case "class":
			var c model.Class
			if metaStr, ok := sym.Metadata["class"]; ok {
				if err := json.Unmarshal([]byte(metaStr), &c); err != nil {
					t.Fatalf("failed to unmarshal class metadata: %v", err)
				}
			}
			c.Symbol = *sym // Copy common symbol fields
			classes = append(classes, &c)
		}
	}

	if len(functions) != 1 { // my_function
		t.Errorf("expected 1 function, got %d", len(functions))
	}

	if len(classes) != 1 {
		t.Errorf("expected 1 class, got %d", len(classes))
	}

	if len(methods) != 2 { // __init__ and get_name
		t.Errorf("expected 2 methods, got %d", len(methods))
	}
}
