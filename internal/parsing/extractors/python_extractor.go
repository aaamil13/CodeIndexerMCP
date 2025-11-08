package extractors

import (
	"fmt"
	"strings"
	"time"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
	sitter "github.com/smacker/go-tree-sitter"
)

const PythonFunctionQuery = `
(function_definition
  name: (identifier) @func.name
  parameters: (parameters) @func.params
  body: (block) @func.body) @func.def
`

const PythonClassQuery = `
(class_definition
  name: (identifier) @class.name
  superclasses: (argument_list)? @class.bases
  body: (block) @class.body) @class.def
`

const PythonMethodQuery = `
(class_definition
  body: (block
    (function_definition
      name: (identifier) @method.name
      parameters: (parameters) @method.params
      body: (block) @method.body))) @method.def
`

type PythonExtractor struct {
    BaseExtractor
    queryEngine *parsing.QueryEngine
}

func NewPythonExtractor(qe *parsing.QueryEngine) *PythonExtractor {
    return &PythonExtractor{
        BaseExtractor: BaseExtractor{Language: "python"},
        queryEngine:   qe,
    }
}

func (pe *PythonExtractor) ExtractFunctions(parseResult *parsing.ParseResult, filePath string) ([]*model.Function, error) {
	queryResult, err := pe.queryEngine.Execute(parseResult, PythonFunctionQuery)
	if err != nil {
		return nil, err
	}

	functions := make([]*model.Function, 0)

	for _, match := range queryResult.Matches {
		var funcName, body string
		var funcNode, paramsNode *sitter.Node

		for _, capture := range match.Captures {
			switch capture.Name {
			case "func.name":
				funcName = capture.Text
			case "func.params":
				paramsNode = capture.Node
			case "func.body":
				body = capture.Text
			case "func.def":
				funcNode = capture.Node
			}
		}

		if funcName == "" || funcNode == nil {
			continue
		}

		// Filter out methods (functions that have a class_definition as an ancestor)
		if funcNode.Parent() != nil && funcNode.Parent().Type() == "block" && funcNode.Parent().Parent() != nil && funcNode.Parent().Parent().Type() == "class_definition" {
			continue
		}

		pos := pe.NodeToPosition(funcNode)
		funcRange := pe.NodeToRange(funcNode)

		parameters := pe.parseParametersFromNode(paramsNode, parseResult.SourceCode)

		contentHash := pe.ComputeContentHash(body)

		function := &model.Function{
			Symbol: model.Symbol{
				ID:            pe.GenerateID("function", funcName, filePath, pos),
				Name:          funcName,
				Kind:          "function",
				File:          filePath,
				Range:         funcRange,
				Signature:     pe.buildSignature(funcName, parameters, ""),
				Documentation: pe.ExtractDocumentation(funcNode, parseResult.SourceCode),
				Language:      "python",
				ContentHash:   contentHash,
				Status:        pe.ExtractStatusFromComments(funcNode, parseResult.SourceCode),
				Priority:      pe.ExtractPriorityFromComments(funcNode, parseResult.SourceCode),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			Parameters: parameters,
			Body:       body,
		}

		functions = append(functions, function)
	}

	return functions, nil
}

func (pe *PythonExtractor) buildSignature(name string, params []model.Parameter, returnType string) string {
	paramStrs := make([]string, len(params))
	for i, p := range params {
		paramStrs[i] = p.Name
	}

	sig := fmt.Sprintf("def %s(%s)", name, strings.Join(paramStrs, ", "))
	if returnType != "" {
		sig += " -> " + returnType
	}
	return sig
}

func (pe *PythonExtractor) parseParametersFromNode(paramsNode *sitter.Node, source []byte) []model.Parameter {
	// TODO: Implement this function
	return []model.Parameter{}
}

func (pe *PythonExtractor) ExtractClasses(parseResult *parsing.ParseResult, filePath string) ([]*model.Class, error) {

	queryResult, err := pe.queryEngine.Execute(parseResult, PythonClassQuery)

	if err != nil {

		return nil, err

	}



	classes := make([]*model.Class, 0)



	for _, match := range queryResult.Matches {

		var className, bases string

		var classNode *sitter.Node



		for _, capture := range match.Captures {

			switch capture.Name {

			case "class.name":

				className = capture.Text

			case "class.bases":

				bases = capture.Text

			case "class.def":

				classNode = capture.Node

			}

		}



		if className == "" || classNode == nil {

			continue

		}



		pos := pe.NodeToPosition(classNode)

		classRange := pe.NodeToRange(classNode)



		if bases != "" {

			bases = strings.TrimPrefix(bases, "(")

			bases = strings.TrimSuffix(bases, ")")

		}



		class := &model.Class{

			Symbol: model.Symbol{

				ID:            pe.GenerateID("class", className, filePath, pos),

				Name:          className,

				Kind:          "class",

				File:          filePath,

				Range:         classRange,

				Signature:     fmt.Sprintf("class %s(%s)", className, bases),

				Documentation: pe.ExtractDocumentation(classNode, parseResult.SourceCode),

				Language:      "python",

				Status:        pe.ExtractStatusFromComments(classNode, parseResult.SourceCode),

				Priority:      pe.ExtractPriorityFromComments(classNode, parseResult.SourceCode),

				CreatedAt:     time.Now(),

				UpdatedAt:     time.Now(),

			},

			BaseClasses: strings.Split(bases, ", "),

		}



		classes = append(classes, class)

	}



	return classes, nil

}



func (pe *PythonExtractor) ExtractMethods(parseResult *parsing.ParseResult, filePath string) ([]*model.Method, error) {

	queryResult, err := pe.queryEngine.Execute(parseResult, PythonMethodQuery)

	if err != nil {

		return nil, err

	}



	methods := make([]*model.Method, 0)



	for _, match := range queryResult.Matches {

		var methodName, body string

		var methodNode, paramsNode *sitter.Node



		for _, capture := range match.Captures {

			switch capture.Name {

			case "method.name":

				methodName = capture.Text

			case "method.params":

				paramsNode = capture.Node

			case "method.body":

				body = capture.Text

			case "method.def":

				methodNode = capture.Node

			}

		}



		if methodName == "" || methodNode == nil {

			continue

		}



		pos := pe.NodeToPosition(methodNode)

		methodRange := pe.NodeToRange(methodNode)



		parameters := pe.parseParametersFromNode(paramsNode, parseResult.SourceCode)



		contentHash := pe.ComputeContentHash(body)



		method := &model.Method{

			Function: model.Function{

				Symbol: model.Symbol{

					ID:            pe.GenerateID("method", methodName, filePath, pos),

					Name:          methodName,

					Kind:          "method",

					File:          filePath,

					Range:         methodRange,

					Signature:     pe.buildSignature(methodName, parameters, ""),

					Documentation: pe.ExtractDocumentation(methodNode, parseResult.SourceCode),

					Language:      "python",

					ContentHash:   contentHash,

					Status:        pe.ExtractStatusFromComments(methodNode, parseResult.SourceCode),

					Priority:      pe.ExtractPriorityFromComments(methodNode, parseResult.SourceCode),

					CreatedAt:     time.Now(),

					UpdatedAt:     time.Now(),

				},

				Parameters: parameters,

				Body:       body,

			},

			ReceiverType: "", // Python methods don't have a distinct receiver type in signature

		}



		methods = append(methods, method)

	}



	return methods, nil

}



func (pe *PythonExtractor) ExtractAll(parseResult *parsing.ParseResult, filePath string) (*model.FileSymbols, error) {

	functions, err := pe.ExtractFunctions(parseResult, filePath)

	if err != nil {

		return nil, err

	}



	classes, err := pe.ExtractClasses(parseResult, filePath)

	if err != nil {

		return nil, err

	}



	methods, err := pe.ExtractMethods(parseResult, filePath)

	if err != nil {

		return nil, err

	}



	return &model.FileSymbols{

		FilePath:  filePath,

		Language:  "python",

		Functions: functions,

		Classes:   classes,

		Methods:   methods,

		ParseTime: time.Now(),

	},

	nil

}
