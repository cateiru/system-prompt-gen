package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToRelativePath(t *testing.T) {
	// 現在の作業ディレクトリを取得
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("現在のディレクトリの取得に失敗しました: %v", err)
	}

	tests := []struct {
		name     string
		fullPath string
		want     string
	}{
		{
			name:     "current directory file",
			fullPath: filepath.Join(currentDir, "test.txt"),
			want:     "test.txt",
		},
		{
			name:     "subdirectory file",
			fullPath: filepath.Join(currentDir, "sub", "test.txt"),
			want:     filepath.Join("sub", "test.txt"),
		},
		{
			name:     "nested subdirectory file",
			fullPath: filepath.Join(currentDir, "deep", "nested", "test.txt"),
			want:     filepath.Join("deep", "nested", "test.txt"),
		},
		{
			name:     "parent directory file",
			fullPath: filepath.Join(filepath.Dir(currentDir), "test.txt"),
			want:     filepath.Join("..", "test.txt"),
		},
		{
			name:     "current directory itself",
			fullPath: currentDir,
			want:     ".",
		},
		{
			name:     "relative path input (should work correctly)",
			fullPath: "already/relative/path.txt",
			want:     filepath.Join("already", "relative", "path.txt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToRelativePath(tt.fullPath)

			// デバッグ情報を追加
			t.Logf("Current dir: %s", currentDir)
			t.Logf("Input: %s", tt.fullPath)
			t.Logf("Got: %s", got)
			t.Logf("Want: %s", tt.want)

			// パスが絶対パスの場合は、相対パス変換を期待する
			if filepath.IsAbs(tt.fullPath) {
				expectedRel, err := filepath.Rel(currentDir, tt.fullPath)
				if err != nil {
					// エラーの場合はフォールバックとして元のパスが返されることを期待
					if got != tt.fullPath {
						t.Errorf("ToRelativePath() should fallback to original path on error, got %v, want %v", got, tt.fullPath)
					}
				} else {
					if got != expectedRel {
						t.Errorf("ToRelativePath() = %v, want %v", got, expectedRel)
					}
				}
			} else {
				// 相対パスの場合
				if got != tt.want {
					t.Errorf("ToRelativePath() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestToRelativePathFallback(t *testing.T) {
	// 無効なパスでフォールバック動作をテスト
	invalidPath := string([]byte{0x00}) // 無効な文字を含むパス
	result := ToRelativePath(invalidPath)

	// フォールバックとして元のパスが返されることを確認
	if result != invalidPath {
		t.Errorf("ToRelativePath() should fallback to original path for invalid input, got %v, want %v", result, invalidPath)
	}
}

func TestToRelativePathWithDotFiles(t *testing.T) {
	// プロジェクトのルートディレクトリを基準にテストを実行
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("現在のディレクトリの取得に失敗しました: %v", err)
	}

	// プロジェクトルートに移動
	projectRoot := filepath.Join(originalDir, "..", "..")
	err = os.Chdir(projectRoot)
	if err != nil {
		t.Fatalf("プロジェクトルートへの移動に失敗しました: %v", err)
	}

	// テスト後に元のディレクトリに戻る
	defer func() {
		err := os.Chdir(originalDir)
		require.NoError(t, err)
	}()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("プロジェクトルートディレクトリの取得に失敗しました: %v", err)
	}

	tests := []struct {
		name     string
		fullPath string
		want     string
	}{
		{
			name:     "dot file in current directory",
			fullPath: filepath.Join(currentDir, ".gitignore"),
			want:     ".gitignore",
		},
		{
			name:     "dot file in subdirectory",
			fullPath: filepath.Join(currentDir, ".github", "workflows", "test.yml"),
			want:     filepath.Join(".github", "workflows", "test.yml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToRelativePath(tt.fullPath)
			if got != tt.want {
				t.Errorf("ToRelativePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
