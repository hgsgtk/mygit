# Git-like Version Control System (CLI Tool)

A simple CLI tool that mimics a tiny subset of Git functionality, providing basic version control operations.

## 🔧 Features

- `init` - Initialize a new repository
- `add` - Add files to staging area  
- `commit` - Commit changes to repository
- `log` - Show commit history

## 🚀 Quick Start

```bash
# Initialize a new repository
./mygit init

# Create some files
touch file1.txt
touch file2.txt

# Add files to staging area
./mygit add file1.txt
./mygit add file2.txt

# Commit the changes
./mygit commit -m "Initial commit"

# View commit history
./mygit log
```

## 📋 Requirements

- [x] Initialize repo (`.mygit/`)
- [x] Track files in a staging area
- [x] Commit snapshot to disk
- [x] Show commit logs with timestamps
- [x] Store metadata as JSON
- [x] Hash contents (SHA-1)
- [x] Design simple commit object structure
- [x] Implement CLI with argparse or Click

## 🎯 Bonus Features (Planned)

- [ ] Branching support
- [ ] Diffs between commits
- [ ] Undo last commit
- [ ] Pattern matching in subdirectories for add command
- [ ] File deletion support

## 📖 CLI Commands

### `init` - Initialize Repository
```bash
./mygit init
```
- **Input**: None
- **Output**: Success or failure message
- **Description**: Initialize a new repository
- **Implementation**:
  - Create `.mygit` folder in current directory
  - Create empty `metadata.json` file
  - If `.mygit` already exists, do nothing

### `add` - Add Files to Staging Area
```bash
./mygit add <file_path>
./mygit add <pattern>  # e.g., *.txt
```
- **Input**: File path or pattern
- **Output**: Success or failure message
- **Description**: Add files to staging area
- **Implementation**:
  - If directory: add all files in directory
  - If file: add single file
  - If pattern: add all matching files
  - Update staging area with file paths and SHA-1 hashes

### `commit` - Commit Changes
```bash
./mygit commit -m "Commit message"
```
- **Input**: Commit message
- **Output**: Success or failure message
- **Description**: Commit staged changes to repository
- **Implementation**:
  - Create commit object with metadata
  - Store commit in repository
  - Clear staging area

### `log` - Show Commit History
```bash
./mygit log
```
- **Input**: None
- **Output**: Commit history
- **Description**: Display commit history
- **Implementation**:
  - Show commits in reverse chronological order
  - Display commit ID, message, and timestamp
  - Show "No commits yet" if empty

## 🏗️ Data Structure Design

### Repository Structure
```
.mygit/
├── metadata.json      # Repository metadata and commit history
└── staging.json       # Staging area information
```

### Metadata Format
When no commits exist:
```json
{}
```

With commit history:
```json
{
    "commit_history": [
        {
            "commit_id": "1234567890",
            "commit_message": "Initial commit",
            "commit_timestamp": "2021-01-01 00:00:00",
            "files": [
                {
                    "file_path": "file1.txt",
                    "file_hash": "abc123..."
                }
            ],
            "parent_commit_id": null
        }
    ]
}
```

### Staging Area Format
```json
{
    "staging_area": [
        {
            "file_path": "file1.txt",
            "file_hash": "1234567890"
        },
        {
            "file_path": "file2.txt", 
            "file_hash": "0987654321"
        }
    ]
}
```

### Commit Object Structure
Each commit contains:
- `commit_id` - SHA-1 hash of commit object
- `commit_message` - User-provided commit message
- `commit_timestamp` - Timestamp of commit
- `files` - List of files with paths and hashes
- `parent_commit_id` - SHA-1 of parent commit (null for first commit)

## 🔄 Implementation Status

- [x] CLI boilerplate
- [x] `init` command implementation
- [x] `add` command implementation  
- [x] `commit` command implementation
- [x] `log` command implementation
- [x] Write tests and refactor code

## 🧪 Testing

Run tests with:
```bash
go test ./...
```

## 📝 License

This project is for educational purposes to understand version control system concepts.

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
