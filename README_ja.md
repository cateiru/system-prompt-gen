# system-prompt-gen

**日本語** | [English](./README.md)

複数のAIシステムプロンプトファイルを統合し、様々なAIツール（Claude、Cline、カスタムツール）用の統一されたプロンプトファイルを生成するGo製CLIツール。コマンドライン実行とインタラクティブTUIモードの両方に対応し、完全な国際化サポートを提供します。

## 特徴

- 🚀 `.system_prompt/*.md` ファイルを統合して各種AIツール用ファイルを生成
- 🎛️ インタラクティブなTUIモードとコマンドラインモードを選択可能
- 🌍 完全な国際化サポート（日本語・英語）
- ⚙️ TOML設定ファイルによる柔軟な設定管理
- 🔧 カスタムAIツールへの対応
- 🎨 Bubble Teaを使用した美しいTUI

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

### 基本的な使用方法

```bash
# カレントディレクトリの.system_prompt/を使用
system-prompt-gen

# 設定ファイルの場所を指定
system-prompt-gen -s /path/to/settings.toml

# インタラクティブモードで実行
system-prompt-gen -i

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
# 言語設定は --language (-l) フラグで指定

[tools.claude]
generate = true       # 生成を無効にするにはfalseに設定、デフォルトはtrue
dir_name = ""         # ディレクトリ名（空文字列 = カレントディレクトリ）
file_name = ""        # ファイル名（空文字列 = デフォルト: "CLAUDE.md"）

[tools.cline]
generate = true
dir_name = ""
file_name = ""        # デフォルトは".clinerules"

[tools.github_copilot]
generate = false      # GitHub Copilot用のビルトインサポート
dir_name = ".github"  # デフォルト: .github/copilot-instructions.md
file_name = "copilot-instructions.md"

[tools.custom_tool]   # カスタムAIツールの追加
generate = true
dir_name = "./custom" # カスタムツールには必須
file_name = "custom.md"  # カスタムツールには必須
```

## 開発

### ビルドとテストコマンド

```bash
# プロジェクトをビルド
make build

# テストコマンド（開発用）
make test-unit      # ユニットテストを実行
make test-coverage  # カバレッジレポート付きでテスト実行
make test-verbose   # レース検出付きでテスト実行

# 統合テスト
make test          # サンプル設定での統合テスト
make interactive   # インタラクティブモードでの統合テスト

# ビルド成果物と生成ファイルのクリーンアップ
make clean

# システムPATHにインストール
make install
```

### アーキテクチャ

#### 設定システム

主にTOMLベースの設定を使用：

1. **TOML設定** (`.system_prompt/settings.toml`) - AIツール設定とアプリ設定のメイン設定
2. **JSON設定** (レガシー) - 後方互換性のためのみ維持

#### ジェネレーター処理フロー

`internal/generator/generator.go` がコアワークフローを制御：

1. `.system_prompt/*.md` ファイルをスキャン（設定の除外パターンを適用）
2. ファイル名のアルファベット順でソート
3. 設定されたヘッダー・フッターとコンテンツをマージ
4. TOML設定に基づいて複数のターゲットに出力

#### 国際化システム

`internal/i18n/i18n.go` が包括的なi18nサポートを提供：

- 埋め込みJSON翻訳ファイルと `github.com/nicksnyder/go-i18n/v2` を使用
- 言語検出：settings.toml → LANG環境変数 → フォールバック（ja → en）
- すべてのユーザー向けメッセージ（CLI、TUI、エラー）がローカライズ済み

## サポートするAIツール

- **Claude** - Anthropic Claude用プロンプトファイル
- **Cline** - VS Code拡張のCline用ルールファイル
- **GitHub Copilot** - GitHub Copilot用指示ファイル
- **カスタムツール** - 任意のAIツール用カスタムファイル

## ライセンス

MIT License

## 貢献

プルリクエストやイシューを歓迎します。貢献前に既存のコードスタイルに従ってください。

## サポート

バグレポートや機能リクエストは [Issues](https://github.com/yourusername/system-prompt-gen/issues) でお願いします。
