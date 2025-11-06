package api

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// T056: Golden file tests for Equal/Hash generation
// These tests ensure the generated code remains stable across changes

func TestGoldenItem_EqualGeneration(t *testing.T) {
	generatedFile := "oas_golden_item_equal_gen.go"
	goldenFile := filepath.Join("..", "golden", "golden_item_equal.golden")

	generated, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	golden, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
	}

	if !bytes.Equal(generated, golden) {
		t.Errorf("Generated Equal() code does not match golden file.\n"+
			"If this change is intentional, update the golden file:\n"+
			"  cp %s %s", generatedFile, goldenFile)
	}
}

func TestGoldenItem_HashGeneration(t *testing.T) {
	generatedFile := "oas_golden_item_hash_gen.go"
	goldenFile := filepath.Join("..", "golden", "golden_item_hash.golden")

	generated, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	golden, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
	}

	if !bytes.Equal(generated, golden) {
		t.Errorf("Generated Hash() code does not match golden file.\n"+
			"If this change is intentional, update the golden file:\n"+
			"  cp %s %s", generatedFile, goldenFile)
	}
}

func TestMetadata_EqualGeneration(t *testing.T) {
	generatedFile := "oas_metadata_equal_gen.go"
	goldenFile := filepath.Join("..", "golden", "metadata_equal.golden")

	generated, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	golden, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
	}

	if !bytes.Equal(generated, golden) {
		t.Errorf("Generated Equal() code does not match golden file.\n"+
			"If this change is intentional, update the golden file:\n"+
			"  cp %s %s", generatedFile, goldenFile)
	}
}

func TestMetadata_HashGeneration(t *testing.T) {
	generatedFile := "oas_metadata_hash_gen.go"
	goldenFile := filepath.Join("..", "golden", "metadata_hash.golden")

	generated, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	golden, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
	}

	if !bytes.Equal(generated, golden) {
		t.Errorf("Generated Hash() code does not match golden file.\n"+
			"If this change is intentional, update the golden file:\n"+
			"  cp %s %s", generatedFile, goldenFile)
	}
}
