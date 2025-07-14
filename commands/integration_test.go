package commands_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hgsgtk/mygit/commands"
)

// TestFullWorkflow tests the complete workflow from init to log
func TestFullWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	os.Chdir(tempDir)

	// Step 1: Initialize repository
	err := commands.Init()
	if err != nil {
		t.Fatalf("Failed to initialize repository: %v", err)
	}

	// Step 2: Create some test files
	testFiles := map[string]string{
		"file1.txt": "Content of file 1",
		"file2.txt": "Content of file 2",
		"file3.md":  "# Markdown file\n\nSome content",
	}

	for filename, content := range testFiles {
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Step 3: Add files to staging area
	err = commands.Add([]string{"file1.txt", "file2.txt"})
	if err != nil {
		t.Fatalf("Failed to add files: %v", err)
	}

	// Step 4: Commit the changes
	err = commands.Commit("Initial commit with two files")
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Step 5: Add another file
	err = commands.Add([]string{"file3.md"})
	if err != nil {
		t.Fatalf("Failed to add third file: %v", err)
	}

	// Step 6: Commit the second change
	err = commands.Commit("Add markdown file")
	if err != nil {
		t.Fatalf("Failed to commit second change: %v", err)
	}

	// Step 7: Check log output
	err = commands.Log()
	if err != nil {
		t.Fatalf("Failed to show log: %v", err)
	}

	// Verify that .mygit directory and metadata.json exist
	if _, err := os.Stat(commands.MyGitDir); os.IsNotExist(err) {
		t.Error(".mygit directory does not exist")
	}

	metadataPath := filepath.Join(commands.MyGitDir, commands.MetadataFile)
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Error("metadata.json does not exist")
	}
}

// TestErrorHandling tests various error conditions
func TestErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	os.Chdir(tempDir)

	// Test commit without init
	err := commands.Commit("test commit")
	if err == nil {
		t.Error("Expected error when committing without init, got none")
	}

	// Test add without init
	err = commands.Add([]string{"test.txt"})
	if err == nil {
		t.Error("Expected error when adding without init, got none")
	}

	// Test log without init
	err = commands.Log()
	if err == nil {
		t.Error("Expected error when logging without init, got none")
	}

	// Initialize repository
	commands.Init()

	// Test commit without staged files
	err = commands.Commit("empty commit")
	if err == nil {
		t.Error("Expected error when committing without staged files, got none")
	}
}

// TestFilePatterns tests various file patterns in add command
func TestFilePatterns(t *testing.T) {
	tempDir := t.TempDir()
	os.Chdir(tempDir)

	commands.Init()

	// Create test files
	testFiles := map[string]string{
		"file1.txt": "Content 1",
		"file2.txt": "Content 2",
		"file3.md":  "Markdown content",
		"file4.go":  "Go code",
	}

	for filename, content := range testFiles {
		os.WriteFile(filename, []byte(content), 0644)
	}

	// Test glob pattern
	err := commands.Add([]string{"*.txt"})
	if err != nil {
		t.Fatalf("Failed to add files with glob pattern: %v", err)
	}

	// Test multiple patterns
	err = commands.Add([]string{"*.md", "*.go"})
	if err != nil {
		t.Fatalf("Failed to add files with multiple patterns: %v", err)
	}
} 