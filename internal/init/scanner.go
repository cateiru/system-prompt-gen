package init

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cateiru/system-prompt-gen/internal/config"
)

// FileScanner は既存のシステムプロンプトファイルをスキャンする
type FileScanner struct {
	WorkDir string
}

// NewFileScanner は新しい FileScanner を作成する
func NewFileScanner(workDir string) *FileScanner {
	return &FileScanner{
		WorkDir: workDir,
	}
}

// ScanExistingFiles は既存のシステムプロンプトファイルをスキャンする
func (scanner *FileScanner) ScanExistingFiles() ([]ExistingFile, error) {
	var files []ExistingFile

	for toolName, paths := range config.DefaultKnownToolFileNames {
		file, found, err := scanner.findToolFile(toolName, paths)
		if err != nil {
			return nil, fmt.Errorf("error scanning %s files: %w", toolName, err)
		}
		if found {
			files = append(files, file)
		}
	}

	return files, nil
}

func (scanner *FileScanner) findToolFile(toolName string, paths config.AIToolPaths) (ExistingFile, bool, error) {
	// パスを構築
	var filePath string
	if string(paths.DirName) == "" {
		filePath = filepath.Join(scanner.WorkDir, string(paths.FileName))
	} else {
		filePath = filepath.Join(scanner.WorkDir, string(paths.DirName), string(paths.FileName))
	}

	// ファイルの存在を確認
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return ExistingFile{}, false, nil
	}
	if err != nil {
		return ExistingFile{}, false, err
	}

	// ディレクトリの場合はスキップ
	if info.IsDir() {
		return ExistingFile{}, false, nil
	}

	// ファイルの内容を読み取り
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ExistingFile{}, false, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// 空のファイルはスキップ
	contentStr := strings.TrimSpace(string(content))
	if contentStr == "" {
		return ExistingFile{}, false, nil
	}

	file := ExistingFile{
		Path:     scanner.makeRelativePath(filePath),
		ToolName: toolName,
		Content:  contentStr,
	}

	return file, true, nil
}

func (scanner *FileScanner) makeRelativePath(fullPath string) string {
	relPath, err := filepath.Rel(scanner.WorkDir, fullPath)
	if err != nil {
		// 相対パス作成に失敗した場合は絶対パスを返す
		return fullPath
	}
	return relPath
}