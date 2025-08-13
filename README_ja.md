# system-prompt-gen

**日本語** | [English](./README.md)

複数のAIシステムプロンプトファイルを統合し、様々なAIツール（Claude、Cline、カスタムツール）用の統一されたプロンプトファイルを生成するGo製CLIツール。コマンドライン実行とインタラクティブTUIモードの両方に対応し、完全な国際化サポートを提供します。

## 特徴

- 🚀 `.system_prompt/*.md` ファイルを統合して各種AIツール用ファイルを生成
- 🎛️ インタラクティブなTUIモードとコマンドラインモードを選択可能
- 🌍 完全な国際化サポート（日本語・英語）
- ⚙️ TOML設定ファイルによる柔軟な設定管理
- 🔧 カスタムAIツールへの対応
- 🚫🔍 ツール別ファイル包含/除外パターン機能
- 🎨 Bubble Teaを使用した美しいTUI
- 🎯 自動ファイル検出・移行機能付きプロジェクト初期化

## インストール

### バイナリからインストール

Releasesページから最新版をダウンロードしてください。

### ソースからビルド

```bash
git clone https://github.com/yourusername/system-prompt-gen
cd system-prompt-gen
make build
```

## 使用方法

### 出力ファイル形式

生成されるファイルは以下の形式に従います：
```markdown
[settings.tomlのヘッダー内容]
# ファイル名（拡張子なし）

ファイル内容...

# 別のファイル名

別のファイル内容...

[settings.tomlのフッター内容]
```

例えば、`01-base.md` と `02-coding.md` ファイルがある場合、出力は以下のようになります：
```markdown
# 01-base
[01-base.mdの内容]

# 02-coding
[02-coding.mdの内容]
```

### 初回セットアップ

```bash
# 新しいプロジェクトを初期化（インタラクティブTUI）
system-prompt-gen init

# 実行内容：
# 1. .system_prompt/ ディレクトリの作成
# 2. 既存のAIツールファイル（CLAUDE.md、.clinerules等）のスキャン
# 3. ツール選択とファイル移行のガイド
# 4. 初期settings.toml設定の生成
```

### 基本的な使用方法

```bash
# カレントディレクトリの.system_prompt/を使用
system-prompt-gen

# 設定ファイルの場所を指定
system-prompt-gen -s /path/to/settings.toml

# インタラクティブモードで実行（デフォルト: true）
system-prompt-gen -i

# 非インタラクティブモードで実行（自動化/CI用）
system-prompt-gen -i=false

# 言語を指定
system-prompt-gen --language ja
system-prompt-gen -l en
```

### ディレクトリ構造

ツールは以下のディレクトリ構造を想定しています：

```txt
your-project/
├── .system_prompt/
│   ├── settings.toml      # 設定ファイル（オプション）
│   ├── 01-base.md         # プロンプトファイル
│   ├── 02-context.md      # プロンプトファイル
│   └── 03-rules.md        # プロンプトファイル
├── CLAUDE.md              # 生成されるファイル
└── .clinerules            # 生成されるファイル
```

## 設定ファイル

`.system_prompt/settings.toml` に設定ファイルを配置します：

```toml
# アプリケーション設定
[app]
header = "カスタムヘッダー内容"    # 全生成ファイルに追加するヘッダー（オプション）
footer = "カスタムフッター内容"    # 全生成ファイルに追加するフッター（オプション）

[tools.claude]
generate = true       # 生成を無効にするにはfalseに設定、デフォルトはtrue
dir_name = ""         # ディレクトリ名（空文字列 = カレントディレクトリ）
file_name = ""        # ファイル名（空文字列 = デフォルト: "CLAUDE.md"）
include = ["01-*.md", "02-*.md"]  # 特定パターンのみ包含（オプション、未定義なら全て包含）
exclude = ["003_*.md", "temp*.md"]  # ファイル除外パターン（excludeがincludeより優先）

[tools.cline]
generate = true
dir_name = ""
file_name = ""        # デフォルトは".clinerules"
include = ["*"]       # 全ファイルを包含（明示的指定）
exclude = ["01-*.md"]              # ツール固有の除外パターン

[tools.github_copilot]
generate = false      # GitHub Copilot用のビルトインサポート
dir_name = ".github"  # デフォルト: .github/copilot-instructions.md
file_name = "copilot-instructions.md"

[tools.aider]        # カスタムツール例: Aider AIコーディングアシスタント
generate = true
dir_name = ""         # カレントディレクトリ（カスタムツールに必須）
file_name = ".aider_prompt"  # カスタムツールに必須
include = ["01-*.md", "02-*.md"]  # 基本設定ファイルのみ包含

[tools.custom_tool]   # その他のカスタムAIツール
generate = true
dir_name = "./custom" # カスタムツールには必須
file_name = "custom.md"  # カスタムツールには必須
include = ["public_*.md", "common_*.md"]  # 公開ファイルと共通ファイルのみ包含
exclude = ["private*.md"]           # 機密ファイルをカスタムツールから除外
```

### 包含/除外パターン

各ツールは `.system_prompt/` からファイルをフィルタリングする `include` と `exclude` パターンを定義できます：

#### 包含パターン (Include)
- `include = ["pattern1", "pattern2"]` - これらのパターンに該当するファイルのみを包含
- 未定義の場合、デフォルトで全ファイルが包含される
- シェル形式のglobパターンを使用（`*`、`?`、`[...]`）
- パターンは `.system_prompt/` ディレクトリからの相対パスに対してマッチ
- 一般的なパターン例：`"01-*.md"`、`"public_*.md"`、`"*"`（全ファイル）

#### 除外パターン (Exclude)
- `exclude = ["pattern1", "pattern2"]` - これらのパターンに該当するファイルを除外
- **除外が優先** - includeとexclude両方に該当するファイルは除外される
- シェル形式のglobパターンを使用（`*`、`?`、`[...]`）
- 一般的なパターン例：`"003_*.md"`、`"temp*.md"`、`"private*.md"`、`"draft_*.md"`

#### 処理順序
1. `include` が定義されている場合、includeパターンに該当するファイルのみが考慮される
2. `include` が未定義の場合、全ファイルが考慮される
3. `exclude` パターンに該当するファイルが除去される（excludeが優先）
4. 各ツールは残ったファイルのみを処理

## 開発

### ビルドとテストコマンド

```bash
# プロジェクトをビルド
make build

# テストコマンド（開発用）
make test-unit      # tparse形式でユニットテストを実行（要: go install github.com/mfridman/tparse@latest）
make test-coverage  # カバレッジレポート付きでテスト実行（coverage.htmlを生成）
make test-verbose   # レース検出付きでテスト実行

# 統合テスト
make test          # サンプル設定での統合テスト
make interactive   # インタラクティブモードでの統合テスト

# ビルド成果物と生成ファイルのクリーンアップ
make clean

# システムPATHにインストール
make install

# ヘルプコマンド
make help          # 利用可能な全Makefileターゲットを表示
```

### アーキテクチャ

#### 設定システム

主にTOMLベースの設定を使用：

1. **TOML設定** (`.system_prompt/settings.toml`) - AIツール設定とアプリ設定のメイン設定
2. **JSON設定** (レガシー) - 後方互換性のためのみ維持

#### ジェネレーター処理フロー

`internal/generator/generator.go` がコアワークフローを制御：

1. 有効な各ツールに対して、`.system_prompt/*.md` ファイルを収集（ツール固有の包含/除外パターンを適用）
2. ファイル名のアルファベット順でソート
3. 設定されたヘッダー・フッターとコンテンツをマージ
4. TOML設定に基づいてツール固有の出力ファイルを生成

#### 国際化システム

`internal/i18n/i18n.go` が包括的なi18nサポートを提供：

- 埋め込みJSON翻訳ファイルと `github.com/nicksnyder/go-i18n/v2` を使用
- 言語検出：settings.toml → LANG環境変数 → フォールバック（ja → en）
- すべてのユーザー向けメッセージ（CLI、TUI、エラー）がローカライズ済み

## サポートするAIツール

### ビルトインツール
- **Claude** - Anthropic Claude用プロンプトファイル
- **Cline** - VS Code拡張のCline用ルールファイル
- **GitHub Copilot** - GitHub Copilot用指示ファイル

### カスタムツール
- **任意AIツール** - `dir_name` と `file_name` 設定でカスタムツールを定義
- **例**: Aider、Cursor、その他のAIツールをカスタムツールとして設定可能

## ライセンス

MIT License

## 貢献

プルリクエストやイシューを歓迎します。貢献前に既存のコードスタイルに従ってください。

## サポート

バグレポートや機能リクエストは [Issues](https://github.com/yourusername/system-prompt-gen/issues) でお願いします。
