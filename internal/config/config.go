package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type AIToolSettings struct {
	Generate bool   `toml:"generate"`
	Path     string `toml:"path"`
	FileName string `toml:"file_name"`
}

type AppSettings struct {
	Language string `toml:"language"`
	Header   string `toml:"header"`
	Footer   string `toml:"footer"`
}

type Settings struct {
	App    AppSettings               `toml:"app"`
	Claude AIToolSettings            `toml:"claude"`
	Cline  AIToolSettings            `toml:"cline"`
	Custom map[string]AIToolSettings `toml:"custom"`
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
	// FIXME: この定義好きじゃない。メソッド分けたい
	if settings.Claude.FileName == "" {
		settings.Claude.FileName = "CLAUDE.md"
	}
	if settings.Cline.FileName == "" {
		settings.Cline.FileName = ".clinerules"
	}

	return &settings, nil
}
