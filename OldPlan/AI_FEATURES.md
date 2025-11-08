# AI-Powered Features 🧠

Code Indexer MCP now includes advanced AI-powered features designed to help AI agents understand, analyze, and refactor code more effectively.

## 🎯 Overview

These features dramatically improve AI agent's ability to:
- ✅ **Understand context quickly** - Get comprehensive code context with examples
- ✅ **Refactor safely** - Analyze impact before making changes
- ✅ **Assess quality** - Get code metrics and quality ratings
- ✅ **Navigate efficiently** - Extract smart snippets with dependencies
- ✅ **Detect issues** - Find unused code, complexity problems
- ✅ **Suggest improvements** - AI-powered refactoring opportunities

## 🛠️ AI-Powered MCP Tools

### 1. get_code_context

Get comprehensive context for a symbol including usage examples, dependencies, and relationships.

**Input:**
```json
{
  "symbol_name": "MyFunction",
  "depth": 5
}
```

**Output:**
- Symbol definition with signature
- File information
- Actual source code
- All dependencies (imports)
- Related symbols in the same file
- Callers (who calls this)
- Callees (what this calls)
- 5 real usage examples from codebase
- Full documentation
- Surrounding context with line markers

**Use Cases:**
- Understanding how a function works
- Learning API usage patterns
- Preparing for refactoring
- Generating documentation

### 2. analyze_change_impact

Analyze the impact of changing or refactoring a symbol.

**Input:**
```json
{
  "symbol_name": "calculateTotal"
}
```

**Output:**
- **Risk Level**: low, medium, or high
- **Direct References**: Count of direct usages
- **Indirect References**: Transitive usages
- **Affected Files**: List of files that would be impacted
- **Affected Symbols**: Symbols that reference this
- **Breaking Changes**: Whether this breaks public API
- **Suggestions**: Refactoring recommendations

**Risk Calculation:**
- High: >50 references OR >20 files OR exported with >20 refs
- Medium: >10 references OR >5 files OR exported with >5 refs
- Low: Everything else

**Use Cases:**
- Pre-refactoring risk assessment
- Understanding code dependencies
- Planning API changes
- Avoiding breaking changes

### 3. get_code_metrics

Calculate comprehensive code quality metrics.

**Input:**
```json
{
  "symbol_name": "processData"
}
```

**Output:**
- **Cyclomatic Complexity**: Decision point count
- **Cognitive Complexity**: Nesting-weighted complexity
- **Maintainability Index**: 0-100 score
- **Lines of Code**: Function length
- **Parameters**: Parameter count
- **Return Statements**: Multiple return detection
- **Max Nesting Depth**: Deepest nesting level
- **Comment Density**: Percentage of comment lines
- **Has Documentation**: Whether documented
- **Quality Rating**: excellent, good, fair, or poor

**Quality Criteria:**
- **Excellent** (6-7 points):
  - Low complexity (cyclomatic ≤10, cognitive ≤15)
  - High maintainability (≥80)
  - Good documentation
- **Good** (4-5 points): Moderate complexity, decent maintainability
- **Fair** (2-3 points): Higher complexity or lower maintainability
- **Poor** (0-1 points): High complexity, poor maintainability

**Use Cases:**
- Code review assistance
- Refactoring prioritization
- Quality gate enforcement
- Technical debt tracking

### 4. extract_smart_snippet

Extract a self-contained code snippet with all dependencies.

**Input:**
```json
{
  "symbol_name": "UserService"
}
```

**Output:**
- **Symbol**: The target symbol
- **Code**: Main source code
- **Dependencies**: Required imports
- **Related Code**: Helper functions, types used
- **Documentation**: Full docstring
- **Usage Hints**: How to use this code
- **Complete**: Whether it's runnable as-is

**Usage Hints Include:**
- Calling patterns
- Async/await requirements
- Import instructions
- Type annotations
- Language-specific tips

**Use Cases:**
- Code reuse assistance
- Documentation generation
- Example code creation
- Learning code patterns

### 5. get_usage_statistics

Get detailed usage statistics and patterns for a symbol.

**Input:**
```json
{
  "symbol_name": "validateEmail"
}
```

**Output:**
- **Usage Count**: Total times used
- **File Count**: Number of files using it
- **Usage By File**: Per-file breakdown
- **Common Patterns**: Detected usage patterns
- **Is Deprecated**: Whether marked deprecated
- **Alternatives**: Alternative symbols (if deprecated)

**Detected Patterns:**
- "Called X times"
- "Assigned X times"
- "Used as type X times"
- "Heavily used across codebase" (>50 uses)
- "Heavily used in one file"
- "Widely used across N files"

**Use Cases:**
- Understanding API popularity
- Deprecation planning
- Finding alternatives
- Usage pattern detection

### 6. suggest_refactorings

Get AI-powered refactoring suggestions based on analysis.

**Input:**
```json
{
  "symbol_name": "DataProcessor"
}
```

**Output Array of Opportunities:**

Each opportunity includes:
- **Type**: Refactoring type (increase_visibility, extract_interface, consolidate_usage)
- **Symbol**: Target symbol
- **Description**: What to refactor
- **Reason**: Why refactor
- **Impact**: Impact level (low/medium/high)
- **Effort**: Effort required (low/medium/high)
- **Benefits**: List of benefits
- **Risks**: Potential risks

**Suggestion Types:**

1. **increase_visibility**: Private symbol used extensively (>20 times)
2. **extract_interface**: Very high usage (>100 times) - hard to change
3. **consolidate_usage**: Usage spread across many files (>30) - coupling

**Use Cases:**
- Refactoring planning
- Technical debt reduction
- API design improvements
- Code organization

### 7. find_unused_symbols

Find unused/dead code in the project.

**Input:**
```json
{}
```

**Output:**
- **unused_symbols**: Array of unused symbols
- **count**: Total count
- **suggestion**: Removal/refactoring advice

**Detection Rules:**
- Only non-exported symbols (private/internal)
- Zero references in codebase
- Excludes exported APIs (might be used externally)

**Use Cases:**
- Dead code cleanup
- Reducing codebase size
- Improving maintainability
- Finding forgotten code

## 📊 Implementation Details

### Context Extractor

**Features:**
- Extracts code with configurable line ranges
- Provides surrounding context with markers (▶)
- Finds real usage examples from references
- Builds caller/callee relationships
- Handles errors gracefully (non-fatal)

**Example Context Output:**
```
  145 | func CalculateTotal(items []Item) float64 {
  146 |     var total float64
▶ 147 |     for _, item := range items {
▶ 148 |         total += item.Price * item.Quantity
▶ 149 |     }
▶ 150 |     return total
  151 | }
```

### Impact Analyzer

**Risk Assessment Algorithm:**
```go
if directRefs > 50 || affectedFiles > 20:
    risk = HIGH
elif isExported && (directRefs > 20 || affectedFiles > 10):
    risk = HIGH
elif isExported && (directRefs > 5 || affectedFiles > 3):
    risk = MEDIUM
elif directRefs > 10 || affectedFiles > 5:
    risk = MEDIUM
else:
    risk = LOW
```

**Suggestions Generated:**
- Deprecation period for high-usage symbols
- Migration guides for exported APIs
- Wrapper functions for compatibility
- Automated refactoring tool recommendations
- Comprehensive testing advice
- Staged refactoring for large impacts

### Metrics Calculator

**Cyclomatic Complexity:**
```
Base = 1
+ Each if, for, while, case
+ Each && or ||
+ Each catch, except
+ Language-specific constructs
```

**Cognitive Complexity:**
```
Each decision point = 1 + nesting_level
(Penalizes deeply nested code more heavily)
```

**Maintainability Index:**
```
171 - 5.2 * ln(HV) - 0.23 * CC - 16.2 * ln(LOC)
Simplified: 171 - 5.2*CC - 0.23*CC - 16.2*LOC/10
Range: 0-100 (higher is better)
```

### Snippet Extractor

**Completeness Check:**
- Functions with <5 dependencies → Usually complete
- Few external dependencies → More likely runnable
- Self-contained logic → Complete

**Related Code Detection:**
- Private helper functions in same file
- Type definitions used by the symbol
- Interfaces implemented
- Structs/classes used

### Usage Analyzer

**Pattern Detection:**
```
Patterns identified:
- Heavy usage: >50 total references
- Concentrated usage: >10 in single file
- Wide spread: Used in >20 files
- Call patterns vs assignments vs type references
```

**Deprecation Detection:**
```
Searches documentation for:
- "deprecated"
- "@deprecated"
- "obsolete"
- "use X instead"
```

## 💡 Best Practices for AI Agents

### 1. Start with Context

Always use `get_code_context` first to understand a symbol before modifying it.

```
AI: I need to refactor calculateDiscount()
1. get_code_context("calculateDiscount", depth=3)
2. See how it's actually used
3. Understand dependencies
4. Check usage examples
5. Then plan refactoring
```

### 2. Assess Impact Before Changes

Use `analyze_change_impact` before any refactoring.

```
AI: Should I rename this function?
1. analyze_change_impact("oldFunctionName")
2. Check risk level
3. Review affected files
4. Read suggestions
5. Decide based on impact
```

### 3. Use Metrics for Prioritization

Check metrics to prioritize refactoring work.

```
AI: Which functions need refactoring most?
1. get_code_metrics for each function
2. Sort by complexity or quality
3. Focus on "poor" quality functions
4. Consider effort vs. benefit
```

### 4. Extract Smart Snippets for Examples

Use smart snippets when generating documentation or examples.

```
AI: Document this API
1. extract_smart_snippet("ApiMethod")
2. Get complete example with imports
3. Use usage hints for documentation
4. Include related code if needed
```

### 5. Find Opportunities Proactively

Use suggestion tools to find refactoring opportunities.

```
AI: How can I improve this codebase?
1. suggest_refactorings for key symbols
2. find_unused_symbols
3. get_usage_statistics for APIs
4. Prioritize by impact and effort
```

## 🔧 Configuration

All AI features are automatically initialized when the indexer starts. No configuration needed!

```go
indexer, err := core.NewIndexer(projectPath, nil)
// AI helpers automatically available
```

## 📈 Performance

AI features are optimized for speed:
- **Context extraction**: ~50-100ms per symbol
- **Impact analysis**: ~100-200ms (depends on usage)
- **Metrics calculation**: ~10-50ms per function
- **Snippet extraction**: ~50-100ms
- **Usage analysis**: ~100-300ms (depends on references)

All operations use efficient database queries and caching where possible.

## 🚀 Future Enhancements

Planned AI features:
- **Semantic code search** with embeddings
- **Code pattern detection** (design patterns)
- **Similar code finder** (code clone detection)
- **Test coverage mapping**
- **Documentation quality scoring**
- **API evolution tracking**
- **Breaking change detection**
- **Automatic fix suggestions**

## 📚 Examples

See [EXAMPLES.md](./EXAMPLES.md) for detailed usage examples of each AI-powered tool.

---

**These AI features make Code Indexer MCP the most powerful code analysis tool for AI agents!** 🎉
