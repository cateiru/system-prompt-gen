BINARY_NAME=system-prompt-gen
BUILD_DIR=.bin

.PHONY: build clean test run example

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

clean:
	@rm -rf $(BUILD_DIR)
	@rm -f CLAUDE.md .clinerules
	@rm -f example/CLAUDE.md example/.clinerules
	@rm -f example/config/*.md example/config/.clinerules
	@rm -f example/cursor_rules.md example/.aider_prompt

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
	@echo "  build      - プログラムをビルド"
	@echo "  clean      - ビルドファイルと生成ファイルを削除"
	@echo "  test       - exampleディレクトリでテスト実行"
	@echo "  interactive- exampleディレクトリでインタラクティブモード実行"
	@echo "  run        - カレントディレクトリで実行"
	@echo "  install    - バイナリを~/binまたは/usr/local/binにインストール"