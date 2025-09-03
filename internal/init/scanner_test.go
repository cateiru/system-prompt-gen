package init

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileScanner_ScanExistingFiles(t *testing.T) {
	tempDir := t.TempDir()
	
	// テスト用ファイルを作成
	claudeFile := filepath.Join(tempDir, "CLAUDE.md")
	err := os.WriteFile(claudeFile, []byte("Claude prompt content"), 0644)
	require.NoError(t, err)
	
	clineFile := filepath.Join(tempDir, ".clinerules")
	err = os.WriteFile(clineFile, []byte("Cline rules content"), 0644)
	require.NoError(t, err)
	
	// GitHub Copilotディレクトリとファイルを作成
	githubDir := filepath.Join(tempDir, ".github")
	err = os.MkdirAll(githubDir, 0755)
	require.NoError(t, err)
	
	copilotFile := filepath.Join(githubDir, "copilot-instructions.md")
	err = os.WriteFile(copilotFile, []byte("Copilot instructions"), 0644)
	require.NoError(t, err)
	
	scanner := NewFileScanner(tempDir)
	files, err := scanner.ScanExistingFiles()
	require.NoError(t, err)
	
	// 3つのファイルが見つかることを確認
	assert.Len(t, files, 3)
	
	// ファイルの内容を確認
	fileMap := make(map[string]ExistingFile)
	for _, file := range files {
		fileMap[file.ToolName] = file
	}
	
	assert.Contains(t, fileMap, "claude")
	assert.Equal(t, "CLAUDE.md", fileMap["claude"].Path)
	assert.Equal(t, "Claude prompt content", fileMap["claude"].Content)
	
	assert.Contains(t, fileMap, "cline")
	assert.Equal(t, ".clinerules", fileMap["cline"].Path)
	assert.Equal(t, "Cline rules content", fileMap["cline"].Content)
	
	assert.Contains(t, fileMap, "github_copilot")
	assert.Equal(t, ".github/copilot-instructions.md", fileMap["github_copilot"].Path)
	assert.Equal(t, "Copilot instructions", fileMap["github_copilot"].Content)
}

func TestFileScanner_ScanExistingFiles_NoFiles(t *testing.T) {
	tempDir := t.TempDir()
	
	scanner := NewFileScanner(tempDir)
	files, err := scanner.ScanExistingFiles()
	require.NoError(t, err)
	
	// ファイルが見つからないことを確認
	assert.Empty(t, files)
}

func TestFileScanner_ScanExistingFiles_EmptyFiles(t *testing.T) {
	tempDir := t.TempDir()
	
	// 空のファイルを作成
	claudeFile := filepath.Join(tempDir, "CLAUDE.md")
	err := os.WriteFile(claudeFile, []byte(""), 0644)
	require.NoError(t, err)
	
	// 空白のみのファイルを作成
	clineFile := filepath.Join(tempDir, ".clinerules")
	err = os.WriteFile(clineFile, []byte("   \n\t  \n"), 0644)
	require.NoError(t, err)
	
	scanner := NewFileScanner(tempDir)
	files, err := scanner.ScanExistingFiles()
	require.NoError(t, err)
	
	// 空のファイルはスキップされることを確認
	assert.Empty(t, files)
}

func TestFileScanner_makeRelativePath(t *testing.T) {
	tempDir := t.TempDir()
	scanner := NewFileScanner(tempDir)
	
	// 同じディレクトリのファイル
	fullPath := filepath.Join(tempDir, "CLAUDE.md")
	relPath := scanner.makeRelativePath(fullPath)
	assert.Equal(t, "CLAUDE.md", relPath)
	
	// サブディレクトリのファイル
	fullPath = filepath.Join(tempDir, ".github", "copilot-instructions.md")
	relPath = scanner.makeRelativePath(fullPath)
	assert.Equal(t, ".github/copilot-instructions.md", relPath)
}

func TestFileScanner_findToolFile_Directory(t *testing.T) {
	tempDir := t.TempDir()
	
	// ディレクトリを作成（ファイルではない）
	claudeDir := filepath.Join(tempDir, "CLAUDE.md")
	err := os.MkdirAll(claudeDir, 0755)
	require.NoError(t, err)
	
	scanner := NewFileScanner(tempDir)
	
	// ディレクトリは検出されないことを確認
	files, err := scanner.ScanExistingFiles()
	require.NoError(t, err)
	assert.Empty(t, files)
}