package init

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInitState(t *testing.T) {
	state, err := NewInitState()
	require.NoError(t, err)

	assert.NotEmpty(t, state.WorkDir)
	assert.Contains(t, state.SystemPromptDir, ".system_prompt")
	assert.Empty(t, state.ExistingFiles)
	assert.Empty(t, state.SelectedFiles)
	assert.Empty(t, state.SelectedTools)
	assert.True(t, state.OverwriteConfirmed)
}

func TestCheckSystemPromptDir(t *testing.T) {
	tempDir := t.TempDir()

	state := &InitState{
		WorkDir:         tempDir,
		SystemPromptDir: filepath.Join(tempDir, ".system_prompt"),
	}

	// ディレクトリが存在しない場合
	exists, err := state.CheckSystemPromptDir()
	require.NoError(t, err)
	assert.False(t, exists)

	// ディレクトリを作成
	err = os.MkdirAll(state.SystemPromptDir, 0755)
	require.NoError(t, err)

	// ディレクトリが存在する場合
	exists, err = state.CheckSystemPromptDir()
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestCreateSystemPromptDir(t *testing.T) {
	tempDir := t.TempDir()

	state := &InitState{
		WorkDir:         tempDir,
		SystemPromptDir: filepath.Join(tempDir, ".system_prompt"),
	}

	err := state.CreateSystemPromptDir()
	require.NoError(t, err)

	// ディレクトリが作成されたことを確認
	_, err = os.Stat(state.SystemPromptDir)
	assert.NoError(t, err)
}

func TestWriteDefaultFile(t *testing.T) {
	tempDir := t.TempDir()
	systemPromptDir := filepath.Join(tempDir, ".system_prompt")
	err := os.MkdirAll(systemPromptDir, 0755)
	require.NoError(t, err)

	state := &InitState{
		WorkDir:         tempDir,
		SystemPromptDir: systemPromptDir,
		SelectedFiles: []ExistingFile{
			{Path: "CLAUDE.md", ToolName: "claude", Content: "Claude prompt content"},
			{Path: ".clinerules", ToolName: "cline", Content: "Cline rules content"},
		},
	}

	err = state.WriteDefaultFile()
	require.NoError(t, err)

	// ファイルが作成されたことを確認
	filePath := filepath.Join(systemPromptDir, "001_default.md")
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "Claude prompt content")
	assert.Contains(t, contentStr, "Cline rules content")
	assert.Contains(t, contentStr, "# Imported from CLAUDE.md")
	assert.Contains(t, contentStr, "# Imported from .clinerules")
}

func TestWriteDefaultFileEmpty(t *testing.T) {
	tempDir := t.TempDir()
	systemPromptDir := filepath.Join(tempDir, ".system_prompt")
	err := os.MkdirAll(systemPromptDir, 0755)
	require.NoError(t, err)

	state := &InitState{
		WorkDir:         tempDir,
		SystemPromptDir: systemPromptDir,
		SelectedFiles:   []ExistingFile{}, // 空のファイルリスト
	}

	err = state.WriteDefaultFile()
	require.NoError(t, err)

	// 空のファイルが作成されたことを確認
	filePath := filepath.Join(systemPromptDir, "001_default.md")
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	assert.Empty(t, string(content))
}

func TestWriteSettingsFile(t *testing.T) {
	tempDir := t.TempDir()
	systemPromptDir := filepath.Join(tempDir, ".system_prompt")
	err := os.MkdirAll(systemPromptDir, 0755)
	require.NoError(t, err)

	state := &InitState{
		WorkDir:         tempDir,
		SystemPromptDir: systemPromptDir,
		SelectedTools:   []string{"claude", "cline"},
	}

	err = state.WriteSettingsFile()
	require.NoError(t, err)

	// ファイルが作成されたことを確認
	filePath := filepath.Join(systemPromptDir, "settings.toml")
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "[tools.claude]")
	assert.Contains(t, contentStr, "[tools.cline]")
	assert.Contains(t, contentStr, "[tools.github_copilot]")
	assert.Contains(t, contentStr, "generate = true")  // 選択されたツール
	assert.Contains(t, contentStr, "generate = false") // 選択されなかったツール
}

func TestGenerateSettingsContent(t *testing.T) {
	state := &InitState{
		SelectedTools: []string{"claude"},
	}

	content := state.generateSettingsContent()

	// 選択されたツールはgenerate = true
	assert.Contains(t, content, "[tools.claude]")
	assert.Regexp(t, `(?m)^\[tools\.claude\]$`, content)

	// 選択されなかったツールはgenerate = false
	assert.Contains(t, content, "[tools.cline]")
	assert.Contains(t, content, "[tools.github_copilot]")

	// claudeのみgenerate = true
	claudeSection := `[tools.claude]
generate = true`
	assert.Contains(t, content, claudeSection)

	clineSection := `[tools.cline]
generate = false`
	assert.Contains(t, content, clineSection)
}
