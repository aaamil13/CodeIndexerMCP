package extractors

import (
    "fmt"
    "strings"
    "time"
    
    "github.com/aaamil13/CodeIndexerMCP/internal/model"
    "github.com/aaamil13/CodeIndexerMCP/internal/parsing"
    sitter "github.com/smacker/go-tree-sitter"
)

const GoFunctionQuery = `
(function_declaration
  name: (identifier) @func.name
  parameters: (parameter_list) @func.params
  result: (_)? @func.return
  body: (block) @func.body) @func.def
`

const GoMethodQuery = `
(method_declaration
  receiver: (parameter_list
    (parameter_declaration
      type: (_) @method.receiver_type)) @method.receiver
  name: (field_identifier) @method.name
  parameters: (parameter_list) @method.params
  result: (_)? @method.return
  body: (block) @method.body) @method.def
`

const GoStructQuery = `
(type_declaration
  (type_spec
    name: (type_identifier) @struct.name
    type: (struct_type
      (field_declaration_list) @struct.fields))) @struct.def
`

const GoInterfaceQuery = `
(type_declaration
  (type_spec
    name: (type_identifier) @interface.name
    type: (interface_type) @interface.body)) @interface.def
`

const GoImportQuery = `
(import_declaration
  (import_spec
    path: (interpreted_string_literal) @import.path
    name: (package_identifier)? @import.alias)) @import
`

type GoExtractor struct {
    BaseExtractor
    queryEngine *parsing.QueryEngine
}

func NewGoExtractor(qe *parsing.QueryEngine) *GoExtractor {
    return &GoExtractor{
        BaseExtractor: BaseExtractor{Language: "go"},
        queryEngine:   qe,
    }
}

func (ge *GoExtractor) ExtractFunctions(parseResult *parsing.ParseResult, filePath string) ([]*model.Function, error) {
    queryResult, err := ge.queryEngine.Execute(parseResult, GoFunctionQuery)
    if err != nil {
        return nil, err
    }
    
    functions := make([]*model.Function, 0)
    
    for _, match := range queryResult.Matches {
        var funcName, returnType, body string
        var funcNode, paramsNode *sitter.Node  // ПРОМЯНА: запазваме node вместо string
        
        for _, capture := range match.Captures {
            switch capture.Name {
            case "func.name":
                funcName = capture.Text
            case "func.params":
                paramsNode = capture.Node  // ПРОМЯНА: запазваме node
            case "func.return":
                returnType = capture.Text
            case "func.body":
                body = capture.Text
            case "func.def":
                funcNode = capture.Node
            }
        }
        
        if funcName == "" || funcNode == nil {
            continue
        }
        
        pos := ge.NodeToPosition(funcNode)
        funcRange := ge.NodeToRange(funcNode)
        
        // ПРОМЯНА: използваме parseParametersFromNode вместо parseParameters
        parameters := ge.parseParametersFromNode(paramsNode, parseResult.SourceCode)
        
        // 💡 ПОДОБРЕНИЕ #5: Изчисляване на content hash
        contentHash := ge.ComputeContentHash(body)
        
        function := &model.Function{
            Symbol: model.Symbol{
                ID:            ge.GenerateID("function", funcName, filePath, pos),
                Name:          funcName,
                Kind:          "function",
                File:          filePath,
                Range:         funcRange,
                Signature:     ge.buildSignature(funcName, parameters, returnType),
                Documentation: ge.ExtractDocumentation(funcNode, parseResult.SourceCode),
                Language:      "go",
                ContentHash:   contentHash,  // НОВО
                Status:        ge.ExtractStatusFromComments(funcNode, parseResult.SourceCode),
                Priority:      ge.ExtractPriorityFromComments(funcNode, parseResult.SourceCode),
                CreatedAt:     time.Now(),
                UpdatedAt:     time.Now(),
            },
            Parameters: parameters,
            ReturnType: strings.TrimSpace(returnType),
            Body:       body,
        }
        
        functions = append(functions, function)
    }
    
    return functions, nil
}

func (ge *GoExtractor) ExtractMethods(parseResult *parsing.ParseResult, filePath string) ([]*model.Method, error) {
    queryResult, err := ge.queryEngine.Execute(parseResult, GoMethodQuery)
    if err != nil {
        return nil, err
    }
    
    methods := make([]*model.Method, 0)
    
    for _, match := range queryResult.Matches {
        var methodName, receiverType, params, returnType, body string
        var methodNode *sitter.Node
        
        for _, capture := range match.Captures {
            switch capture.Name {
            case "method.name":
                methodName = capture.Text
            case "method.receiver_type":
                receiverType = capture.Text
            case "method.params":
                params = capture.Text
            case "method.return":
                returnType = capture.Text
            case "method.body":
                body = capture.Text
            case "method.def":
                methodNode = capture.Node
            }
        }
        
        if methodName == "" || methodNode == nil {
            continue
        }
        
        pos := ge.NodeToPosition(methodNode)
        methodRange := ge.NodeToRange(methodNode)
        
        method := &model.Method{
            Function: model.Function{
                Symbol: model.Symbol{
                    ID:            ge.GenerateID("method", methodName, filePath, pos),
                    Name:          methodName,
                    Kind:          "method",
                    File:          filePath,
                    Range:         methodRange,
                    Signature:     fmt.Sprintf("func (%s) %s%s %s", receiverType, methodName, params, returnType),
                    Documentation: ge.ExtractDocumentation(methodNode, parseResult.SourceCode),
                    Language:      "go",
                    Status:        ge.ExtractStatusFromComments(methodNode, parseResult.SourceCode),
                    Priority:      ge.ExtractPriorityFromComments(methodNode, parseResult.SourceCode),
                    CreatedAt:     time.Now(),
                    UpdatedAt:     time.Now(),
                },
                Parameters: ge.parseParameters(params),
                ReturnType: strings.TrimSpace(returnType),
                Body:       body,
            },
            ReceiverType: receiverType,
        }
        
        methods = append(methods, method)
    }
    
    return methods, nil
}

func (ge *GoExtractor) ExtractStructs(parseResult *parsing.ParseResult, filePath string) ([]*model.Class, error) {
    queryResult, err := ge.queryEngine.Execute(parseResult, GoStructQuery)
    if err != nil {
        return nil, err
    }
    
    structs := make([]*model.Class, 0)
    
    for _, match := range queryResult.Matches {
        var structName string
        var structNode *sitter.Node
        
        for _, capture := range match.Captures {
            switch capture.Name {
            case "struct.name":
                structName = capture.Text
            case "struct.def":
                structNode = capture.Node
            }
        }
        
        if structName == "" || structNode == nil {
            continue
        }
        
        pos := ge.NodeToPosition(structNode)
        structRange := ge.NodeToRange(structNode)
        
        class := &model.Class{
            Symbol: model.Symbol{
                ID:            ge.GenerateID("struct", structName, filePath, pos),
                Name:          structName,
                Kind:          "struct",
                File:          filePath,
                Range:         structRange,
                Signature:     fmt.Sprintf("type %s struct", structName),
                Documentation: ge.ExtractDocumentation(structNode, parseResult.SourceCode),
                Language:      "go",
                Status:        ge.ExtractStatusFromComments(structNode, parseResult.SourceCode),
                Priority:      ge.ExtractPriorityFromComments(structNode, parseResult.SourceCode),
                CreatedAt:     time.Now(),
                UpdatedAt:     time.Now(),
            },
            Methods: make([]model.Method, 0),
            Fields:  ge.extractStructFields(structNode, parseResult.SourceCode),
        }
        
        structs = append(structs, class)
    }
    
    return structs, nil
}

func (ge *GoExtractor) parseParameters(paramsStr string) []model.Parameter {
    // СТАР ПОДХОД: String parsing (нестабилен за сложни сигнатури)
    // ПРОБЛЕМ: "(ctx context.Context, options ...func(cfg *Config))" ще се счупи
    
    // НОВ ПОДХОД в подобрение #4 по-долу
    return []model.Parameter{}
}

// 💡 ПОДОБРЕНИЕ #4: Използване на Tree-sitter за Парсване на Параметри

func (ge *GoExtractor) parseParametersFromNode(paramsNode *sitter.Node, source []byte) []model.Parameter {
    params := make([]model.Parameter, 0)
    
    if paramsNode == nil || paramsNode.Type() != "parameter_list" {
        return params
    }
    
    // Обхождане на всички parameter_declaration nodes
    for i := 0; i < int(paramsNode.ChildCount()); i++ {
        child := paramsNode.Child(i)
        
        if child.Type() != "parameter_declaration" {
            continue
        }
        
        param := ge.extractParameter(child, source)
        if param != nil {
            params = append(params, *param)
        }
    }
    
    return params
}

func (ge *GoExtractor) extractParameter(paramNode *sitter.Node, source []byte) *model.Parameter {
    var name, paramType string
    var isVariadic bool
    
    // Обхождане на под-нодовете на параметъра
    for i := 0; i < int(paramNode.ChildCount()); i++ {
        child := paramNode.Child(i)
        
        switch child.Type() {
        case "identifier":
            // Име на параметър
            name = ge.ExtractText(child, source)
            
        case "type_identifier", "qualified_type", "pointer_type", 
             "array_type", "slice_type", "struct_type", "interface_type",
             "function_type", "map_type", "channel_type":
            // Тип на параметър
            paramType = ge.ExtractText(child, source)
            
        case "variadic_parameter_declaration":
            // Variadic параметър (...Type)
            isVariadic = true
            // Извличане на типа от variadic декларацията
            if child.ChildCount() > 0 {
                typeNode := child.Child(child.ChildCount() - 1)
                paramType = "..." + ge.ExtractText(typeNode, source)
            }
        }
    }
    
    // Ако няма име, но има тип, това е анонимен параметър
    if name == "" && paramType != "" {
        name = "_"
    }
    
    if paramType == "" {
        return nil
    }
    
    return &model.Parameter{
        Name:       name,
        Type:       paramType,
        IsVariadic: isVariadic,
    }
}

func (ge *GoExtractor) buildSignature(name string, params []model.Parameter, returnType string) string {
    paramStrs := make([]string, len(params))
    for i, p := range params {
        if p.Name == "_" {
            paramStrs[i] = p.Type
        } else {
            paramStrs[i] = fmt.Sprintf("%s %s", p.Name, p.Type)
        }
    }
    
    sig := fmt.Sprintf("func %s(%s)", name, strings.Join(paramStrs, ", "))
    if returnType != "" {
        sig += " " + returnType
    }
    return sig
}

func (ge *GoExtractor) extractStructFields(node *sitter.Node, source []byte) []model.Field {
    // Извличане на полетата на struct
    fields := make([]model.Field, 0)
    
    // Обходи child nodes и извлечи field_declaration
    // Simplified имплементация
    
    return fields
}

func (ge *GoExtractor) ExtractAll(parseResult *parsing.ParseResult, filePath string) (*model.FileSymbols, error) {
    functions, err := ge.ExtractFunctions(parseResult, filePath)
    if err != nil {
        return nil, err
    }
    
    methods, err := ge.ExtractMethods(parseResult, filePath)
    if err != nil {
        return nil, err
    }
    
    structs, err := ge.ExtractStructs(parseResult, filePath)
    if err != nil {
        return nil, err
    }
    
    return &model.FileSymbols{
        FilePath:  filePath,
        Language:  "go",
        Functions: functions,
        Methods:   methods,
        Classes:   structs,
        ParseTime: time.Now(),
    }, nil
}
