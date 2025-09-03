package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type FileName string
type DirName string

type AIToolSettings struct {
	Generate bool     `toml:"generate"`
	Include  []string `toml:"include"`
	Exclude  []string `toml:"exclude"`
	AIToolPaths
}

type AIToolPaths struct {
	DirName  DirName  `toml:"dir_name"`
	FileName FileName `toml:"file_name"`
}

type AppSettings struct {
	Header    string `toml:"header"`
	Footer    string `toml:"footer"`
	InputDir  string `toml:"input_dir"`
	OutputDir string `toml:"output_dir"`
}

type Settings struct {
	App   AppSettings               `toml:"app"`
	Tools map[string]AIToolSettings `toml:"tools"`
}

var DefaultKnownToolFileNames = map[string]AIToolPaths{
    "claude": {
        DirName:  "",
        FileName: "CLAUDE.md",
    },
    "cline": {
        DirName:  "",
        FileName: ".clinerules",
    },
    "github_copilot": {
        DirName:  ".github",
        FileName: "copilot-instructions.md",
    },
    "agents": {
        DirName:  "",
        FileName: "AGENTS.md",
    },
}

// DefaultSettings はアプリケーションの設定 (Settings) のデフォルト値を返します。
// これには App/Claude/Cline/Custom の初期値が含まれます。
func DefaultSettings(currentDir string) (*Settings, error) {
	inputDir := filepath.Join(currentDir, ".system_prompt")

	tools := make(map[string]AIToolSettings)

	for name, paths := range DefaultKnownToolFileNames {
		tools[name] = AIToolSettings{
			Generate: true,
			AIToolPaths: AIToolPaths{
				DirName:  paths.DirName,
				FileName: paths.FileName,
			},
		}
	}

	return &Settings{
		App: AppSettings{
			InputDir: inputDir,
		},
		Tools: tools,
	}, nil
}

// LoadSettings は指定された TOML ファイル (settingsPath) から設定を読み込みます。
// ファイルが存在しない場合はデフォルト設定を返します。
// また、Claude/Cline の FileName が空の場合は既定値を補完します。
func LoadSettings(settingsPath string) (*Settings, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return DefaultSettings(currentDir)
	}

	var settings Settings
	if _, err := toml.DecodeFile(settingsPath, &settings); err != nil {
		return nil, err
	}

	if settings.App.InputDir == "" {
		settings.App.InputDir = filepath.Join(currentDir, ".system_prompt")
	}
	if settings.App.OutputDir == "" {
		settings.App.OutputDir = currentDir
	}

	var newTools = make(map[string]AIToolSettings)

	for name, tool := range settings.Tools {
		if !tool.Generate {
			continue
		}

		knownTool, ok := DefaultKnownToolFileNames[name]
		if ok {
			dirName := tool.DirName
			if dirName == "" {
				dirName = knownTool.DirName
			}
			fileName := tool.FileName
			if fileName == "" {
				fileName = knownTool.FileName
			}

			newTools[name] = AIToolSettings{
				Generate: true,
				Include:  tool.Include,
				Exclude:  tool.Exclude,
				AIToolPaths: AIToolPaths{
					DirName:  dirName,
					FileName: fileName,
				},
			}
		} else {
			if tool.FileName == "" {
				return nil, fmt.Errorf("tool %q is missing file_name", name)
			}

			newTools[name] = tool
		}
	}

	settings.Tools = newTools

	return &settings, nil
}
