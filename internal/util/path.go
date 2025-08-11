package util

import (
	"os"
	"path/filepath"
)

// ToRelativePath converts a full path to a relative path from the current directory.
// If the conversion fails, it returns the original full path as a fallback.
func ToRelativePath(fullPath string) string {
	// 現在の作業ディレクトリを取得
	currentDir, err := os.Getwd()
	if err != nil {
		// フォールバック: 現在のディレクトリ取得に失敗した場合は元のパスを返す
		return fullPath
	}

	relPath, err := filepath.Rel(currentDir, fullPath)
	if err != nil {
		// フォールバック: 相対パス変換に失敗した場合はフルパスを返す
		return fullPath
	}
	return relPath
}
