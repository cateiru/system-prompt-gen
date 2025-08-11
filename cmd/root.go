package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cateiru/system-prompt-gen/internal/config"
	"github.com/cateiru/system-prompt-gen/internal/generator"
	"github.com/cateiru/system-prompt-gen/internal/ui"
)

var (
	configFile      string
	interactiveMode bool
)

var rootCmd = &cobra.Command{
	Use:   "system-prompt-gen",
	Short: "システムプロンプトファイルを統合するツール",
	Long: `system-prompt-gen は .system_prompt/*.md ファイルを収集し、
CLAUDE.md や .clinerules などの単一ファイルに統合するツールです。`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(); err != nil {
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

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", defaultConfigPath, "設定ファイルのパス")
	rootCmd.PersistentFlags().BoolVarP(&interactiveMode, "interactive", "i", false, "インタラクティブモードで起動")
}

func run() error {
	// settings.tomlの読み込みを試行
	settingsPath := filepath.Join(".", ".system_prompt", "settings.toml")
	cfg, err := config.LoadConfigWithSettings(configFile, settingsPath)
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	if interactiveMode {
		return ui.RunInteractive(cfg)
	}

	gen := generator.New(cfg)
	if err := gen.Run(); err != nil {
		return err
	}

	files, _ := gen.CollectPromptFiles()
	fmt.Printf("✅ %d個のプロンプトファイルを統合しました\n", len(files))

	// 生成されたファイルの一覧を表示
	targets := gen.GetGeneratedTargets()
	for _, target := range targets {
		fmt.Printf("📄 %s を生成しました\n", target)
	}

	return nil
}
