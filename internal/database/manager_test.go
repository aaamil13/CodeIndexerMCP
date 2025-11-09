package database_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*database.Manager, string) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := utils.NewLogger("[TestDB]")

	dbManager, err := database.NewManager(dbPath, logger)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := dbManager.Close()
		assert.NoError(t, err, "Failed to close database manager")
	})
	return dbManager, dbPath
}

func TestManager_CreateAndGetProject(t *testing.T) {
	db, _ := setupTestDB(t)

	projectPath := "/test/project"
	projectName := "test-project"

	// Create project
	project := &model.Project{
		Path:          projectPath,
		Name:          projectName,
		LanguageStats: map[string]int{"go": 10, "python": 5},
		CreatedAt:     time.Now().Add(-24 * time.Hour),
		LastIndexed:   time.Now().Add(-1 * time.Hour),
	}
	err := db.CreateProject(project)
	require.NoError(t, err)
	assert.NotZero(t, project.ID)

	// Get project
	fetchedProject, err := db.GetProject(projectPath)
	require.NoError(t, err)
	require.NotNil(t, fetchedProject)

	assert.Equal(t, project.ID, fetchedProject.ID)
	assert.Equal(t, project.Path, fetchedProject.Path)
	assert.Equal(t, project.Name, fetchedProject.Name)
	assert.Equal(t, project.LanguageStats, fetchedProject.LanguageStats)
	// Compare time with some tolerance due to possible precision issues in SQLite
	assert.WithinDuration(t, project.CreatedAt, fetchedProject.CreatedAt, time.Second)
	assert.WithinDuration(t, project.LastIndexed, fetchedProject.LastIndexed, time.Second)
}

func TestManager_UpdateProject(t *testing.T) {
	db, _ := setupTestDB(t)

	projectPath := "/test/project"
	projectName := "test-project"

	project := &model.Project{
		Path:          projectPath,
		Name:          projectName,
		LanguageStats: map[string]int{"go": 10},
		CreatedAt:     time.Now().Add(-24 * time.Hour),
		LastIndexed:   time.Now().Add(-1 * time.Hour),
	}
	err := db.CreateProject(project)
	require.NoError(t, err)

	// Update project
	project.Name = "updated-project"
	project.LanguageStats["python"] = 20
	project.LastIndexed = time.Now()
	err = db.UpdateProject(project)
	require.NoError(t, err)

	fetchedProject, err := db.GetProject(projectPath)
	require.NoError(t, err)
	require.NotNil(t, fetchedProject)

	assert.Equal(t, project.Name, fetchedProject.Name)
	assert.Equal(t, project.LanguageStats, fetchedProject.LanguageStats)
	assert.WithinDuration(t, project.LastIndexed, fetchedProject.LastIndexed, time.Second)
}

func TestManager_SaveFileAndGet(t *testing.T) {
	db, _ := setupTestDB(t)

	projectPath := "/test/project"
	projectName := "test-project"
	project := &model.Project{Path: projectPath, Name: projectName}
	db.CreateProject(project)

	filePath := filepath.Join(projectPath, "main.go")
	relPath := "main.go"
	file := &model.File{
		ProjectID:    project.ID,
		Path:         filePath,
		RelativePath: relPath,
		Language:     "go",
		Size:         1024,
		LinesOfCode:  50,
		Hash:         "abcdef123",
		LastModified: time.Now().Add(-time.Hour),
		LastIndexed:  time.Now(),
	}

	err := db.SaveFile(file)
	require.NoError(t, err)
	assert.NotZero(t, file.ID)

	fetchedFile, err := db.GetFileByPath(project.ID, relPath)
	require.NoError(t, err)
	require.NotNil(t, fetchedFile)

	assert.Equal(t, file.ID, fetchedFile.ID)
	assert.Equal(t, file.Path, fetchedFile.Path)
	assert.Equal(t, file.RelativePath, fetchedFile.RelativePath)
	assert.Equal(t, file.Language, fetchedFile.Language)
	assert.Equal(t, file.Size, fetchedFile.Size)
	assert.Equal(t, file.LinesOfCode, fetchedFile.LinesOfCode)
	assert.Equal(t, file.Hash, fetchedFile.Hash)
	assert.WithinDuration(t, file.LastModified, fetchedFile.LastModified, time.Second)
	assert.WithinDuration(t, file.LastIndexed, fetchedFile.LastIndexed, time.Second)
}

func TestManager_SaveSymbolAndGet(t *testing.T) {
	db, _ := setupTestDB(t)

	// Create project and file first
	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.CreateProject(project)
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	symbol := &model.Symbol{
		ID:            "sym1",
		Name:          "MyFunc",
		Kind:          model.SymbolKindFunction,
		File:          file.Path,
		Language:      "go",
		Signature:     "func MyFunc()",
		Documentation: "A test function",
		Visibility:    model.VisibilityPublic,
		Range: model.Range{
			Start: model.Position{Line: 1, Column: 1, Byte: 0},
			End:   model.Position{Line: 5, Column: 1, Byte: 100},
		},
		ContentHash:   "hash123",
		Status:        model.StatusCompleted,
		Priority:      5,
		AssignedAgent: "AI",
		CreatedAt:     time.Now().Add(-time.Minute),
		UpdatedAt:     time.Now(),
		Metadata:      map[string]string{"key": "value"},
	}

	err := db.SaveSymbol(symbol)
	require.NoError(t, err)

	fetchedSymbol, err := db.GetSymbol(symbol.ID)
	require.NoError(t, err)
	require.NotNil(t, fetchedSymbol)

	assert.Equal(t, symbol.ID, fetchedSymbol.ID)
	assert.Equal(t, symbol.Name, fetchedSymbol.Name)
	assert.Equal(t, symbol.Kind, fetchedSymbol.Kind)
	assert.Equal(t, symbol.File, fetchedSymbol.File)
	assert.Equal(t, symbol.Language, fetchedSymbol.Language)
	assert.Equal(t, symbol.Signature, fetchedSymbol.Signature)
	assert.Equal(t, symbol.Documentation, fetchedSymbol.Documentation)
	assert.Equal(t, symbol.Visibility, fetchedSymbol.Visibility)
	assert.Equal(t, symbol.Range, fetchedSymbol.Range)
	assert.Equal(t, symbol.ContentHash, fetchedSymbol.ContentHash)
	assert.Equal(t, symbol.Status, fetchedSymbol.Status)
	assert.Equal(t, symbol.Priority, fetchedSymbol.Priority)
	assert.Equal(t, symbol.AssignedAgent, fetchedSymbol.AssignedAgent)
	assert.WithinDuration(t, symbol.CreatedAt, fetchedSymbol.CreatedAt, time.Second)
	assert.WithinDuration(t, symbol.UpdatedAt, fetchedSymbol.UpdatedAt, time.Second)
	assert.Equal(t, symbol.Metadata, fetchedSymbol.Metadata)
}

func TestManager_SaveFunctionAndGet(t *testing.T) {
	db, _ := setupTestDB(t)

	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.CreateProject(project)
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	function := &model.Function{
		Symbol: model.Symbol{
			ID:   "func1",
			Name: "Add",
			Kind: model.SymbolKindFunction,
			File: file.Path,
			Language: "go",
			Signature: "(a int, b int) int",
			ContentHash: "funcHash1",
		},
		Parameters: []model.Parameter{
			{Name: "a", Type: "int"},
			{Name: "b", Type: "int"},
		},
		ReturnType: "int",
		Body:       "return a + b",
	}

	err := db.SaveFunction(function)
	require.NoError(t, err)

	// Verify symbol is saved
	fetchedSymbol, err := db.GetSymbol(function.ID)
	require.NoError(t, err)
	require.NotNil(t, fetchedSymbol)
	assert.Equal(t, function.Name, fetchedSymbol.Name)

	// TODO: Add actual GetFunctionDetails/GetParameters to Manager to fully test
}

func TestManager_SaveClassAndGet(t *testing.T) {
	db, _ := setupTestDB(t)

	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.CreateProject(project)
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	class := &model.Class{
		Symbol: model.Symbol{
			ID:          "class1",
			Name:        "MyClass",
			Kind:        model.SymbolKindClass,
			File:        file.Path,
			Language:    "go",
			Signature:   "type MyClass struct",
			ContentHash: "classHash1",
		},
		IsAbstract: false,
		Fields: []model.Field{
			{Name: "Field1", Type: "string", Visibility: string(model.VisibilityPublic)},
		},
	}

	err := db.SaveClass(class)
	require.NoError(t, err)

	// Verify symbol is saved
	fetchedSymbol, err := db.GetSymbol(class.ID)
	require.NoError(t, err)
	require.NotNil(t, fetchedSymbol)
	assert.Equal(t, class.Name, fetchedSymbol.Name)

	// TODO: Add actual GetClassDetails/GetFields to Manager to fully test
}

func TestManager_SaveMethod(t *testing.T) {
	db, _ := setupTestDB(t)

	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.CreateProject(project)
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	method := &model.Method{
		Function: model.Function{
			Symbol: model.Symbol{
				ID:   "method1",
				Name: "MyMethod",
				Kind: model.SymbolKindMethod,
				File: file.Path,
				Language: "go",
				Signature: "(c *MyClass) MyMethod()",
				ContentHash: "methodHash1",
			},
			Parameters: []model.Parameter{
				{Name: "param1", Type: "string"},
			},
			ReturnType: "error",
			Body:       "return nil",
		},
		ReceiverType: "*MyClass",
		IsStatic:     false,
	}

	err := db.SaveMethod(method)
	require.NoError(t, err)

	// Verify symbol is saved
	fetchedSymbol, err := db.GetSymbol(method.ID)
	require.NoError(t, err)
	require.NotNil(t, fetchedSymbol)
	assert.Equal(t, method.Name, fetchedSymbol.Name)
}

func TestManager_HasSymbolChanged(t *testing.T) {
	db, _ := setupTestDB(t)

	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.CreateProject(project)
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	symbolID := "sym_changed_test"
	initialHash := "initialHash"
	updatedHash := "updatedHash"

	symbol := &model.Symbol{
		ID:          symbolID,
		Name:        "TestFunc",
		Kind:        model.SymbolKindFunction,
		File:        file.Path,
		Language:    "go",
		ContentHash: initialHash,
		Range:       model.Range{},
	}

	// New symbol, should report as changed
	changed, err := db.HasSymbolChanged(symbolID, initialHash)
	require.NoError(t, err)
	assert.True(t, changed)

	// Save initial symbol
	err = db.SaveSymbol(symbol)
	require.NoError(t, err)

	// Same hash, should not be changed
	changed, err = db.HasSymbolChanged(symbolID, initialHash)
	require.NoError(t, err)
	assert.False(t, changed)

	// Different hash, should be changed
	changed, err = db.HasSymbolChanged(symbolID, updatedHash)
	require.NoError(t, err)
	assert.True(t, changed)
}

func TestManager_SaveSymbolIfChanged(t *testing.T) {
	db, _ := setupTestDB(t)

	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.CreateProject(project)
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	symbolID := "sym_save_if_changed"
	hashV1 := "hashV1"
	hashV2 := "hashV2"

	// New symbol
	symbolV1 := &model.Symbol{
		ID:          symbolID,
		Name:        "FuncV1",
		Kind:        model.SymbolKindFunction,
		File:        file.Path,
		Language:    "go",
		ContentHash: hashV1,
		Range:       model.Range{},
	}
	saved, err := db.SaveSymbolIfChanged(symbolV1)
	require.NoError(t, err)
	assert.True(t, saved, "Expected symbol to be saved initially")

	// No change, should not save
	saved, err = db.SaveSymbolIfChanged(symbolV1)
	require.NoError(t, err)
	assert.False(t, saved, "Expected symbol not to be saved if hash is same")

	// Change hash, should save
	symbolV2 := &model.Symbol{
		ID:          symbolID,
		Name:        "FuncV2", // Name change also
		Kind:        model.SymbolKindFunction,
		File:        file.Path,
		Language:    "go",
		ContentHash: hashV2,
		Range:       model.Range{},
	}
	saved, err = db.SaveSymbolIfChanged(symbolV2)
	require.NoError(t, err)
	assert.True(t, saved, "Expected symbol to be saved if hash is different")

	fetchedSymbol, err := db.GetSymbol(symbolID)
	require.NoError(t, err)
	assert.NotNil(t, fetchedSymbol)
	assert.Equal(t, "FuncV2", fetchedSymbol.Name) // Verify name update
	assert.Equal(t, hashV2, fetchedSymbol.ContentHash)
}

func TestManager_SaveFileSymbols(t *testing.T) {
	db, _ := setupTestDB(t)

	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.CreateProject(project)
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	func1 := &model.Function{
		Symbol: model.Symbol{
			ID:   "func1", Name: "F1", Kind: model.SymbolKindFunction, File: file.Path, Language: "go", ContentHash: "h1", Range: model.Range{}},
	}
	class1 := &model.Class{
		Symbol: model.Symbol{
			ID:   "class1", Name: "C1", Kind: model.SymbolKindClass, File: file.Path, Language: "go", ContentHash: "h2", Range: model.Range{}},
	}
	method1 := &model.Method{
		Function: model.Function{
			Symbol: model.Symbol{
				ID:   "method1", Name: "M1", Kind: model.SymbolKindMethod, File: file.Path, Language: "go", ContentHash: "h3", Range: model.Range{}},
		}}

	fileSymbols := &model.FileSymbols{
		FilePath: file.Path,
		Language: "go",
		Functions: []*model.Function{func1},
		Classes: []*model.Class{class1},
		Methods: []*model.Method{method1},
	}

	err := db.SaveFileSymbols(fileSymbols)
	require.NoError(t, err)

	// Verify symbols are saved
	symbols, err := db.GetSymbolsByFile(file.Path)
	require.NoError(t, err)
	assert.Len(t, symbols, 3)

	// Update symbols and save again
	func2 := &model.Function{
		Symbol: model.Symbol{
			ID:   "func2", Name: "F2", Kind: model.SymbolKindFunction, File: file.Path, Language: "go", ContentHash: "h4", Range: model.Range{}},
	}
	fileSymbols.Functions = []*model.Function{func2}
	fileSymbols.Classes = []*model.Class{} // Removed class
	fileSymbols.Methods = []*model.Method{} // Removed method

	err = db.SaveFileSymbols(fileSymbols)
	require.NoError(t, err)

	symbols, err = db.GetSymbolsByFile(file.Path)
	require.NoError(t, err)
	assert.Len(t, symbols, 1) // Only func2 should remain
	assert.Equal(t, "func2", symbols[0].ID)
}

func TestManager_AITaskMethods(t *testing.T) {
	db, _ := setupTestDB(t)

	// Create a project
	project := &model.Project{Path: "/test/project", Name: "test-project"}
	db.CreateProject(project)
	file := &model.File{ProjectID: project.ID, Path: "/test/project/main.go", RelativePath: "main.go", Language: "go"}
	db.SaveFile(file)

	// Create a symbol
	symbol := &model.Symbol{
		ID:       "sym_ai",
		Name:     "AISymbol",
		Kind:     model.SymbolKindFunction,
		File:     file.Path,
		Language: "go",
		Range:    model.Range{},
		Status:   model.StatusPlanned,
	}
	db.SaveSymbol(symbol)

	// Test CreateBuildTask
	task := &model.BuildTask{
		ID:           "task1",
		Type:         "implement_feature",
		TargetSymbol: symbol.ID,
		Description:  "Implement new AI feature",
		Status:       model.StatusPlanned,
		Priority:     10,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := db.CreateBuildTask(task)
	require.NoError(t, err)

	// Test GetTasksByStatus
	plannedTasks, err := db.GetTasksByStatus(model.StatusPlanned)
	require.NoError(t, err)
	assert.Len(t, plannedTasks, 1)
	assert.Equal(t, task.ID, plannedTasks[0].ID)

	// Test UpdateSymbolStatus
	err = db.UpdateSymbolStatus(symbol.ID, model.StatusInProgress)
	require.NoError(t, err)

	fetchedSymbol, err := db.GetSymbol(symbol.ID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusInProgress, fetchedSymbol.Status)

	// Test GetSymbolsByStatus
	inProgressSymbols, err := db.GetSymbolsByStatus(model.StatusInProgress)
	require.NoError(t, err)
	assert.Len(t, inProgressSymbols, 1)
	assert.Equal(t, symbol.ID, inProgressSymbols[0].ID)
}
