BINARY_NAME=system-prompt-gen
BUILD_DIR=.bin

.PHONY: build clean test run example test-unit test-coverage test-verbose

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
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "カバレッジレポートがcoverage.htmlに生成されました"

# Run tests in verbose mode
test-verbose:
	go test -v -race -count=1 ./...

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
	@echo "  test-unit    - ユニットテストを実行"
	@echo "  test-coverage- カバレッジ付きでテストを実行"
	@echo "  test-verbose - 詳細モードでテストを実行"
	@echo "  test         - exampleディレクトリで統合テスト実行"
	@echo "  interactive  - exampleディレクトリでインタラクティブモード実行"
	@echo "  run          - カレントディレクトリで実行"
	@echo "  install      - バイナリを~/binまたは/usr/local/binにインストール"