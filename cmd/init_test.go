package cmd

import (
	"testing"

	"github.com/cateiru/system-prompt-gen/internal/i18n"
	"github.com/stretchr/testify/assert"
)

func TestInitCommandHelp(t *testing.T) {
	// i18nを初期化
	err := i18n.Initialize("en")
	assert.NoError(t, err)

	// initコマンドの存在を確認
	assert.NotNil(t, initCmd)
	
	// コマンドの基本情報を確認
	assert.Equal(t, "init", initCmd.Use)
	assert.NotEmpty(t, initCmd.Short)
	assert.NotEmpty(t, initCmd.Long)
}

func TestRunInitWithoutTTY(t *testing.T) {
	// i18nを初期化
	err := i18n.Initialize("en")
	assert.NoError(t, err)

	// 非TTY環境（テスト環境）でrunInitを実行
	err = runInit(initCmd)
	
	// TTYエラーが発生することを確認（具体的なメッセージ内容はi18nに依存するため、エラーの存在のみ確認）
	assert.Error(t, err)
	assert.NotEmpty(t, err.Error())
}