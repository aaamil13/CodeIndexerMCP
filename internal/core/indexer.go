package core

import (
	gosql "database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/aaamil13/CodeIndexerMCP/internal/ai"
	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/parser"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/bash"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/c"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/config"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/cpp"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/csharp"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/css"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/golang"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/html"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/java"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/kotlin"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/php"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/powershell"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/python"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/rst"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/ruby"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/rust"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/sql"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/swift"
	"github.com/aaamil13/CodeIndexerMCP/internal/parsers/typescript"
	"github.com/aaamil13/CodeIndexerMCP/internal/utils"
)

// Indexer is the main code indexer
type Indexer struct {
	projectPath      string
	db               *database.Manager
	parsers          *parser.Registry
	ignoreMatcher    *utils.IgnoreMatcher
	project          *model.Project
	logger           *utils.Logger
	config           *Config
	watcher          *Watcher
	// AI helpers
	contextExtractor *ai.ContextExtractor
	impactAnalyzer   *ai.ImpactAnalyzer
	metricsCalc      *ai.MetricsCalculator
	snippetExtractor *ai.SnippetExtractor
	usageAnalyzer    *ai.UsageAnalyzer
	changeTracker    *ai.ChangeTracker
	depGraphBuilder  *ai.DependencyGraphBuilder
	typeValidator    *ai.TypeValidator
}

// Config holds indexer configuration
type Config struct {
	IndexDir    string   // Directory for index data (default: .projectIndex)
	WorkerCount int      // Number of parallel workers (default: CPU count)
	BatchSize   int      // Batch size for database operations
	Exclude     []string // Additional exclude patterns
}

// NewIndexer creates a new indexer for the given project path
func NewIndexer(projectPath string, cfg *Config) (*Indexer, error) {
	if cfg == nil {
		cfg = &Config{
			IndexDir:    ".projectIndex",
			WorkerCount: runtime.NumCPU(),
			BatchSize:   100,
		}
	}

	logger := utils.NewLogger("[Indexer]")

	// Initialize parser registry
	reg := parser.NewRegistry()

	// Register all built-in parsers (23 languages)
	// Core languages
	reg.RegisterParser(golang.NewParser())
	reg.RegisterParser(python.NewParser())
	reg.RegisterParser(typescript.NewTypeScriptParser())

	// JVM languages
	reg.RegisterParser(java.NewParser())
	reg.RegisterParser(kotlin.NewParser())

	// .NET languages
	reg.RegisterParser(csharp.NewParser())

	// System languages
	reg.RegisterParser(c.NewParser())
	reg.RegisterParser(cpp.NewParser())
	reg.RegisterParser(rust.NewParser())

	// Web languages
	reg.RegisterParser(php.NewParser())
	reg.RegisterParser(ruby.NewParser())

	// Mobile
	reg.RegisterParser(swift.NewParser())

	// Scripting
	reg.RegisterParser(bash.NewParser())
	reg.RegisterParser(powershell.NewParser())

	// Database
	reg.RegisterParser(sql.NewParser())

	// Web markup and styling
	reg.RegisterParser(html.NewParser())
	reg.RegisterParser(css.NewParser())

	// Configuration files
	reg.RegisterParser(config.NewJSONParser())
	reg.RegisterParser(config.NewYAMLParser())
	reg.RegisterParser(config.NewTOMLParser())
	reg.RegisterParser(config.NewXMLParser())
	reg.RegisterParser(config.NewMarkdownParser())

	// Documentation
	reg.RegisterParser(rst.NewParser())

	// Initialize ignore matcher
	ignoreMatcher, err := utils.NewIgnoreMatcher(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create ignore matcher: %w", err)
	}

	indexer := &Indexer{
		projectPath:   projectPath,
		parsers:       reg,
		ignoreMatcher: ignoreMatcher,
		logger:        logger,
		config:        cfg,
	}

	return indexer, nil
}

// Initialize initializes the indexer and database
func (idx *Indexer) Initialize() error {
	idx.logger.Info("Initializing indexer for project:", idx.projectPath)

	// Create index directory
	indexDir := filepath.Join(idx.projectPath, idx.config.IndexDir)
	if err := utils.EnsureDir(indexDir); err != nil {
		return fmt.Errorf("failed to create index directory: %w", err)
	}

	// Open database
	dbPath := filepath.Join(indexDir, "index_test.db") // Changed from index.db
	db, err := database.NewManager(dbPath, idx.logger) // Pass idx.logger here
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	idx.db = db

	// Get or create project
	projectName := filepath.Base(idx.projectPath)
	project, err := idx.db.GetProject(idx.projectPath)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	if project == nil {
		// Create new project
		project = &model.Project{
			Path:          idx.projectPath,
			Name:          projectName,
			LanguageStats: make(map[string]int), // Initialize empty map
			CreatedAt:     time.Now(),
			LastIndexed:   time.Time{}, // Initialize to zero time
		}

		if err := idx.db.CreateProject(project); err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
		idx.project = project // Update idx.project with the newly created project (which now has an ID)

		idx.logger.Info("Created new project:", projectName)
	} else {
		// Ensure LanguageStats is not nil if loaded from DB
		if project.LanguageStats == nil {
			project.LanguageStats = make(map[string]int)
		}
		idx.logger.Info("Loaded existing project:", projectName)
	}

	idx.project = project
	idx.logger.Debug("Initialized project with ID:", idx.project.ID)

	// Initialize AI helpers
	idx.contextExtractor = ai.NewContextExtractor(idx.db)
	idx.impactAnalyzer = ai.NewImpactAnalyzer(idx.db)
	idx.metricsCalc = ai.NewMetricsCalculator(idx.db)
	idx.snippetExtractor = ai.NewSnippetExtractor(idx.db)
	idx.usageAnalyzer = ai.NewUsageAnalyzer(idx.db)
	idx.changeTracker = ai.NewChangeTracker(idx.db)
	idx.depGraphBuilder = ai.NewDependencyGraphBuilder(idx.db)
	idx.typeValidator = ai.NewTypeValidator(idx.db)

	return nil
}

// Close closes the indexer and releases resources
func (idx *Indexer) Close() error {
	if idx.db != nil {
		err := idx.db.Close()
		idx.db = nil // Set db to nil after closing
		return err
	}
	return nil
}

// IndexAll indexes all files in the project
func (idx *Indexer) IndexAll() error {
	if idx.db == nil {
		return fmt.Errorf("indexer is closed")
	}
	idx.logger.Info("Starting full index of project")
	startTime := time.Now()

	// Scan for files
	files, err := idx.scanFiles()
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	idx.logger.Infof("Found %d files to index", len(files))

	// Index files concurrently
	if err := idx.indexFiles(files); err != nil {
		return fmt.Errorf("failed to index files: %w", err)
	}

	// Update project stats
	idx.project.LastIndexed = time.Now()
	// Recalculate language stats
	allFiles, err := idx.db.GetAllFilesForProject(idx.project.ID)
	if err != nil {
		return fmt.Errorf("failed to get all files for language stats: %w", err)
	}
	newLanguageStats := make(map[string]int)
	for _, file := range allFiles {
		newLanguageStats[file.Language]++
	}
	idx.project.LanguageStats = newLanguageStats

	if err := idx.db.UpdateProject(idx.project); err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	duration := time.Since(startTime)
	idx.logger.Infof("Indexing completed in %v", duration)

	return nil
}

// IndexFile indexes a single file
func (idx *Indexer) IndexFile(filePath string) error {
	// Make path relative to project
	relPath, err := filepath.Rel(idx.projectPath, filePath)
	if err != nil {
		idx.logger.Errorf("Failed to get relative path for %s: %v", filePath, err)
		return err
	}

	// Check if should ignore
	if idx.ignoreMatcher.ShouldIgnore(relPath) {
		idx.logger.Debugf("Ignoring file: %s", relPath)
		return nil
	}

	// Check if we can parse this file
	if _, err := idx.parsers.GetParserForFile(filePath); err != nil {
		idx.logger.Debugf("Skipping unsupported file: %s", relPath)
		return nil // Skip unsupported files silently
	}

	idx.logger.Debugf("Indexing file: %s", relPath)

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		idx.logger.Errorf("Failed to get file info for %s: %v", relPath, err)
		return err
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		idx.logger.Errorf("Failed to read file content for %s: %v", relPath, err)
		return err
	}

	// Calculate hash
	hash := utils.HashBytes(content)

	// Check if file has changed
	existingFile, err := idx.db.GetFileByPath(idx.project.ID, relPath)
	if err != nil {
		idx.logger.Errorf("Failed to get existing file from DB for %s: %v", relPath, err)
		return err
	}

	if existingFile != nil && existingFile.Hash == hash {
		// File hasn't changed, skip
		idx.logger.Debugf("File unchanged, skipping: %s", relPath)
		return nil
	}

	// Parse file
	parser, err := idx.parsers.GetParserForFile(filePath)
	if err != nil {
		idx.logger.Errorf("Failed to get parser for %s: %v", relPath, err)
		return err
	}

	parseResult, err := parser.Parse(content, filePath)
	if err != nil {
		idx.logger.Warnf("Failed to parse %s: %v", relPath, err)
		return nil // Don't fail on parse errors
	}
	idx.logger.Debugf("Parsed file %s, found %d symbols", relPath, len(parseResult.Symbols))

	// Count lines
	lines, _ := utils.CountLines(filePath)

	// Save to database in transaction
	err = idx.db.Transaction(func(tx *gosql.Tx) error {
		// Save file
		file := &model.File{
			ProjectID:    idx.project.ID,
			Path:         filePath,
			RelativePath: relPath,
			Language:     parser.Language(),
			Size:         fileInfo.Size(),
			LinesOfCode:  lines,
			Hash:         hash,
			LastModified: fileInfo.ModTime(),
			LastIndexed:  time.Now(),
		}

		if existingFile != nil {
			file.ID = existingFile.ID // Set the ID for update
		}

		if err := idx.db.SaveFile(file); err != nil { // Use idx.db directly
			return fmt.Errorf("failed to save file %s: %w", relPath, err)
		}

		// Delete old symbols/imports for this file
		if existingFile != nil {
			idx.db.DeleteSymbolsByFile(file.ID) // Use idx.db directly
			idx.db.DeleteImportsByFile(file.ID) // Use idx.db directly
		}

		// Save symbols
		for _, symbol := range parseResult.Symbols {
			symbol.File = file.Path // Set the file path for the symbol
			symbol.ID = utils.GenerateID(symbol.File, symbol.Name, string(symbol.Kind), symbol.Range.Start.Line)
			symbol.CreatedAt = time.Now()
			symbol.UpdatedAt = time.Now()
			symbol.Language = file.Language   // Populate Language from file
			symbol.ContentHash = file.Hash    // Populate ContentHash from file
			if symbol.Status == "" {
				symbol.Status = model.StatusCompleted // Ensure status is set
			}
			if err := idx.db.SaveSymbol(symbol); err != nil { // Use idx.db directly
				return fmt.Errorf("failed to save symbol %s in file %s: %w", symbol.Name, relPath, err)
			}
		}

		// Save imports
		for _, imp := range parseResult.Imports {
			imp.FilePath = file.Path // Set the file path for the import
			if err := idx.db.SaveImport(imp); err != nil { // Use idx.db directly
				return fmt.Errorf("failed to save import %s in file %s: %w", imp.Path, relPath, err)
			}
		}

		// Save relationships
		for _, rel := range parseResult.Relationships {
			rel.FilePath = file.Path // Set the file path for the relationship
			if err := idx.db.SaveRelationship(rel); err != nil { // Use idx.db directly
				return fmt.Errorf("failed to save relationship in file %s: %w", relPath, err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to save parse results: %w", err)
	}

	idx.logger.Debugf("Indexed file: %s (%d symbols, %d imports)",
		relPath, len(parseResult.Symbols), len(parseResult.Imports))

	return nil
}

// scanFiles scans the project directory for files
func (idx *Indexer) scanFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(idx.projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			idx.logger.Errorf("Error walking file %s: %v", path, err)
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Check if should ignore this directory
			relPath, _ := filepath.Rel(idx.projectPath, path)
			if relPath != "." && idx.ignoreMatcher.ShouldIgnore(relPath) {
				idx.logger.Debugf("Ignoring directory: %s", relPath)
				return filepath.SkipDir
			}
			return nil
		}

		// Check if should ignore this file
		relPath, _ := filepath.Rel(idx.projectPath, path)
		if idx.ignoreMatcher.ShouldIgnore(relPath) {
			idx.logger.Debugf("Ignoring file: %s", relPath)
			return nil
		}

		// Check if we can parse this file
		if _, err := idx.parsers.GetParserForFile(path); err == nil {
			files = append(files, path)
		} else {
			idx.logger.Debugf("No parser found for file: %s", relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan files: %w", err)
	}

	idx.logger.Infof("scanFiles found %d files", len(files))
	return files, err
}

// indexFiles indexes multiple files synchronously
func (idx *Indexer) indexFiles(files []string) error {
	idx.logger.Info("Starting synchronous file indexing")

	for _, filePath := range files {
		if err := idx.IndexFile(filePath); err != nil {
			idx.logger.Errorf("Failed to index %s: %v", filePath, err)
			// Continue to next file, or return error if strict
			// For now, we'll continue to process other files
		}
	}

	idx.logger.Info("Synchronous file indexing completed")
	return nil
}

// SearchSymbols searches for symbols
func (idx *Indexer) SearchSymbols(opts model.SearchOptions) ([]*model.Symbol, error) {
	return idx.db.SearchSymbols(opts)
}

// GetFileStructure returns the structure of a file
func (idx *Indexer) GetFileStructure(filePath string) (*model.ParseResult, error) {
	relPath, err := filepath.Rel(idx.projectPath, filePath)
	if err != nil {
		return nil, err
	}

	file, err := idx.db.GetFileByPath(idx.project.ID, relPath)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("file not found: %s", relPath)
	}

	symbols, err := idx.db.GetSymbolsByFile(file.Path)
	if err != nil {
		return nil, err
	}

	imports, err := idx.db.GetImportsByFile(file.Path)
	if err != nil {
		return nil, err
	}

	return &model.ParseResult{
		FilePath: filePath,
		Language: file.Language,
		Symbols:  symbols,
		Imports:  imports,
	}, nil
}

// GetProjectOverview returns project overview
func (idx *Indexer) GetProjectOverview() (*model.ProjectOverview, error) {
	totalFiles := 0
	files, err := idx.db.GetAllFilesForProject(idx.project.ID)
	if err != nil {
		return nil, err
	}
	totalFiles = len(files)

	totalSymbols := 0
	// TODO: Implement GetTotalSymbols in database.Manager
	// For now, we'll just count symbols from files if available
	for _, file := range files {
		symbols, err := idx.db.GetSymbolsByFile(file.Path)
		if err != nil {
			idx.logger.Warnf("Failed to get symbols for file %s: %v", file.Path, err)
			continue
		}
		totalSymbols += len(symbols)
	}

	return &model.ProjectOverview{
		Project:       idx.project,
		TotalFiles:    totalFiles,
		TotalSymbols:  totalSymbols,
		LanguageStats: idx.project.LanguageStats,
	}, nil
}

// Watch starts watching for file changes and auto-indexes
func (idx *Indexer) Watch() error {
	if idx.watcher != nil {
		return fmt.Errorf("watcher already started")
	}

	watcher, err := NewWatcher(idx)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}

	idx.watcher = watcher

	if err := watcher.Start(); err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}

	idx.logger.Info("File watching started")
	return nil
}

// StopWatch stops the file watcher
func (idx *Indexer) StopWatch() error {
	if idx.watcher == nil {
		return nil
	}

	err := idx.watcher.Stop()
	idx.watcher = nil
	return err
}

// GetSymbolDetails returns detailed information about a symbol
func (idx *Indexer) GetSymbolDetails(symbolName string) (*model.SymbolDetails, error) {
	symbol, err := idx.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	// file, err := idx.db.GetFile(symbol.FileID) // FileID is not directly available in model.Symbol
	// if err != nil {
	// 	return nil, err
	// }

	references, err := idx.db.GetReferencesBySymbol(symbol.ID)
	if err != nil {
		return nil, err
	}

	// relationships, err := idx.db.GetRelationshipsForSymbol(symbol.ID) // This method does not exist in database.Manager
	// if err != nil {
	// 	return nil, err
	// }

	return &model.SymbolDetails{
		Symbol:        symbol,
		// File:          file,
		References:    references,
		// Relationships: relationships,
		Documentation: symbol.Documentation,
	}, nil
}

// FindReferences finds all references to a symbol
func (idx *Indexer) FindReferences(symbolName string) ([]*model.Reference, error) {
	symbol, err := idx.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	return idx.db.GetReferencesBySymbol(symbol.ID)
}

// GetDependencies is deprecated. Use BuildDependencyGraph from the AI helpers instead.
// func (idx *Indexer) GetDependencies(filePath string) (*model.DependencyGraph, error) {
// 	// ... (old implementation removed) ...
// }

// GetAllFiles returns all indexed files
func (idx *Indexer) GetAllFiles() ([]*model.File, error) {
	return idx.db.GetAllFilesForProject(idx.project.ID)
}

// AI Helper Methods

// GetCodeContext extracts comprehensive context for a symbol
func (idx *Indexer) GetCodeContext(symbolName string, depth int) (*model.CodeContext, error) {
	return idx.contextExtractor.ExtractContext(symbolName, depth)
}

// AnalyzeChangeImpact analyzes the impact of changing a symbol
func (idx *Indexer) AnalyzeChangeImpact(symbolName string) (*model.ChangeImpact, error) {
	return idx.impactAnalyzer.AnalyzeChangeImpact(symbolName)
}

// GetCodeMetrics calculates code quality metrics
func (idx *Indexer) GetCodeMetrics(symbolName string) (*model.CodeMetrics, error) {
	symbol, err := idx.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}
	return idx.metricsCalc.CalculateMetrics(symbol)
}

// ExtractSmartSnippet extracts a self-contained code snippet
func (idx *Indexer) ExtractSmartSnippet(symbolName string) (*model.SmartSnippet, error) {
	symbol, err := idx.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}
	return idx.snippetExtractor.ExtractSmartSnippet(symbol, false)
}

// GetUsageStatistics gets usage statistics for a symbol
func (idx *Indexer) GetUsageStatistics(symbolName string) (*model.SymbolUsageStats, error) {
	symbol, err := idx.db.GetSymbolByName(symbolName)
	if err != nil {
		return nil, err
	}
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}
	return idx.usageAnalyzer.AnalyzeUsage(symbol)
}

// SuggestRefactorings suggests refactoring opportunities
func (idx *Indexer) SuggestRefactorings(symbolName string) ([]*model.RefactoringOpportunity, error) {
	return idx.impactAnalyzer.SuggestRefactorings(symbolName)
}

// FindUnusedSymbols finds unused symbols in the project
func (idx *Indexer) FindUnusedSymbols() ([]*model.Symbol, error) {
	return idx.usageAnalyzer.FindUnusedSymbols(fmt.Sprintf("%d", idx.project.ID))
}

// FindMostUsedSymbols finds the most used symbols
func (idx *Indexer) FindMostUsedSymbols(limit int) ([]*model.SymbolUsageStats, error) {
	return idx.usageAnalyzer.FindMostUsedSymbols(fmt.Sprintf("%d", idx.project.ID), limit)
}

// Change Tracking Methods

// SimulateSymbolChange simulates a change without applying it
func (idx *Indexer) SimulateSymbolChange(symbolName string, changeType model.ChangeType, newValue string) (*model.ChangeImpact, error) {
	return idx.changeTracker.SimulateChange(symbolName, changeType, newValue)
}

// ValidateChanges validates a set of changes
func (idx *Indexer) ValidateChanges(changes []*model.Change) (*model.ValidationResult, error) {
	return idx.changeTracker.ValidateChanges(changes)
}

// GenerateAutoFixes generates automatic fixes for a change
func (idx *Indexer) GenerateAutoFixes(change *model.Change) ([]*model.AutoFixSuggestion, error) {
	return idx.changeTracker.GenerateAutoFixes(change)
}

// Dependency Graph Methods

// BuildDependencyGraph builds a dependency graph for a symbol
func (idx *Indexer) BuildDependencyGraph(symbolName string, maxDepth int) (*model.DependencyGraph, error) {
	return idx.depGraphBuilder.BuildSymbolDependencyGraph(symbolName, maxDepth)
}

// GetDependencies gets all dependencies for a symbol
func (idx *Indexer) GetDependencies(symbolName string) ([]*model.Symbol, error) {
	return idx.depGraphBuilder.GetDependenciesFor(symbolName)
}

// GetDependents gets all symbols that depend on a symbol
func (idx *Indexer) GetDependents(symbolName string) ([]*model.Symbol, error) {
	return idx.depGraphBuilder.GetDependentsFor(symbolName)
}

// AnalyzeDependencyChain analyzes the full dependency chain
func (idx *Indexer) AnalyzeDependencyChain(symbolName string) (map[string]interface{}, error) {
	return idx.depGraphBuilder.AnalyzeDependencyChain(symbolName)
}

// Type Validation Methods

// ValidateFileTypes validates all types in a file
func (idx *Indexer) ValidateFileTypes(filePath string) (*model.TypeValidation, error) {
	return idx.typeValidator.ValidateFile(filePath)
}

// FindUndefinedUsages finds all undefined symbol usages in a file
func (idx *Indexer) FindUndefinedUsages(filePath string) ([]*model.UndefinedUsage, error) {
	return idx.typeValidator.FindUndefinedUsages(filePath)
}

// CheckMethodExists checks if a method exists on a type
func (idx *Indexer) CheckMethodExists(typeName, methodName string) (*model.MissingMethod, error) {
	return idx.typeValidator.CheckMethodExists(typeName, methodName, fmt.Sprintf("%d", idx.project.ID))
}

// CalculateTypeSafetyScore calculates type safety score for a file
func (idx *Indexer) CalculateTypeSafetyScore(filePath string) (*model.TypeSafetyScore, error) {
	return idx.typeValidator.CalculateTypeSafetyScore(filePath)
}

// GetProject returns the current project
func (idx *Indexer) GetProject() *model.Project {
	return idx.project
}