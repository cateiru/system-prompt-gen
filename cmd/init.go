package cmd

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/cateiru/system-prompt-gen/internal/i18n"
	initpkg "github.com/cateiru/system-prompt-gen/internal/init"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize project setup",
	Long:  "system-prompt-gen init creates a .system_prompt folder in the current directory,\ndetects existing system prompt files, and initializes the project.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runInit(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command) error {
	// i18nシステムの初期化
	if err := i18n.Initialize(language); err != nil {
		// i18n初期化に失敗した場合でも処理を続行
		fmt.Fprintf(os.Stderr, "Warning: Failed to initialize i18n: %v\n", err)
	}

	// TTY検証 - initコマンドはインタラクティブモードのみサポート
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return fmt.Errorf("%s", i18n.T("init_requires_tty"))
	}

	// init処理を実行
	return initpkg.RunInit()
}