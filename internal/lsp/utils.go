package lsp

import (
	"fmt"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

// symbolTypeToSymbolKind converts internal symbol type to LSP SymbolKind
func symbolTypeToSymbolKind(symbolType model.SymbolType) SymbolKind {
	switch symbolType {
	case model.SymbolTypeFunction:
		return SymbolKindFunction
	case model.SymbolTypeMethod:
		return SymbolKindMethod
	case model.SymbolTypeClass:
		return SymbolKindClass
	case model.SymbolTypeInterface:
		return SymbolKindInterface
	case model.SymbolTypeVariable:
		return SymbolKindVariable
	case model.SymbolTypeConstant:
		return SymbolKindConstant
	case model.SymbolTypeStruct:
		return SymbolKindStruct
	case model.SymbolTypeEnum:
		return SymbolKindEnum
	case model.SymbolTypeConstructor:
		return SymbolKindConstructor
	case model.SymbolTypeField:
		return SymbolKindField
	case model.SymbolTypeProperty:
		return SymbolKindProperty
	case model.SymbolTypeModule:
		return SymbolKindModule
	case model.SymbolTypeNamespace:
		return SymbolKindNamespace
	case model.SymbolTypePackage:
		return SymbolKindPackage
	default:
		return SymbolKindVariable
	}
}

// symbolTypeToCompletionKind converts internal symbol type to LSP CompletionItemKind
func symbolTypeToCompletionKind(symbolType model.SymbolType) CompletionItemKind {
	switch symbolType {
	case model.SymbolTypeFunction:
		return CompletionItemKindFunction
	case model.SymbolTypeMethod:
		return CompletionItemKindMethod
	case model.SymbolTypeClass:
		return CompletionItemKindClass
	case model.SymbolTypeInterface:
		return CompletionItemKindInterface
	case model.SymbolTypeVariable:
		return CompletionItemKindVariable
	case model.SymbolTypeConstant:
		return CompletionItemKindConstant
	case model.SymbolTypeStruct:
		return CompletionItemKindStruct
	case model.SymbolTypeEnum:
		return CompletionItemKindEnum
	case model.SymbolTypeConstructor:
		return CompletionItemKindConstructor
	case model.SymbolTypeField:
		return CompletionItemKindField
	case model.SymbolTypeProperty:
		return CompletionItemKindProperty
	case model.SymbolTypeModule:
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
func createDiagnosticFromUndefinedUsage(usage *model.UndefinedUsage) Diagnostic {
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
func createDiagnosticFromTypeMismatch(mismatch *model.TypeMismatch) Diagnostic {
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
func createDiagnosticFromMissingMethod(missing *model.MissingMethod) Diagnostic {
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

