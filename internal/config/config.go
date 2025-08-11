package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type AIToolSettings struct {
	Generate bool   `toml:"generate"`
	Path     string `toml:"path"`
	FileName string `toml:"file_name"`
}

type AppSettings struct {
	Language string `toml:"language"`
}

type Settings struct {
	App    AppSettings               `toml:"app"`
	Claude AIToolSettings            `toml:"claude"`
	Cline  AIToolSettings            `toml:"cline"`
	Custom map[string]AIToolSettings `toml:"custom"`
}

type Config struct {
	InputDir     string    `json:"inputDir"`
	OutputFiles  []string  `json:"outputFiles"`
	ExcludeFiles []string  `json:"excludeFiles"`
	Header       string    `json:"header"`
	Footer       string    `json:"footer"`
	Settings     *Settings `json:"-"`
}

// DefaultSettings はアプリケーションの設定 (Settings) のデフォルト値を返します。
// これには App/Claude/Cline/Custom の初期値が含まれます。
func DefaultSettings() *Settings {
	return &Settings{
		App: AppSettings{
			Language: "",
		},
		Claude: AIToolSettings{
			Generate: true,
			Path:     "",
			FileName: "CLAUDE.md",
		},
		Cline: AIToolSettings{
			Generate: true,
			Path:     "",
			FileName: ".clinerules",
		},
		Custom: make(map[string]AIToolSettings),
	}
}

func DefaultConfig() *Config {
	return &Config{
		InputDir:     ".system_prompt",
		OutputFiles:  []string{"CLAUDE.md", ".clinerules"},
		ExcludeFiles: []string{},
		Header:       "# System Prompt\n\n",
		Footer:       "",
		Settings:     nil, // レガシーモードを維持するため
	}
}

// LoadSettings は指定された TOML ファイル (settingsPath) から設定を読み込みます。
// ファイルが存在しない場合はデフォルト設定を返します。
// また、Claude/Cline の FileName が空の場合は既定値を補完します。
func LoadSettings(settingsPath string) (*Settings, error) {
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return DefaultSettings(), nil
	}

	var settings Settings
	if _, err := toml.DecodeFile(settingsPath, &settings); err != nil {
		return nil, err
	}

	// デフォルト値を設定
	if settings.Claude.FileName == "" {
		settings.Claude.FileName = "CLAUDE.md"
	}
	if settings.Cline.FileName == "" {
		settings.Cline.FileName = ".clinerules"
	}

	return &settings, nil
}

// LoadConfig は指定された JSON ファイル (configPath) から設定を読み込みます。
// ファイルが存在しない場合はデフォルトの Config を返します。
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// LoadConfigWithSettings は Config と Settings をそれぞれ読み込み、
// 読み込んだ Settings を Config に紐付けて返します。
func LoadConfigWithSettings(configPath, settingsPath string) (*Config, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	settings, err := LoadSettings(settingsPath)
	if err != nil {
		return nil, err
	}

	config.Settings = settings
	return config, nil
}

// Save は Config を JSON 形式で configPath に保存します。
// 保存先ディレクトリが存在しない場合は作成します。
func (c *Config) Save(configPath string) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
