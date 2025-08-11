# CLAUDE.md

このファイルは、このリポジトリで作業する際のClaude Code (claude.ai/code) への指針を提供します。

## プロジェクト概要

system-prompt-genは、`.system_prompt/*.md`にある複数のAIシステムプロンプトファイルを、様々なAIツール（Claude、Cline、カスタムツール）用の統合されたプロンプトファイルに集約するGo製CLIツールです。コマンドライン実行とインタラクティブなTUIモードの両方を提供します。

## ビルド・開発コマンド

```bash
# プロジェクトをビルド
make build
# または
go build -o .bin/system-prompt-gen .

# サンプル設定でテスト
make test
# または
cd example && ../.bin/system-prompt-gen

# インタラクティブモードでテスト
make interactive
# または
cd example && ../.bin/system-prompt-gen -i

# ビルド成果物と生成ファイルをクリーンアップ
make clean

# システムPATHにインストール
make install
```

## アーキテクチャとコアコンポーネント

### 設定システム（二層構造）
ツールは二層の設定システムを使用します：
1. **JSON設定** (`~/.config/system-prompt-gen/config.json`) - レガシーなグローバル設定
2. **TOML設定** (`.system_prompt/settings.toml`) - プロジェクト毎のAIツール設定

`internal/config/config.go`の重要な型：
- `Config`: 後方互換性を持つメイン設定
- `Settings`: TOMLベースのツール別設定
- `AIToolSettings`: 個別ツール設定（生成フラグ、パス、ファイル名）

### Generator処理フロー
`internal/generator/generator.go`がコアワークフローを制御：
1. `.system_prompt/*.md`ファイルをスキャン（設定の除外パターンを適用）
2. ファイル名でアルファベット順にソート
3. 設定されたヘッダー・フッターとコンテンツをマージ
4. TOML設定に基づいて複数のターゲットに出力

### 設定読み込み優先順位
システムは以下の順序で設定を読み込みます：
1. `LoadConfigWithSettings()` でJSON設定とTOML設定の両方を試行
2. `LoadConfig()` にフォールバック（JSON単体、後方互換性）
3. TOMLファイルが存在しない場合は `DefaultSettings()` を使用
4. TOMLの設定がJSON設定の出力動作を上書き

### 出力ターゲット解決
TOML設定が存在する場合、generatorは：
- 各AIツールの `generate` フラグをチェック
- パスを解決（空文字列 = カレントディレクトリ）
- 必要に応じてディレクトリを作成
- `[custom.toolname]` セクションでカスタムツールをサポート

### インタラクティブUI
`internal/ui/tui.go` は3つの状態を持つBubble Tea TUIを提供：
- Loading: ファイル収集フェーズ
- Success: ファイル数とターゲットのプレビュー表示
- Error: エラー表示と再試行オプション

## settings.toml設定

作業ディレクトリに `.system_prompt/settings.toml` を配置：

```toml
[claude]
generate = true
path = ""           # デフォルトはカレントディレクトリ
file_name = ""      # デフォルトは "CLAUDE.md"

[cline] 
generate = true
path = ""
file_name = ""      # デフォルトは ".clinerules"

[custom.toolname]   # カスタムAIツールを追加
generate = true
path = "./custom"   # カスタムツールの場合は必須
file_name = "custom.md"  # カスタムツールの場合は必須
```

## CLI使用パターン

```bash
# 基本的な使用（カレントディレクトリの .system_prompt/ を使用）
system-prompt-gen

# カスタム設定ファイルの場所を指定
system-prompt-gen -c /path/to/config.json

# プレビューと確認のためのインタラクティブモード
system-prompt-gen -i

# ツールは以下を含む .system_prompt/ ディレクトリを期待：
# - *.md ファイル（プロンプトの断片）
# - settings.toml（オプション、ツール固有の設定）
```