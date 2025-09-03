# リポジトリガイドライン

## プロジェクト構成とモジュール構成

- **src**: `cmd/`（Cobra による CLI エントリ）、`internal/`（モジュール: `config`, `generator`, `i18n`, `ui`, `util`）
- **examples**: `example/`（統合テスト用の `.system_prompt/` と生成物を含む）
- **build**: `make build` で `.bin/system-prompt-gen` を作成
- **configs**: `.system_prompt/settings.toml`（ツールの出力、入出力ディレクトリの設定）

## ビルド・テスト・開発コマンド

- `make build`: `.bin/system-prompt-gen` をビルド
- `make run`: ビルド後、カレントディレクトリで実行
- `make test-unit`: `tparse` 形式でユニットテストを実行
- `make test`: ビルドし、`example/` で統合テストを実行
- `make interactive`: `example/` で TUI を起動
- `make clean`: ビルド生成物と出力ファイルを削除
- `make install`: バイナリを `~/bin` または `/usr/local/bin` にインストール

### ローカル実行例

```bash
system-prompt-gen -s ./.system_prompt/settings.toml -l en
system-prompt-gen -i  # インタラクティブモード（既定で有効）
```

## コーディングスタイルと命名規則

- **使用言語**: Go 1.24.x
- **フォーマット**: `gofmt`/`goimports`、CI で `golangci-lint` を実行
- **パッケージ名**: 小文字、アンダースコア禁止（例: `internal/generator`）
- **公開 API**: CamelCase とドキュメントコメント、テストは `*_test.go`
- **ファイル配置**: CLI は `cmd/`、ロジックは `internal/<package>`

## テストガイドライン

- **ツール**: `go test`、アサーションに `testify`、出力整形に `tparse`
- **実行**: ユニットは `make test-unit`、E2E 生成は `make test`
- **構成**: 対象パッケージをミラーし、テスト名は `TestXxx`
- **カバレッジ**: CI で収集、`generator`、`config`、`i18n` のコア経路を重点的に

## コミットとプルリクエストのガイドライン

- **コミット**: 短く命令形のサブジェクト（英日いずれも可）、課題参照例: `Fixes #123`
- **PR**: 概要・背景、必要に応じて生成物のスクリーンショットや差分（例: `CLAUDE.md`、`.clinerules`）、再現手順を含める
- **テスト**: 仕様変更にはテストを追加/調整し、`make test-unit` と `make test` を通過させる
- **ドキュメント**: 新しいフラグや挙動、例を追加したら `README.md` と `README_ja.md` を更新

## 国際化に関する注意

- **メッセージ**: `internal/i18n/locales/{en,ja}.json` に配置、ユーザー向け文言を追加したら両方を更新
- **言語**: フラグ/環境変数で自動検出、各ロケールで簡潔かつ一貫した文言に保つ
