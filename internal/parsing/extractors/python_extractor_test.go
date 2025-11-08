package extractors

import (
	"testing"

	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
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

	if fileSymbols.Language != "python" {
		t.Errorf("expected language 'python', got '%s'", fileSymbols.Language)
	}

	if len(fileSymbols.Functions) != 1 { // my_function
		t.Errorf("expected 1 function, got %d", len(fileSymbols.Functions))
	}

	if len(fileSymbols.Classes) != 1 {
		t.Errorf("expected 1 class, got %d", len(fileSymbols.Classes))
	}

	if len(fileSymbols.Methods) != 2 { // __init__ and get_name
		t.Errorf("expected 2 methods, got %d", len(fileSymbols.Methods))
	}
}
