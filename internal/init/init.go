package init

import (
	"fmt"
	"os"
	"path/filepath"
)

// InitState は初期化プロセスの状態を管理する
type InitState struct {
	WorkDir            string
	SystemPromptDir    string
	ExistingFiles      []ExistingFile
	SelectedFiles      []ExistingFile
	SelectedTools      []string
	OverwriteConfirmed bool
}

// ExistingFile は既存のシステムプロンプトファイルを表す
type ExistingFile struct {
	Path     string
	ToolName string
	Content  string
}

// NewInitState は新しい InitState を作成する
func NewInitState() (*InitState, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	return &InitState{
		WorkDir:            workDir,
		SystemPromptDir:    filepath.Join(workDir, ".system_prompt"),
		ExistingFiles:      []ExistingFile{},
		SelectedFiles:      []ExistingFile{},
		SelectedTools:      []string{},
		OverwriteConfirmed: true,
	}, nil
}

// CheckSystemPromptDir は .system_prompt ディレクトリの存在を確認する
func (state *InitState) CheckSystemPromptDir() (bool, error) {
	_, err := os.Stat(state.SystemPromptDir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check .system_prompt directory: %w", err)
	}
	return true, nil
}

// CreateSystemPromptDir は .system_prompt ディレクトリを作成する
func (state *InitState) CreateSystemPromptDir() error {
	if err := os.MkdirAll(state.SystemPromptDir, 0755); err != nil {
		return fmt.Errorf("failed to create .system_prompt directory: %w", err)
	}
	return nil
}

// WriteDefaultFile は選択されたファイルの内容を 001_default.md に書き込む
func (state *InitState) WriteDefaultFile() error {
	if len(state.SelectedFiles) == 0 {
		// 選択されたファイルがない場合は空のファイルを作成
		return state.writeFile("001_default.md", "")
	}

	var content string
	for i, file := range state.SelectedFiles {
		if i > 0 {
			content += "\n\n"
		}
		content += fmt.Sprintf("# Imported from %s\n\n%s", file.Path, file.Content)
	}

	return state.writeFile("001_default.md", content)
}

// WriteSettingsFile は settings.toml ファイルを生成する
func (state *InitState) WriteSettingsFile() error {
	content := state.generateSettingsContent()
	return state.writeFile("settings.toml", content)
}

func (state *InitState) writeFile(filename, content string) error {
	filePath := filepath.Join(state.SystemPromptDir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}
	return nil
}

func (state *InitState) generateSettingsContent() string {
	content := "[app]\n"
	content += "# header = \"Custom header content\"\n"
	content += "# footer = \"Custom footer content\"\n\n"

	// 選択されたツールのみ generate = true にする
	allTools := []string{"claude", "cline", "github_copilot"}
	selectedToolsMap := make(map[string]bool)
	for _, tool := range state.SelectedTools {
		selectedToolsMap[tool] = true
	}

	for _, tool := range allTools {
		generate := selectedToolsMap[tool]
		content += fmt.Sprintf("[tools.%s]\n", tool)
		content += fmt.Sprintf("generate = %t\n", generate)
		content += "dir_name = \"\"\n"
		content += "file_name = \"\"\n"
		if tool == "github_copilot" && generate {
			content += "# dir_name = \".github\"  # GitHub Copilot uses .github directory\n"
			content += "# file_name = \"copilot-instructions.md\"\n"
		}
		content += "# exclude = [\"temp*.md\"]\n\n"
	}

	return content
}

// RunInit はinit処理のエントリーポイント
func RunInit() error {
	state, err := NewInitState()
	if err != nil {
		return err
	}

	// 既存の .system_prompt ディレクトリをチェック
	exists, err := state.CheckSystemPromptDir()
	if err != nil {
		return err
	}

	if exists {
		state.OverwriteConfirmed = false
	}

	// 既存ファイルをスキャン
	if err := state.scanExistingFiles(); err != nil {
		return err
	}

	// インタラクティブUIを起動
	return runInteractiveInit(state)
}

func (state *InitState) scanExistingFiles() error {
	scanner := NewFileScanner(state.WorkDir)
	files, err := scanner.ScanExistingFiles()
	if err != nil {
		return err
	}
	state.ExistingFiles = files
	return nil
}
