package ai

import (
	"path/filepath"
	"testing"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/utils"
)

func setupTestDBForAI(t *testing.T) *database.Manager {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	logger := utils.NewLogger("[TestChangeTracker]")
	db, err := database.NewManager(dbPath, logger)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create test data
	project := &model.Project{Name: "test", Path: "/test"}
	db.CreateProject(project)

	file := &model.File{ProjectID: project.ID, Path: "/test/file.go", Language: "go"}
	db.SaveFile(file)

	return db
}

func TestAnalyzeSymbolChange_Rename(t *testing.T) {
	db := setupTestDBForAI(t)
	defer db.Close()

	// Create a symbol
	symbol := &model.Symbol{
		File:       "/test/file.go",
		Name:       "OldFunc",
		Kind:       "function",
		Visibility: model.VisibilityPublic,
	}
	db.SaveSymbol(symbol)

	// Create some references
	for i := 0; i < 3; i++ {
		ref := &model.Reference{
			SourceSymbolID: symbol.ID,
			FilePath:       "/test/file.go",
			Line:           10 + i,
			Column:         5,
			ReferenceType:  "call",
		}
		db.SaveReference(ref)
	}

	// Analyze rename change
	tracker := NewChangeTracker(db)
	change := &model.Change{
		Type:   model.ChangeTypeRename,
		Symbol: symbol,
		OldSymbol: &model.Symbol{
			Name: "OldFunc",
		},
	}

	_, err := tracker.AnalyzeSymbolChange(change)
	if err == nil || err.Error() != "not implemented" {
		t.Fatalf("Expected 'not implemented' error, got %v", err)
	}
}

func TestAnalyzeSymbolChange_Delete(t *testing.T) {
	db := setupTestDBForAI(t)
	defer db.Close()

	// Create a symbol
	symbol := &model.Symbol{
		File:       "/test/file.go",
		Name:       "DeprecatedFunc",
		Kind:       "function",
		Visibility: model.VisibilityPublic,
	}
	db.SaveSymbol(symbol)

	// Create references
	ref := &model.Reference{
		SourceSymbolID: symbol.ID,
		FilePath:       "/test/file.go",
		Line:           20,
		Column:         5,
		ReferenceType:  "call",
	}
	db.SaveReference(ref)

	// Analyze delete change
	tracker := NewChangeTracker(db)
	change := &model.Change{
		Type:   model.ChangeTypeDelete,
		Symbol: symbol,
	}

	_, err := tracker.AnalyzeSymbolChange(change)
	if err == nil || err.Error() != "not implemented" {
		t.Fatalf("Expected 'not implemented' error, got %v", err)
	}
}

func TestAnalyzeSymbolChange_Modify(t *testing.T) {
	db := setupTestDBForAI(t)
	defer db.Close()

	// Create a symbol
	symbol := &model.Symbol{
		File:       "/test/file.go",
		Name:       "Calculate",
		Kind:       "function",
		Signature:  "func Calculate(x int) int",
		Visibility: model.VisibilityPublic,
	}
	db.SaveSymbol(symbol)

	// Create references
	ref := &model.Reference{
		SourceSymbolID: symbol.ID,
		FilePath:       "/test/file.go",
		Line:           30,
		ReferenceType:  "call",
	}
	db.SaveReference(ref)

	// Analyze modify change (signature change)
	tracker := NewChangeTracker(db)
	oldSymbol := &model.Symbol{
		Name:      "Calculate",
		Signature: "func Calculate(x int) int",
	}
	newSymbol := &model.Symbol{
		Name:      "Calculate",
		Signature: "func Calculate(x, y int) int",
	}

	change := &model.Change{
		Type:      model.ChangeTypeModify,
		Symbol:    newSymbol,
		OldSymbol: oldSymbol,
	}

	_, err := tracker.AnalyzeSymbolChange(change)
	if err == nil || err.Error() != "not implemented" {
		t.Fatalf("Expected 'not implemented' error, got %v", err)
	}
}

func TestSimulateChange(t *testing.T) {
	db := setupTestDBForAI(t)
	defer db.Close()

	// Create a symbol
	symbol := &model.Symbol{
		File:       "/test/file.go",
		Name:       "TestFunc",
		Kind:       "function",
		Visibility: model.VisibilityPublic,
	}
	db.SaveSymbol(symbol)

	// Simulate change
	tracker := NewChangeTracker(db)
	change := &model.Change{
		Type:   model.ChangeTypeRename,
		Symbol: symbol,
	}

	result, err := tracker.SimulateChange(change.Symbol.Name, change.Type, change.Symbol.Name)
	if err == nil || err.Error() != "not implemented" {
		t.Fatalf("Expected 'not implemented' error, got %v", err)
	}
	if result != nil {
		t.Fatal("Expected nil result from SimulateChange when not implemented")
	}
}

func TestValidateChanges(t *testing.T) {
	db := setupTestDBForAI(t)
	defer db.Close()

	// Create symbols
	symbol1 := &model.Symbol{
		File: "/test/file.go",
		Name:   "Func1",
		Kind:   "function",
	}
	db.SaveSymbol(symbol1)

	symbol2 := &model.Symbol{
		File: "/test/file.go",
		Name:   "Func2",
		Kind:   "function",
	}
	db.SaveSymbol(symbol2)

	// Create changes
	changes := []*model.Change{
		{
			Type:   model.ChangeTypeRename,
			Symbol: symbol1,
		},
		{
			Type:   model.ChangeTypeDelete,
			Symbol: symbol2,
		},
	}

	// Validate changes
	tracker := NewChangeTracker(db)
	result, err := tracker.ValidateChanges(changes)
	if err != nil {
		t.Fatalf("ValidateChanges failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected validation result")
	}

	// Should have analyzed all changes
	if len(result.ChangeSet.Changes) != len(changes) {
		t.Errorf("Expected %d changes, got %d", len(changes), len(result.ChangeSet.Changes))
	}
}

func TestChangeTracker_VisibilityChange(t *testing.T) {
	db := setupTestDBForAI(t)
	defer db.Close()

	// Create a public symbol
	symbol := &model.Symbol{
		File:       "/test/file.go",
		Name:       "PublicFunc",
		Kind:       "function",
		Visibility: model.VisibilityPublic,
	}
	db.SaveSymbol(symbol)

	// Analyze changing to private
	tracker := NewChangeTracker(db)
	oldSymbol := *symbol
	newSymbol := *symbol
	newSymbol.Visibility = model.VisibilityPrivate

	change := &model.Change{
		Type:      model.ChangeTypeModify,
		Symbol:    &newSymbol,
		OldSymbol: &oldSymbol,
	}

	_, err := tracker.AnalyzeSymbolChange(change)
	if err == nil || err.Error() != "not implemented" {
		t.Fatalf("Expected 'not implemented' error, got %v", err)
	}
}

func TestChangeTracker_ConflictDetection(t *testing.T) {
	db := setupTestDBForAI(t)
	defer db.Close()

	// Create existing symbol
	existing := &model.Symbol{
		File: "/test/file.go",
		Name:   "ExistingFunc",
		Kind:   "function",
	}
	db.SaveSymbol(existing)

	// Create symbol to rename
	toRename := &model.Symbol{
		File: "/test/file.go",
		Name:   "OldFunc",
		Kind:   "function",
	}
	db.SaveSymbol(toRename)

	// Try to rename to existing name
	tracker := NewChangeTracker(db)
	newSymbol := *toRename
	newSymbol.Name = "ExistingFunc"

	change := &model.Change{
		Type:      model.ChangeTypeRename,
		Symbol:    &newSymbol,
		OldSymbol: toRename,
	}

	_, err := tracker.AnalyzeSymbolChange(change)
	if err == nil || err.Error() != "not implemented" {
		t.Fatalf("Expected 'not implemented' error, got %v", err)
	}
}
