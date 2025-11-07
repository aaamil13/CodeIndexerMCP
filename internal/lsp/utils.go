package lsp

import (
	"fmt"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// symbolTypeToSymbolKind converts internal symbol type to LSP SymbolKind
func symbolTypeToSymbolKind(symbolType types.SymbolType) SymbolKind {
	switch symbolType {
	case types.SymbolTypeFunction:
		return SymbolKindFunction
	case types.SymbolTypeMethod:
		return SymbolKindMethod
	case types.SymbolTypeClass:
		return SymbolKindClass
	case types.SymbolTypeInterface:
		return SymbolKindInterface
	case types.SymbolTypeVariable:
		return SymbolKindVariable
	case types.SymbolTypeConstant:
		return SymbolKindConstant
	case types.SymbolTypeStruct:
		return SymbolKindStruct
	case types.SymbolTypeEnum:
		return SymbolKindEnum
	case types.SymbolTypeConstructor:
		return SymbolKindConstructor
	case types.SymbolTypeField:
		return SymbolKindField
	case types.SymbolTypeProperty:
		return SymbolKindProperty
	case types.SymbolTypeModule:
		return SymbolKindModule
	case types.SymbolTypeNamespace:
		return SymbolKindNamespace
	case types.SymbolTypePackage:
		return SymbolKindPackage
	default:
		return SymbolKindVariable
	}
}

// symbolTypeToCompletionKind converts internal symbol type to LSP CompletionItemKind
func symbolTypeToCompletionKind(symbolType types.SymbolType) CompletionItemKind {
	switch symbolType {
	case types.SymbolTypeFunction:
		return CompletionItemKindFunction
	case types.SymbolTypeMethod:
		return CompletionItemKindMethod
	case types.SymbolTypeClass:
		return CompletionItemKindClass
	case types.SymbolTypeInterface:
		return CompletionItemKindInterface
	case types.SymbolTypeVariable:
		return CompletionItemKindVariable
	case types.SymbolTypeConstant:
		return CompletionItemKindConstant
	case types.SymbolTypeStruct:
		return CompletionItemKindStruct
	case types.SymbolTypeEnum:
		return CompletionItemKindEnum
	case types.SymbolTypeConstructor:
		return CompletionItemKindConstructor
	case types.SymbolTypeField:
		return CompletionItemKindField
	case types.SymbolTypeProperty:
		return CompletionItemKindProperty
	case types.SymbolTypeModule:
		return CompletionItemKindModule
	default:
		return CompletionItemKindText
	}
}

// severityToDiagnosticSeverity converts string severity to LSP DiagnosticSeverity
func severityToDiagnosticSeverity(severity string) DiagnosticSeverity {
	switch severity {
	case "error":
		return DiagnosticSeverityError
	case "warning":
		return DiagnosticSeverityWarning
	case "info", "information":
		return DiagnosticSeverityInformation
	case "hint":
		return DiagnosticSeverityHint
	default:
		return DiagnosticSeverityWarning
	}
}

// createDiagnosticFromUndefinedUsage creates an LSP diagnostic from undefined usage
func createDiagnosticFromUndefinedUsage(usage *types.UndefinedUsage) Diagnostic {
	return Diagnostic{
		Range: Range{
			Start: Position{Line: usage.Line - 1, Character: usage.Column},
			End:   Position{Line: usage.Line - 1, Character: usage.Column + len(usage.Name)},
		},
		Severity: severityToDiagnosticSeverity(usage.Severity),
		Source:   "codeindexer",
		Message:  fmt.Sprintf("Undefined %s: %s", usage.UsageType, usage.Name),
	}
}

// createDiagnosticFromTypeMismatch creates an LSP diagnostic from type mismatch
func createDiagnosticFromTypeMismatch(mismatch *types.TypeMismatch) Diagnostic {
	message := fmt.Sprintf("Type mismatch: expected %s, got %s",
		mismatch.ExpectedType, mismatch.ActualType)

	return Diagnostic{
		Range: Range{
			Start: Position{Line: mismatch.Line - 1, Character: mismatch.Column},
			End:   Position{Line: mismatch.Line - 1, Character: mismatch.Column + 1},
		},
		Severity: DiagnosticSeverityError,
		Source:   "codeindexer",
		Message:  message,
	}
}

// createDiagnosticFromMissingMethod creates an LSP diagnostic from missing method
func createDiagnosticFromMissingMethod(missing *types.MissingMethod) Diagnostic {
	message := fmt.Sprintf("Method '%s' not found on type '%s'",
		missing.MethodName, missing.TypeName)

	if len(missing.AvailableMethods) > 0 {
		message += fmt.Sprintf(". Available methods: %v", missing.AvailableMethods)
	}

	return Diagnostic{
		Range: Range{
			Start: Position{Line: missing.Line - 1, Character: missing.Column},
			End:   Position{Line: missing.Line - 1, Character: missing.Column + len(missing.MethodName)},
		},
		Severity: DiagnosticSeverityError,
		Source:   "codeindexer",
		Message:  message,
	}
}

