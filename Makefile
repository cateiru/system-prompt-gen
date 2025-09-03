BINARY_NAME=system-prompt-gen
BUILD_DIR=.bin

.PHONY: build clean test run example test-unit test-coverage test-verbose test-pretty lint

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

clean:
	@rm -rf $(BUILD_DIR)
	@rm -f CLAUDE.md .clinerules
	@rm -f example/CLAUDE.md example/.clinerules
	@rm -f example/config/*.md example/config/.clinerules
	@rm -f example/cursor_rules.md example/.aider_prompt
	@rm -f coverage.out coverage.html
	@rm -rf testdata/output

# Run unit tests
test-unit:
	@which tparse > /dev/null || (echo "tparseがインストールされていません。'go install github.com/mfridman/tparse@latest'を実行してください" && exit 1)
	@set -o pipefail && go test -json ./... | tparse

# Run tests with coverage report
test-coverage:
	@which tparse > /dev/null || (echo "tparseがインストールされていません。'go install github.com/mfridman/tparse@latest'を実行してください" && exit 1)
	@set -o pipefail && go test -json -coverprofile=coverage.out ./... | tparse
	@go tool cover -html=coverage.out -o coverage.html
	@echo "カバレッジレポートが coverage.html に生成されました"

# Run tests with race detection
test-verbose:
	@which tparse > /dev/null || (echo "tparseがインストールされていません。'go install github.com/mfridman/tparse@latest'を実行してください" && exit 1)
	@set -o pipefail && go test -json -race -v ./... | tparse

# Run lint checks
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lintがインストールされていません。'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'を実行してください" && exit 1)
	golangci-lint run --timeout=10m

# Run integration test with example
test: build
	@cd example && ../$(BUILD_DIR)/$(BINARY_NAME)

interactive: build
	@cd example && ../$(BUILD_DIR)/$(BINARY_NAME) -i

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

install: build
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/bin/$(BINARY_NAME) || cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

help:
	@echo "使用可能なターゲット:"
	@echo "  build        - プログラムをビルド"
	@echo "  clean        - ビルドファイルと生成ファイルを削除"
	@echo "  lint         - golangci-lintでコードを静的解析"
	@echo "  test-unit    - ユニットテストを実行"
	@echo "  test         - exampleディレクトリで統合テスト実行"
	@echo "  interactive  - exampleディレクトリでインタラクティブモード実行"
	@echo "  run          - カレントディレクトリで実行"
	@echo "  install      - バイナリを~/binまたは/usr/local/binにインストール"
