package api

import (
	"testing"

	"github.com/ogen-go/ogen/validate"
)

// T053: Integration test combining all field types
func TestValidateUniqueComprehensiveItem_AllFieldTypes(t *testing.T) {
	// Create two different items with all field types populated
	item1 := ComprehensiveItem{
		// Required primitives
		ID:       "item-1",
		Name:     "First Item",
		Priority: 1,
		Active:   true,

		// Optional primitives
		Description: NewOptString("First description"),
		Score:       NewOptFloat64(95.5),

		// Nullable
		ExternalId: NewOptNilString("ext-001"),

		// Enum
		Status: NewOptComprehensiveItemStatus(ComprehensiveItemStatusDraft),

		// Optional enum
		Category: NewOptComprehensiveItemCategory(ComprehensiveItemCategoryFeature),

		// Array of primitives
		Tags: []string{"important", "urgent"},

		// Optional array
		Labels: []string{"frontend", "backend"},

		// Map
		Metadata: NewOptComprehensiveItemMetadata(map[string]string{
			"project": "alpha",
			"team":    "core",
		}),

		// Optional map
		CustomFields: NewOptComprehensiveItemCustomFields(map[string]string{
			"field1": "value1",
		}),

		// Nested object
		Owner: NewOptUser(User{
			Username: "alice",
			Email:    NewOptString("alice@example.com"),
			FullName: NewOptString("Alice Smith"),
			Verified: NewOptBool(true),
		}),

		// Optional nested object
		Assignee: NewOptUser(User{
			Username: "bob",
			Email:    NewOptString("bob@example.com"),
		}),

		// Array of nested objects
		Watchers: []User{
			{Username: "charlie"},
			{Username: "diana"},
		},

		// Complex nested structure
		Configuration: NewOptConfiguration(Configuration{
			Version: NewOptString("1.0"),
			Settings: NewOptConfigurationSettings(map[string]string{
				"theme": "dark",
			}),
			Features: []Feature{
				{
					Name:    NewOptString("experimental"),
					Enabled: NewOptBool(true),
				},
			},
		}),
	}

	item2 := ComprehensiveItem{
		// Different values for all fields
		ID:       "item-2",
		Name:     "Second Item",
		Priority: 2,
		Active:   false,

		Description: NewOptString("Second description"),
		Score:       NewOptFloat64(87.3),

		ExternalId: NewOptNilString("ext-002"),

		Status: NewOptComprehensiveItemStatus(ComprehensiveItemStatusPublished),

		Category: NewOptComprehensiveItemCategory(ComprehensiveItemCategoryBug),

		Tags: []string{"review", "testing"},

		Labels: []string{"api"},

		Metadata: NewOptComprehensiveItemMetadata(map[string]string{
			"project": "beta",
		}),

		CustomFields: NewOptComprehensiveItemCustomFields(map[string]string{
			"field2": "value2",
		}),

		Owner: NewOptUser(User{
			Username: "eve",
			Email:    NewOptString("eve@example.com"),
		}),

		Assignee: NewOptUser(User{
			Username: "frank",
		}),

		Watchers: []User{
			{Username: "grace"},
		},

		Configuration: NewOptConfiguration(Configuration{
			Version: NewOptString("2.0"),
		}),
	}

	items := []ComprehensiveItem{item1, item2}

	// Should pass - items are different
	err := validateUniqueComprehensiveItem(items)
	if err != nil {
		t.Errorf("Expected no error for different items, got: %v", err)
	}
}

// Test duplicate detection with all field types
func TestValidateUniqueComprehensiveItem_DuplicateDetection(t *testing.T) {
	item := ComprehensiveItem{
		ID:       "item-1",
		Name:     "Test Item",
		Priority: 1,
		Active:   true,

		Description: NewOptString("Test description"),
		Score:       NewOptFloat64(100.0),

		Tags: []string{"tag1", "tag2"},

		Metadata: NewOptComprehensiveItemMetadata(map[string]string{
			"key": "value",
		}),

		Owner: NewOptUser(User{
			Username: "user1",
			Email:    NewOptString("user1@example.com"),
		}),

		Watchers: []User{
			{Username: "watcher1"},
		},

		Configuration: NewOptConfiguration(Configuration{
			Version: NewOptString("1.0"),
			Features: []Feature{
				{Name: NewOptString("feat1"), Enabled: NewOptBool(true)},
			},
		}),
	}

	items := []ComprehensiveItem{item, item}

	err := validateUniqueComprehensiveItem(items)
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

// Test that items differing only in optional fields are detected as different
func TestValidateUniqueComprehensiveItem_OptionalFieldDifferences(t *testing.T) {
	baseItem := ComprehensiveItem{
		ID:       "item-1",
		Name:     "Base Item",
		Priority: 1,
		Active:   true,
		Owner: NewOptUser(User{
			Username: "owner",
		}),
	}

	// Item with optional field set
	itemWithOptional := ComprehensiveItem{
		ID:          "item-1",
		Name:        "Base Item",
		Priority:    1,
		Active:      true,
		Description: NewOptString("Has description"),
		Owner: NewOptUser(User{
			Username: "owner",
		}),
	}

	items := []ComprehensiveItem{baseItem, itemWithOptional}

	err := validateUniqueComprehensiveItem(items)
	if err != nil {
		t.Errorf("Expected no error for items with different optional fields, got: %v", err)
	}
}

// Test that items differing only in nested objects are detected as different
func TestValidateUniqueComprehensiveItem_NestedObjectDifferences(t *testing.T) {
	item1 := ComprehensiveItem{
		ID:       "item-1",
		Name:     "Item",
		Priority: 1,
		Active:   true,
		Owner: NewOptUser(User{
			Username: "alice",
			Email:    NewOptString("alice@example.com"),
		}),
	}

	item2 := ComprehensiveItem{
		ID:       "item-1",
		Name:     "Item",
		Priority: 1,
		Active:   true,
		Owner: NewOptUser(User{
			Username: "bob", // Different owner
			Email:    NewOptString("bob@example.com"),
		}),
	}

	items := []ComprehensiveItem{item1, item2}

	err := validateUniqueComprehensiveItem(items)
	if err != nil {
		t.Errorf("Expected no error for items with different nested objects, got: %v", err)
	}
}

// Test that items differing only in arrays are detected as different
func TestValidateUniqueComprehensiveItem_ArrayDifferences(t *testing.T) {
	item1 := ComprehensiveItem{
		ID:       "item-1",
		Name:     "Item",
		Priority: 1,
		Active:   true,
		Tags:     []string{"tag1", "tag2"},
		Owner: NewOptUser(User{
			Username: "owner",
		}),
	}

	item2 := ComprehensiveItem{
		ID:       "item-1",
		Name:     "Item",
		Priority: 1,
		Active:   true,
		Tags:     []string{"tag1", "tag3"}, // Different tags
		Owner: NewOptUser(User{
			Username: "owner",
		}),
	}

	items := []ComprehensiveItem{item1, item2}

	err := validateUniqueComprehensiveItem(items)
	if err != nil {
		t.Errorf("Expected no error for items with different arrays, got: %v", err)
	}
}

// Test that items differing only in maps are detected as different
func TestValidateUniqueComprehensiveItem_MapDifferences(t *testing.T) {
	item1 := ComprehensiveItem{
		ID:       "item-1",
		Name:     "Item",
		Priority: 1,
		Active:   true,
		Metadata: NewOptComprehensiveItemMetadata(map[string]string{
			"key1": "value1",
		}),
		Owner: NewOptUser(User{
			Username: "owner",
		}),
	}

	item2 := ComprehensiveItem{
		ID:       "item-1",
		Name:     "Item",
		Priority: 1,
		Active:   true,
		Metadata: NewOptComprehensiveItemMetadata(map[string]string{
			"key1": "value2", // Different value
		}),
		Owner: NewOptUser(User{
			Username: "owner",
		}),
	}

	items := []ComprehensiveItem{item1, item2}

	err := validateUniqueComprehensiveItem(items)
	if err != nil {
		t.Errorf("Expected no error for items with different maps, got: %v", err)
	}
}

// Test hash and equality consistency
func TestComprehensiveItem_HashEqualConsistency(t *testing.T) {
	item1 := ComprehensiveItem{
		ID:       "item-1",
		Name:     "Test",
		Priority: 1,
		Active:   true,
		Owner: NewOptUser(User{
			Username: "user",
		}),
	}

	item2 := ComprehensiveItem{
		ID:       "item-1",
		Name:     "Test",
		Priority: 1,
		Active:   true,
		Owner: NewOptUser(User{
			Username: "user",
		}),
	}

	item3 := ComprehensiveItem{
		ID:       "item-2",
		Name:     "Test",
		Priority: 1,
		Active:   true,
		Owner: NewOptUser(User{
			Username: "user",
		}),
	}

	// Equal items must have equal hashes
	if item1.Hash() != item2.Hash() {
		t.Error("Equal items should have equal hashes")
	}

	// Equal items must return true from Equal()
	if !item1.Equal(item2, 0) {
		t.Error("Equal items should return true from Equal()")
	}

	// Different items should return false from Equal()
	if item1.Equal(item3, 0) {
		t.Error("Different items should return false from Equal()")
	}
}

// Test with empty/unset optional fields
func TestValidateUniqueComprehensiveItem_UnsetOptionalFields(t *testing.T) {
	// Minimal items with only required fields
	item1 := ComprehensiveItem{
		ID:       "item-1",
		Name:     "Minimal 1",
		Priority: 1,
		Active:   true,
		Owner: NewOptUser(User{
			Username: "owner1",
		}),
	}

	item2 := ComprehensiveItem{
		ID:       "item-2",
		Name:     "Minimal 2",
		Priority: 2,
		Active:   false,
		Owner: NewOptUser(User{
			Username: "owner2",
		}),
	}

	items := []ComprehensiveItem{item1, item2}

	err := validateUniqueComprehensiveItem(items)
	if err != nil {
		t.Errorf("Expected no error for minimal items, got: %v", err)
	}

	// Test duplicate minimal items
	items = []ComprehensiveItem{item1, item1}
	err = validateUniqueComprehensiveItem(items)
	if err == nil {
		t.Fatal("Expected DuplicateItemsError for duplicate minimal items")
	}
}

// Benchmark with all field types
func BenchmarkValidateUniqueComprehensiveItem_100Items(b *testing.B) {
	items := make([]ComprehensiveItem, 100)
	for i := 0; i < 100; i++ {
		items[i] = ComprehensiveItem{
			ID:       string(rune('A' + (i % 26))),
			Name:     string(rune('a' + (i % 26))),
			Priority: i,
			Active:   i%2 == 0,
			Tags:     []string{string(rune('0' + (i % 10)))},
			Metadata: NewOptComprehensiveItemMetadata(map[string]string{
				"index": string(rune('0' + (i % 10))),
			}),
			Owner: NewOptUser(User{
				Username: string(rune('A' + (i % 26))),
			}),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validateUniqueComprehensiveItem(items)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
