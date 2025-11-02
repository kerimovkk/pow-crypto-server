package quotes

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	quotesFile := filepath.Join(tmpDir, "test_quotes.txt")

	content := `# This is a comment
The only true wisdom is in knowing you know nothing.

# Another comment
The journey of a thousand miles begins with one step.
That which does not kill us makes us stronger.

`
	if err := os.WriteFile(quotesFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test loading
	m := NewManager()
	if err := m.LoadFromFile(quotesFile); err != nil {
		t.Fatalf("LoadFromFile() failed: %v", err)
	}

	// Should have 3 quotes (comments and empty lines ignored)
	if count := m.Count(); count != 3 {
		t.Errorf("Expected 3 quotes, got %d", count)
	}
}

func TestLoadFromFile_NonExistent(t *testing.T) {
	m := NewManager()
	err := m.LoadFromFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestLoadFromFile_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	quotesFile := filepath.Join(tmpDir, "empty_quotes.txt")

	// File with only comments and empty lines
	content := `# Comment
# Another comment

`
	if err := os.WriteFile(quotesFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	m := NewManager()
	err := m.LoadFromFile(quotesFile)
	if err == nil {
		t.Error("Expected error for empty file, got nil")
	}
}

func TestGetRandom(t *testing.T) {
	m := NewManager()
	
	// Test with empty manager
	_, err := m.GetRandom()
	if err == nil {
		t.Error("Expected error for empty manager, got nil")
	}
	
	// Load quotes
	tmpDir := t.TempDir()
	quotesFile := filepath.Join(tmpDir, "quotes.txt")
	content := `Quote 1
Quote 2
Quote 3`
	if err := os.WriteFile(quotesFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	if err := m.LoadFromFile(quotesFile); err != nil {
		t.Fatalf("LoadFromFile() failed: %v", err)
	}
	
	// Get multiple random quotes
	seen := make(map[string]bool)
	for i := 0; i < 10; i++ {
		quote, err := m.GetRandom()
		if err != nil {
			t.Fatalf("GetRandom() failed: %v", err)
		}
		
		if quote == "" {
			t.Error("GetRandom() returned empty quote")
		}
		
		seen[quote] = true
	}
	
	t.Logf("Got %d unique quotes out of 10 attempts", len(seen))
}
