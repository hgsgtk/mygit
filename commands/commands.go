package commands

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	MyGitDir = ".mygit"
	MetadataFile = "metadata.json"
)

// Init initializes a new repository
func Init() error {
	// Check if .mygit directory already exists
	if _, err := os.Stat(MyGitDir); err == nil {
		fmt.Println("Repository already initialized")
		return nil
	}

	// Create .mygit directory
	if err := os.Mkdir(MyGitDir, 0755); err != nil {
		return fmt.Errorf("failed to create .mygit directory: %w", err)
	}

	// Create metadata.json with empty JSON object
	metadata := map[string]any{}
	metadataPath := filepath.Join(MyGitDir, MetadataFile)
	
	file, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to create metadata.json: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(metadata); err != nil {
		return fmt.Errorf("failed to write metadata.json: %w", err)
	}

	fmt.Println("Repository initialized successfully")
	return nil
}

// Add adds files to the staging area
func Add(args []string) error {
	// Check if .mygit exists
	if _, err := os.Stat(MyGitDir); os.IsNotExist(err) {
		return errors.New("not a mygit repository (run 'mygit init' first)")
	}

	// Expand all arguments to file paths
	var filesToAdd []string
	for _, arg := range args {
		matches, err := filepath.Glob(arg)
		if err != nil || matches == nil {
			// If not a glob, check if it's a file or directory
			info, statErr := os.Stat(arg)
			if statErr == nil {
				if info.IsDir() {
					filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
						if err == nil && !info.IsDir() {
							filesToAdd = append(filesToAdd, path)
						}
						return nil
					})
				} else {
					filesToAdd = append(filesToAdd, arg)
				}
			}
			continue
		}
		for _, match := range matches {
			info, statErr := os.Stat(match)
			if statErr == nil {
				if info.IsDir() {
					filepath.Walk(match, func(path string, info os.FileInfo, err error) error {
						if err == nil && !info.IsDir() {
							filesToAdd = append(filesToAdd, path)
						}
						return nil
					})
				} else {
					filesToAdd = append(filesToAdd, match)
				}
			}
		}
	}

	if len(filesToAdd) == 0 {
		fmt.Println("No files to add.")
		return nil
	}

	// Remove duplicates
	fileSet := make(map[string]struct{})
	for _, f := range filesToAdd {
		fileSet[f] = struct{}{}
	}
	uniqueFiles := make([]string, 0, len(fileSet))
	for f := range fileSet {
		uniqueFiles = append(uniqueFiles, f)
	}

	// Read metadata.json
	metadataPath := filepath.Join(MyGitDir, MetadataFile)
	metadata := map[string]any{}
	if file, err := os.Open(metadataPath); err == nil {
		defer file.Close()
		json.NewDecoder(file).Decode(&metadata)
	}

	// Load or create staging_area
	stagingArea := []map[string]string{}
	if sa, ok := metadata["staging_area"]; ok {
		if arr, ok := sa.([]any); ok {
			for _, v := range arr {
				if m, ok := v.(map[string]any); ok {
					entry := map[string]string{}
					for k, val := range m {
						if s, ok := val.(string); ok {
							entry[k] = s
						}
					}
					stagingArea = append(stagingArea, entry)
				}
			}
		}
	}

	// Index for quick lookup
	stagedIndex := make(map[string]int)
	for i, entry := range stagingArea {
		stagedIndex[entry["file_path"]] = i
	}

	// Add/update files
	for _, filePath := range uniqueFiles {
		// Skip files in .mygit
		if strings.HasPrefix(filePath, MyGitDir+string(os.PathSeparator)) || filePath == MyGitDir {
			continue
		}
		f, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not open %s: %v\n", filePath, err)
			continue
		}
		sha := sha1.New()
		if _, err := io.Copy(sha, f); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not hash %s: %v\n", filePath, err)
			f.Close()
			continue
		}
		f.Close()
		hash := fmt.Sprintf("%x", sha.Sum(nil))

		entry := map[string]string{"file_path": filePath, "file_hash": hash}
		if idx, ok := stagedIndex[filePath]; ok {
			stagingArea[idx] = entry
			fmt.Printf("Updated: %s\n", filePath)
		} else {
			stagingArea = append(stagingArea, entry)
			fmt.Printf("Added: %s\n", filePath)
		}
	}

	metadata["staging_area"] = stagingArea

	// Write back to metadata.json
	file, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to update metadata.json: %w", err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(metadata); err != nil {
		return fmt.Errorf("failed to write metadata.json: %w", err)
	}

	return nil
}

// Commit commits the staged changes
func Commit(message string) error {
	// Check if .mygit exists
	if _, err := os.Stat(MyGitDir); os.IsNotExist(err) {
		return errors.New("not a mygit repository (run 'mygit init' first)")
	}

	// Read metadata.json
	metadataPath := filepath.Join(MyGitDir, MetadataFile)
	metadata := map[string]any{}
	if file, err := os.Open(metadataPath); err == nil {
		defer file.Close()
		json.NewDecoder(file).Decode(&metadata)
	}

	// Load staging area
	stagingArea := []map[string]string{}
	if sa, ok := metadata["staging_area"]; ok {
		if arr, ok := sa.([]any); ok {
			for _, v := range arr {
				if m, ok := v.(map[string]any); ok {
					entry := map[string]string{}
					for k, val := range m {
						if s, ok := val.(string); ok {
							entry[k] = s
						}
					}
					stagingArea = append(stagingArea, entry)
				}
			}
		}
	}

	// Check if there are files in staging area
	if len(stagingArea) == 0 {
		return errors.New("no files staged for commit")
	}

	// Get current timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Get parent commit ID
	parentCommitID := ""
	if commitHistory, ok := metadata["commit_history"]; ok {
		if arr, ok := commitHistory.([]any); ok && len(arr) > 0 {
			if lastCommit, ok := arr[len(arr)-1].(map[string]any); ok {
				if id, ok := lastCommit["commit_id"].(string); ok {
					parentCommitID = id
				}
			}
		}
	}

	// Create commit content for hashing
	commitContent := fmt.Sprintf("%s%s%s", timestamp, message, parentCommitID)
	for _, file := range stagingArea {
		commitContent += file["file_path"] + file["file_hash"]
	}

	// Generate commit ID
	sha := sha1.New()
	sha.Write([]byte(commitContent))
	commitID := fmt.Sprintf("%x", sha.Sum(nil))

	// Create commit object
	commit := map[string]any{
		"commit_id":        commitID,
		"commit_message":   message,
		"commit_timestamp": timestamp,
		"files":            stagingArea,
		"parent_commit_id": parentCommitID,
	}

	// Add to commit history
	var commitHistory []any
	if ch, ok := metadata["commit_history"]; ok {
		if arr, ok := ch.([]any); ok {
			commitHistory = arr
		}
	}
	commitHistory = append(commitHistory, commit)
	metadata["commit_history"] = commitHistory

	// Clear staging area
	metadata["staging_area"] = []any{}

	// Write back to metadata.json
	file, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to update metadata.json: %w", err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(metadata); err != nil {
		return fmt.Errorf("failed to write metadata.json: %w", err)
	}

	// Print success message
	fmt.Printf("Committed %d files\n", len(stagingArea))
	fmt.Printf("Commit ID: %s\n", commitID)
	fmt.Printf("Message: %s\n", message)

	return nil
}

// Log shows the commit history
func Log() error {
	// Check if .mygit exists
	if _, err := os.Stat(MyGitDir); os.IsNotExist(err) {
		return errors.New("not a mygit repository (run 'mygit init' first)")
	}

	// Read metadata.json
	metadataPath := filepath.Join(MyGitDir, MetadataFile)
	metadata := map[string]any{}
	if file, err := os.Open(metadataPath); err == nil {
		defer file.Close()
		json.NewDecoder(file).Decode(&metadata)
	}

	// Load commit history
	var commitHistory []any
	if ch, ok := metadata["commit_history"]; ok {
		if arr, ok := ch.([]any); ok {
			commitHistory = arr
		}
	}

	// Check if there are any commits
	if len(commitHistory) == 0 {
		fmt.Println("No commits yet")
		return nil
	}

	// Display commits in reverse chronological order (newest first)
	for i := len(commitHistory) - 1; i >= 0; i-- {
		commit := commitHistory[i]
		if commitMap, ok := commit.(map[string]any); ok {
			// Extract commit information
			commitID, _ := commitMap["commit_id"].(string)
			commitMessage, _ := commitMap["commit_message"].(string)
			commitTimestamp, _ := commitMap["commit_timestamp"].(string)

			// Display commit
			fmt.Printf("commit %s\n", commitID)
			fmt.Printf("Date: %s\n", commitTimestamp)
			fmt.Println()
			fmt.Printf("    %s\n", commitMessage)
			fmt.Println()
		}
	}

	return nil
} 