package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cateiru/system-prompt-gen/internal/config"
	"github.com/cateiru/system-prompt-gen/internal/generator"
	"github.com/cateiru/system-prompt-gen/internal/i18n"
	"github.com/cateiru/system-prompt-gen/internal/ui"
)

var (
	configFile      string
	interactiveMode bool
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
	homeDir, _ := os.UserHomeDir()
	defaultConfigPath := filepath.Join(homeDir, ".config", "system-prompt-gen", "config.json")

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", defaultConfigPath, "Path to configuration file")
	rootCmd.PersistentFlags().BoolVarP(&interactiveMode, "interactive", "i", false, "Launch in interactive mode")
}

func runWithCmd(cmd *cobra.Command) error {
	return runWithCmdAndSettings(cmd, true)
}

func runWithCmdAndSettings(cmd *cobra.Command, useSettings bool) error {
	var cfg *config.Config
	var err error

	if useSettings {
		// settings.tomlの読み込みを試行
		settingsPath := filepath.Join(".", ".system_prompt", "settings.toml")
		cfg, err = config.LoadConfigWithSettings(configFile, settingsPath)
	} else {
		// レガシーモード（configのみ）
		cfg, err = config.LoadConfig(configFile)
	}

	if err != nil {
		return fmt.Errorf("%s", i18n.T("config_load_error", map[string]interface{}{"Error": err}))
	}

	// i18nシステムの初期化
	var language string
	if cfg.Settings != nil {
		language = cfg.Settings.App.Language
	}
	if err := i18n.Initialize(language); err != nil {
		// i18n初期化に失敗した場合でも処理を続行
		fmt.Fprintf(os.Stderr, "Warning: Failed to initialize i18n: %v\n", err)
	}

	// i18n初期化後にコマンドの説明を更新（NOTE: 実行時に行う）

	if interactiveMode {
		return ui.RunInteractive(cfg)
	}

	gen := generator.New(cfg)
	if err := gen.Run(); err != nil {
		return err
	}

	files, _ := gen.CollectPromptFiles()
	cmd.Printf("%s\n", i18n.T("files_processed", map[string]interface{}{"Count": len(files)}))

	// 生成されたファイルの一覧を表示
	targets := gen.GetGeneratedTargets()
	for _, target := range targets {
		cmd.Printf("%s\n", i18n.T("file_generated", map[string]interface{}{"FileName": target}))
	}

	return nil
}

func run() error {
	return runWithCmd(rootCmd)
}
