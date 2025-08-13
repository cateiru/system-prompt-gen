package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/cateiru/system-prompt-gen/internal/config"
	"github.com/cateiru/system-prompt-gen/internal/generator"
	"github.com/cateiru/system-prompt-gen/internal/i18n"
	"github.com/cateiru/system-prompt-gen/internal/ui"
	"github.com/cateiru/system-prompt-gen/internal/util"
)

var (
	settingFile     string
	interactiveMode bool
	language        string
)

var rootCmd = &cobra.Command{
	Use:   "system-prompt-gen",
	Short: "Tool to integrate system prompt files",
	Long:  "system-prompt-gen collects .system_prompt/*.md files and integrates them into single files like CLAUDE.md and .clinerules.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runWithCmd(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	defaultSettingFullPath := path.Join(currentDir, ".system_prompt", "settings.toml")

	rootCmd.PersistentFlags().StringVarP(&settingFile, "setting", "s", defaultSettingFullPath, "Path to settings.toml config file")
	rootCmd.PersistentFlags().BoolVarP(&interactiveMode, "interactive", "i", true, "Launch in interactive mode")
	rootCmd.PersistentFlags().StringVarP(&language, "language", "l", "", "Language setting (ja, en, or empty for auto-detect)")
}

func runWithCmd(cmd *cobra.Command) error {
	// i18nシステムの初期化
	if err := i18n.Initialize(language); err != nil {
		// i18n初期化に失敗した場合でも処理を続行
		fmt.Fprintf(os.Stderr, "Warning: Failed to initialize i18n: %v\n", err)
	}

	// settings.tomlの読み込みを試行
	settings, err := config.LoadSettings(settingFile)
	if err != nil {
		return fmt.Errorf("%s", i18n.T("config_load_error", map[string]any{"Error": err.Error()}))
	}

	// i18n初期化後にコマンドの説明を更新（NOTE: 実行時に行う）

	// TTY検出による自動フォールバック
	effectiveInteractiveMode := interactiveMode
	if interactiveMode && !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		effectiveInteractiveMode = false
		cmd.Printf("%s\n", i18n.T("tty_fallback_message"))
	}

	if effectiveInteractiveMode {
		return ui.RunInteractive(settings)
	}

	gen := generator.New(settings)
	if err := gen.Run(); err != nil {
		return err
	}

	files, _ := gen.CollectPromptFiles()
	cmd.Printf("%s\n", i18n.T("files_processed", map[string]any{"Count": len(files)}))

	// 生成されたファイルの一覧を表示
	targets := gen.GetGeneratedTargets()
	for _, target := range targets {
		cmd.Printf("%s\n", i18n.T("file_generated", map[string]any{"FileName": util.ToRelativePath(target)}))
	}

	return nil
}
