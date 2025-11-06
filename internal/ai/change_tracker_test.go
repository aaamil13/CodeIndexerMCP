package ai

import (
	"path/filepath"
	"testing"

	"github.com/aaamil13/CodeIndexerMCP/internal/database"
	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

func setupTestDBForAI(t *testing.T) *database.Database {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create test data
	project := &types.Project{Name: "test", Path: "/test"}
	db.CreateProject(project)

	file := &types.File{ProjectID: project.ID, Path: "/test/file.go", Language: "go"}
	db.CreateFile(file)

	return db
}

func TestAnalyzeSymbolChange_Rename(t *testing.T) {
	db := setupTestDBForAI(t)
	defer db.Close()

	// Create a symbol
	symbol := &types.Symbol{
		FileID:     1,
		Name:       "OldFunc",
		Type:       types.SymbolTypeFunction,
		Visibility: types.VisibilityPublic,
	}
	db.CreateSymbol(symbol)

	// Create some references
	for i := 0; i < 3; i++ {
		ref := &types.Reference{
			SymbolID: symbol.ID,
			FileID:   1,
			Line:     10 + i,
			Column:   5,
			Context:  "OldFunc()",
		}
		db.CreateReference(ref)
	}

	// Analyze rename change
	tracker := NewChangeTracker(db)
	change := &types.Change{
		Type:   types.ChangeTypeRename,
		Symbol: symbol,
		OldSymbol: &types.Symbol{
			Name: "OldFunc",
		},
	}

	impact, err := tracker.AnalyzeSymbolChange(change)
	if err != nil {
		t.Fatalf("AnalyzeSymbolChange failed: %v", err)
	}

	// Should have affected symbols (references)
	if len(impact.AffectedSymbols) == 0 {
		t.Error("Expected affected symbols for rename")
	}

	// Should have required updates
	if len(impact.RequiredUpdates) == 0 {
		t.Error("Expected required updates for rename")
	}

	// Should have auto-fix suggestions
	if len(impact.AutoFixSuggestions) == 0 {
		t.Error("Expected auto-fix suggestions for rename")
	}

	// Should be able to auto-fix
	if !impact.CanAutoFix {
		t.Error("Rename should be auto-fixable")
	}
}

func TestAnalyzeSymbolChange_Delete(t *testing.T) {
	db := setupTestDBForAI(t)
	defer db.Close()

	// Create a symbol
	symbol := &types.Symbol{
		FileID:     1,
		Name:       "DeprecatedFunc",
		Type:       types.SymbolTypeFunction,
		Visibility: types.VisibilityPublic,
	}
	db.CreateSymbol(symbol)

	// Create references
	ref := &types.Reference{
		SymbolID: symbol.ID,
		FileID:   1,
		Line:     20,
		Column:   5,
		Context:  "DeprecatedFunc()",
	}
	db.CreateReference(ref)

	// Analyze delete change
	tracker := NewChangeTracker(db)
	change := &types.Change{
		Type:   types.ChangeTypeDelete,
		Symbol: symbol,
	}

	impact, err := tracker.AnalyzeSymbolChange(change)
	if err != nil {
		t.Fatalf("AnalyzeSymbolChange failed: %v", err)
	}

	// Should have broken references
	if len(impact.BrokenReferences) == 0 {
		t.Error("Expected broken references for delete")
	}

	// Should have validation errors
	if len(impact.ValidationErrors) == 0 {
		t.Error("Expected validation errors for delete")
	}

	// Should NOT be auto-fixable
	if impact.CanAutoFix {
		t.Error("Delete with references should not be auto-fixable")
	}
}

func TestAnalyzeSymbolChange_Modify(t *testing.T) {
	db := setupTestDBForAI(t)
	defer db.Close()

	// Create a symbol
	symbol := &types.Symbol{
		FileID:     1,
		Name:       "Calculate",
		Type:       types.SymbolTypeFunction,
		Signature:  "func Calculate(x int) int",
		Visibility: types.VisibilityPublic,
	}
	db.CreateSymbol(symbol)

	// Create references
	ref := &types.Reference{
		SymbolID: symbol.ID,
		FileID:   1,
		Line:     30,
	}
	db.CreateReference(ref)

	// Analyze modify change (signature change)
	tracker := NewChangeTracker(db)
	oldSymbol := &types.Symbol{
		Name:      "Calculate",
		Signature: "func Calculate(x int) int",
	}
	newSymbol := &types.Symbol{
		Name:      "Calculate",
		Signature: "func Calculate(x, y int) int",
	}

	change := &types.Change{
		Type:      types.ChangeTypeModify,
		Symbol:    newSymbol,
		OldSymbol: oldSymbol,
	}

	impact, err := tracker.AnalyzeSymbolChange(change)
	if err != nil {
		t.Fatalf("AnalyzeSymbolChange failed: %v", err)
	}

	// Should detect signature change
	if len(impact.ValidationErrors) == 0 {
		t.Error("Expected validation errors for signature change")
	}
}

func TestSimulateChange(t *testing.T) {
	db := setupTestDBForAI(t)
	defer db.Close()

	// Create a symbol
	symbol := &types.Symbol{
		FileID:     1,
		Name:       "TestFunc",
		Type:       types.SymbolTypeFunction,
		Visibility: types.VisibilityPublic,
	}
	db.CreateSymbol(symbol)

	// Simulate change
	tracker := NewChangeTracker(db)
	change := &types.Change{
		Type:   types.ChangeTypeRename,
		Symbol: symbol,
	}

	result, err := tracker.SimulateChange(change)
	if err != nil {
		t.Fatalf("SimulateChange failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result from SimulateChange")
	}

	// Should return impact analysis
	if result.Changes == nil {
		t.Error("Expected changes in result")
	}
}

func TestValidateChanges(t *testing.T) {
	db := setupTestDBForAI(t)
	defer db.Close()

	// Create symbols
	symbol1 := &types.Symbol{
		FileID: 1,
		Name:   "Func1",
		Type:   types.SymbolTypeFunction,
	}
	db.CreateSymbol(symbol1)

	symbol2 := &types.Symbol{
		FileID: 1,
		Name:   "Func2",
		Type:   types.SymbolTypeFunction,
	}
	db.CreateSymbol(symbol2)

	// Create changes
	changes := []*types.Change{
		{
			Type:   types.ChangeTypeRename,
			Symbol: symbol1,
		},
		{
			Type:   types.ChangeTypeDelete,
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
	if len(result.Changes) != len(changes) {
		t.Errorf("Expected %d changes, got %d", len(changes), len(result.Changes))
	}
}

func TestChangeTracker_VisibilityChange(t *testing.T) {
	db := setupTestDBForAI(t)
	defer db.Close()

	// Create a public symbol
	symbol := &types.Symbol{
		FileID:     1,
		Name:       "PublicFunc",
		Type:       types.SymbolTypeFunction,
		Visibility: types.VisibilityPublic,
	}
	db.CreateSymbol(symbol)

	// Analyze changing to private
	tracker := NewChangeTracker(db)
	oldSymbol := *symbol
	newSymbol := *symbol
	newSymbol.Visibility = types.VisibilityPrivate

	change := &types.Change{
		Type:      types.ChangeTypeModify,
		Symbol:    &newSymbol,
		OldSymbol: &oldSymbol,
	}

	impact, err := tracker.AnalyzeSymbolChange(change)
	if err != nil {
		t.Fatalf("AnalyzeSymbolChange failed: %v", err)
	}

	// Should warn about visibility change
	if len(impact.ValidationErrors) == 0 {
		t.Error("Expected validation errors for visibility change")
	}
}

func TestChangeTracker_ConflictDetection(t *testing.T) {
	db := setupTestDBForAI(t)
	defer db.Close()

	// Create existing symbol
	existing := &types.Symbol{
		FileID: 1,
		Name:   "ExistingFunc",
		Type:   types.SymbolTypeFunction,
	}
	db.CreateSymbol(existing)

	// Create symbol to rename
	toRename := &types.Symbol{
		FileID: 1,
		Name:   "OldFunc",
		Type:   types.SymbolTypeFunction,
	}
	db.CreateSymbol(toRename)

	// Try to rename to existing name
	tracker := NewChangeTracker(db)
	newSymbol := *toRename
	newSymbol.Name = "ExistingFunc"

	change := &types.Change{
		Type:      types.ChangeTypeRename,
		Symbol:    &newSymbol,
		OldSymbol: toRename,
	}

	impact, err := tracker.AnalyzeSymbolChange(change)
	if err != nil {
		t.Fatalf("AnalyzeSymbolChange failed: %v", err)
	}

	// Should detect naming conflict
	if len(impact.ValidationErrors) == 0 {
		t.Error("Expected validation errors for naming conflict")
	}
}
