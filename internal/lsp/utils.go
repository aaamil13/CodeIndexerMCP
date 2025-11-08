package lsp

import (
	"fmt"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
)

// symbolTypeToSymbolKind converts internal symbol type to LSP SymbolKind
func symbolTypeToSymbolKind(symbolKind model.SymbolKind) SymbolKind {
	switch symbolKind {
	case model.SymbolKindFunction:
		return SymbolKindFunction
	case model.SymbolKindMethod:
		return SymbolKindMethod
	case model.SymbolKindClass:
		return SymbolKindClass
	case model.SymbolKindInterface:
		return SymbolKindInterface
	case model.SymbolKindVariable:
		return SymbolKindVariable
	case model.SymbolKindConstant:
		return SymbolKindConstant
	case model.SymbolKindStruct:
		return SymbolKindStruct
	case model.SymbolKindEnum:
		return SymbolKindEnum
	case model.SymbolKindConstructor:
		return SymbolKindConstructor
	case model.SymbolKindField:
		return SymbolKindField
	case model.SymbolKindProperty:
		return SymbolKindProperty
	case model.SymbolKindModule:
		return SymbolKindModule
	case model.SymbolKindNamespace:
		return SymbolKindNamespace
	case model.SymbolKindPackage:
		return SymbolKindPackage
	default:
		return SymbolKindVariable
	}
}

// symbolTypeToCompletionKind converts internal symbol type to LSP CompletionItemKind
func symbolTypeToCompletionKind(symbolKind model.SymbolKind) CompletionItemKind {
	switch symbolKind {
	case model.SymbolKindFunction:
		return CompletionItemKindFunction
	case model.SymbolKindMethod:
		return CompletionItemKindMethod
	case model.SymbolKindClass:
		return CompletionItemKindClass
	case model.SymbolKindInterface:
		return CompletionItemKindInterface
	case model.SymbolKindVariable:
		return CompletionItemKindVariable
	case model.SymbolKindConstant:
		return CompletionItemKindConstant
	case model.SymbolKindStruct:
		return CompletionItemKindStruct
	case model.SymbolKindEnum:
		return CompletionItemKindEnum
	case model.SymbolKindConstructor:
		return CompletionItemKindConstructor
	case model.SymbolKindField:
		return CompletionItemKindField
	case model.SymbolKindProperty:
		return CompletionItemKindProperty
	case model.SymbolKindModule:
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
			End:   Position{Line: usage.Line - 1, Character: usage.Column + len(usage.SymbolName)},
		},
		Severity: severityToDiagnosticSeverity(usage.Severity),
		Source:   "codeindexer",
		Message:  fmt.Sprintf("Undefined %s: %s", usage.UsageType, usage.SymbolName),
	}
}

// createDiagnosticFromTypeMismatch creates an LSP diagnostic from type mismatch
func createDiagnosticFromTypeMismatch(mismatch *model.TypeMismatch) Diagnostic {
	message := fmt.Sprintf("Type mismatch: expected %s, got %s",
		mismatch.Expected, mismatch.Found)

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

