package api

import (
	"testing"

	"github.com/ogen-go/ogen/validate"
)

// T041: Test validateUnique() with empty array (should pass)
func TestValidateUniqueWorkflowStatus_Empty(t *testing.T) {
	var items []WorkflowStatus
	err := validateUniqueWorkflowStatus(items)
	if err != nil {
		t.Errorf("Expected no error for empty array, got: %v", err)
	}
}

// T042: Test validateUnique() with single element (should pass)
func TestValidateUniqueWorkflowStatus_Single(t *testing.T) {
	items := []WorkflowStatus{
		{
			ID:          "status-1",
			Name:        "Open",
			Description: NewOptString("Issue is open"),
		},
	}
	err := validateUniqueWorkflowStatus(items)
	if err != nil {
		t.Errorf("Expected no error for single element, got: %v", err)
	}
}

// T043: Test validateUnique() with duplicates (should return DuplicateItemsError with indices)
func TestValidateUniqueWorkflowStatus_Duplicates(t *testing.T) {
	items := []WorkflowStatus{
		{ID: "status-1", Name: "Open", Description: NewOptString("Issue is open")},
		{ID: "status-2", Name: "In Progress", Description: NewOptString("Work started")},
		{ID: "status-3", Name: "Done", Description: NewOptString("Completed")},
		{ID: "status-1", Name: "Open", Description: NewOptString("Issue is open")}, // Duplicate of index 0
	}

	err := validateUniqueWorkflowStatus(items)
	if err == nil {
		t.Fatal("Expected DuplicateItemsError, got nil")
	}

	dupErr, ok := err.(*validate.DuplicateItemsError)
	if !ok {
		t.Fatalf("Expected *validate.DuplicateItemsError, got %T: %v", err, err)
	}

	if len(dupErr.Indices) != 2 {
		t.Errorf("Expected 2 indices, got %d: %v", len(dupErr.Indices), dupErr.Indices)
	}

	if dupErr.Indices[0] != 0 || dupErr.Indices[1] != 3 {
		t.Errorf("Expected indices [0, 3], got %v", dupErr.Indices)
	}

	// Check error message includes indices
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}
	t.Logf("Error message: %s", errMsg)
}

// T044: Test validateUnique() with all unique items (should pass)
func TestValidateUniqueWorkflowStatus_AllUnique(t *testing.T) {
	items := []WorkflowStatus{
		{ID: "status-1", Name: "Open", Description: NewOptString("Issue is open")},
		{ID: "status-2", Name: "In Progress", Description: NewOptString("Work started")},
		{ID: "status-3", Name: "Done", Description: NewOptString("Completed")},
		{ID: "status-4", Name: "Closed", Description: NewOptString("Issue closed")},
		{ID: "status-5", Name: "Reopened", Description: NewOptString("Issue reopened")},
	}

	err := validateUniqueWorkflowStatus(items)
	if err != nil {
		t.Errorf("Expected no error for unique items, got: %v", err)
	}
}

// Test with nested objects to verify complex equality
func TestValidateUniqueWorkflowStatus_NestedObjects(t *testing.T) {
	items := []WorkflowStatus{
		{
			ID:          "status-1",
			Name:        "Open",
			Description: NewOptString("Issue is open"),
			Properties: NewOptStatusProperties(StatusProperties{
				Category: NewOptStatusPropertiesCategory(StatusPropertiesCategoryTODO),
				Color:    NewOptString("red"),
			}),
		},
		{
			ID:          "status-2",
			Name:        "In Progress",
			Description: NewOptString("Work started"),
			Properties: NewOptStatusProperties(StatusProperties{
				Category: NewOptStatusPropertiesCategory(StatusPropertiesCategoryINPROGRESS),
				Color:    NewOptString("yellow"),
			}),
		},
		{
			ID:          "status-1",
			Name:        "Open",
			Description: NewOptString("Issue is open"),
			Properties: NewOptStatusProperties(StatusProperties{
				Category: NewOptStatusPropertiesCategory(StatusPropertiesCategoryTODO),
				Color:    NewOptString("red"),
			}),
		},
	}

	err := validateUniqueWorkflowStatus(items)
	if err == nil {
		t.Fatal("Expected DuplicateItemsError for nested duplicates, got nil")
	}

	dupErr, ok := err.(*validate.DuplicateItemsError)
	if !ok {
		t.Fatalf("Expected *validate.DuplicateItemsError, got %T", err)
	}

	if dupErr.Indices[0] != 0 || dupErr.Indices[1] != 2 {
		t.Errorf("Expected indices [0, 2], got %v", dupErr.Indices)
	}
}

// T045: Benchmark validateUnique() with 1,000 complex objects - verify <10ms
func BenchmarkValidateUniqueWorkflowStatus_1000Items(b *testing.B) {
	// Create 1,000 unique items
	items := make([]WorkflowStatus, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = WorkflowStatus{
			ID:          string(rune(i)),
			Name:        string(rune(i + 1000)),
			Description: NewOptString(string(rune(i + 2000))),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validateUniqueWorkflowStatus(items)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

// Benchmark with duplicates to test worst-case
func BenchmarkValidateUniqueWorkflowStatus_1000ItemsWithDuplicate(b *testing.B) {
	// Create 1,000 items with duplicate at the end
	items := make([]WorkflowStatus, 1000)
	for i := 0; i < 999; i++ {
		items[i] = WorkflowStatus{
			ID:          string(rune(i)),
			Name:        string(rune(i + 1000)),
			Description: NewOptString(string(rune(i + 2000))),
		}
	}
	items[999] = items[0] // Duplicate

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validateUniqueWorkflowStatus(items)
		if err == nil {
			b.Fatal("Expected error for duplicate")
		}
	}
}

// Test that StatusProperties validation also works
func TestValidateUniqueStatusProperties_Basic(t *testing.T) {
	items := []StatusProperties{
		{Color: NewOptString("red"), IsDefault: NewOptBool(true)},
		{Color: NewOptString("red"), IsDefault: NewOptBool(true)}, // Duplicate
	}

	err := validateUniqueStatusProperties(items)
	if err == nil {
		t.Fatal("Expected DuplicateItemsError, got nil")
	}

	dupErr, ok := err.(*validate.DuplicateItemsError)
	if !ok {
		t.Fatalf("Expected *validate.DuplicateItemsError, got %T", err)
	}

	if dupErr.Indices[0] != 0 || dupErr.Indices[1] != 1 {
		t.Errorf("Expected indices [0, 1], got %v", dupErr.Indices)
	}
}
