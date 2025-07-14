package commands_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hgsgtk/mygit/commands"
)

// TestInit tests the init command
func TestInit(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func() string
		expectedError  bool
		expectedOutput string
	}{
		{
			name: "successful initialization",
			setupFunc: func() string {
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				return tempDir
			},
			expectedError:  false,
			expectedOutput: "Repository initialized successfully",
		},
		{
			name: "already initialized",
			setupFunc: func() string {
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				// Create .mygit directory first
				os.Mkdir(commands.MyGitDir, 0755)
				metadata := map[string]any{}
				metadataPath := filepath.Join(commands.MyGitDir, commands.MetadataFile)
				file, _ := os.Create(metadataPath)
				defer file.Close()
				json.NewEncoder(file).Encode(metadata)
				return tempDir
			},
			expectedError:  false,
			expectedOutput: "Repository already initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := tt.setupFunc()
			defer os.RemoveAll(tempDir)

			err := commands.Init()

			if tt.expectedError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check if .mygit directory was created
			if !tt.expectedError {
				if _, err := os.Stat(commands.MyGitDir); os.IsNotExist(err) {
					t.Errorf(".mygit directory was not created")
				}

				// Check if metadata.json was created
				metadataPath := filepath.Join(commands.MyGitDir, commands.MetadataFile)
				if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
					t.Errorf("metadata.json was not created")
				}
			}
		})
	}
}

// TestAdd tests the add command
func TestAdd(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func() string
		files         []string
		expectedError bool
	}{
		{
			name: "add single file",
			setupFunc: func() string {
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				commands.Init()
				// Create test file
				os.WriteFile("test.txt", []byte("test content"), 0644)
				return tempDir
			},
			files:         []string{"test.txt"},
			expectedError: false,
		},
		{
			name: "add multiple files",
			setupFunc: func() string {
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				commands.Init()
				// Create test files
				os.WriteFile("file1.txt", []byte("content1"), 0644)
				os.WriteFile("file2.txt", []byte("content2"), 0644)
				return tempDir
			},
			files:         []string{"file1.txt", "file2.txt"},
			expectedError: false,
		},
		{
			name: "add non-existent file",
			setupFunc: func() string {
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				commands.Init()
				return tempDir
			},
			files:         []string{"nonexistent.txt"},
			expectedError: false, // Should not error, just warn
		},
		{
			name: "add without init",
			setupFunc: func() string {
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				os.WriteFile("test.txt", []byte("test content"), 0644)
				return tempDir
			},
			files:         []string{"test.txt"},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := tt.setupFunc()
			defer os.RemoveAll(tempDir)

			err := commands.Add(tt.files)

			if tt.expectedError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check if files were added to staging area
			if !tt.expectedError {
				metadataPath := filepath.Join(commands.MyGitDir, commands.MetadataFile)
				file, _ := os.Open(metadataPath)
				defer file.Close()
				var metadata map[string]any
				json.NewDecoder(file).Decode(&metadata)

				if stagingArea, ok := metadata["staging_area"]; ok {
					if arr, ok := stagingArea.([]any); ok {
						if len(arr) == 0 {
							t.Errorf("staging area is empty, expected files to be added")
						}
					}
				}
			}
		})
	}
}

// TestCommit tests the commit command
func TestCommit(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func() string
		message       string
		expectedError bool
	}{
		{
			name: "successful commit",
			setupFunc: func() string {
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				commands.Init()
				os.WriteFile("test.txt", []byte("test content"), 0644)
				commands.Add([]string{"test.txt"})
				return tempDir
			},
			message:       "Initial commit",
			expectedError: false,
		},
		{
			name: "commit without staged files",
			setupFunc: func() string {
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				commands.Init()
				return tempDir
			},
			message:       "Empty commit",
			expectedError: true,
		},
		{
			name: "commit without init",
			setupFunc: func() string {
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				return tempDir
			},
			message:       "Test commit",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := tt.setupFunc()
			defer os.RemoveAll(tempDir)

			err := commands.Commit(tt.message)

			if tt.expectedError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check if commit was added to history
			if !tt.expectedError {
				metadataPath := filepath.Join(commands.MyGitDir, commands.MetadataFile)
				file, _ := os.Open(metadataPath)
				defer file.Close()
				var metadata map[string]any
				json.NewDecoder(file).Decode(&metadata)

				if commitHistory, ok := metadata["commit_history"]; ok {
					if arr, ok := commitHistory.([]any); ok {
						if len(arr) == 0 {
							t.Errorf("commit history is empty, expected commit to be added")
						}
					}
				}

				// Check if staging area was cleared
				if stagingArea, ok := metadata["staging_area"]; ok {
					if arr, ok := stagingArea.([]any); ok {
						if len(arr) != 0 {
							t.Errorf("staging area was not cleared after commit")
						}
					}
				}
			}
		})
	}
}

// TestLog tests the log command
func TestLog(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func() string
		expectedError bool
	}{
		{
			name: "log with commits",
			setupFunc: func() string {
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				commands.Init()
				os.WriteFile("test.txt", []byte("test content"), 0644)
				commands.Add([]string{"test.txt"})
				commands.Commit("Initial commit")
				return tempDir
			},
			expectedError: false,
		},
		{
			name: "log without commits",
			setupFunc: func() string {
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				commands.Init()
				return tempDir
			},
			expectedError: false,
		},
		{
			name: "log without init",
			setupFunc: func() string {
				tempDir := t.TempDir()
				os.Chdir(tempDir)
				return tempDir
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := tt.setupFunc()
			defer os.RemoveAll(tempDir)

			err := commands.Log()

			if tt.expectedError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestMultipleCommits tests multiple commits and log order
func TestMultipleCommits(t *testing.T) {
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.RemoveAll(tempDir)

	// Initialize repository
	commands.Init()

	// Create and commit first file
	os.WriteFile("file1.txt", []byte("content1"), 0644)
	commands.Add([]string{"file1.txt"})
	commands.Commit("First commit")

	// Create and commit second file
	os.WriteFile("file2.txt", []byte("content2"), 0644)
	commands.Add([]string{"file2.txt"})
	commands.Commit("Second commit")

	// Verify commit history
	metadataPath := filepath.Join(commands.MyGitDir, commands.MetadataFile)
	file, _ := os.Open(metadataPath)
	defer file.Close()
	var metadata map[string]any
	json.NewDecoder(file).Decode(&metadata)

	if commitHistory, ok := metadata["commit_history"]; ok {
		if arr, ok := commitHistory.([]any); ok {
			if len(arr) != 2 {
				t.Errorf("expected 2 commits, got %d", len(arr))
			}

			// Check that commits are in chronological order (oldest first in array)
			firstCommit := arr[0].(map[string]any)
			secondCommit := arr[1].(map[string]any)

			if firstCommit["commit_message"] != "First commit" {
				t.Errorf("first commit message mismatch")
			}
			if secondCommit["commit_message"] != "Second commit" {
				t.Errorf("second commit message mismatch")
			}
		}
	}
}

// TestFileHashing tests that file hashing works correctly
func TestFileHashing(t *testing.T) {
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.RemoveAll(tempDir)

	commands.Init()

	// Create a file with known content
	content := "test content for hashing"
	os.WriteFile("test.txt", []byte(content), 0644)

	// Add the file
	commands.Add([]string{"test.txt"})

	// Read metadata to check hash
	metadataPath := filepath.Join(commands.MyGitDir, commands.MetadataFile)
	file, _ := os.Open(metadataPath)
	defer file.Close()
	var metadata map[string]any
	json.NewDecoder(file).Decode(&metadata)

	if stagingArea, ok := metadata["staging_area"]; ok {
		if arr, ok := stagingArea.([]any); ok {
			if len(arr) > 0 {
				if fileEntry, ok := arr[0].(map[string]any); ok {
					if hash, ok := fileEntry["file_hash"].(string); ok {
						if len(hash) != 40 { // SHA-1 is 40 characters
							t.Errorf("hash length is %d, expected 40", len(hash))
						}
					}
				}
			}
		}
	}
}