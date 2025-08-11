package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, ".system_prompt", config.InputDir)
	assert.Equal(t, []string{"CLAUDE.md", ".clinerules"}, config.OutputFiles)
	assert.Equal(t, []string{}, config.ExcludeFiles)
	assert.Equal(t, "# System Prompt\n\n", config.Header)
	assert.Equal(t, "", config.Footer)
	assert.NotNil(t, config.Settings)
}

func TestDefaultSettings(t *testing.T) {
	settings := DefaultSettings()

	assert.Equal(t, "", settings.App.Language)
	assert.True(t, settings.Claude.Generate)
	assert.Equal(t, "", settings.Claude.Path)
	assert.Equal(t, "CLAUDE.md", settings.Claude.FileName)
	assert.True(t, settings.Cline.Generate)
	assert.Equal(t, "", settings.Cline.Path)
	assert.Equal(t, ".clinerules", settings.Cline.FileName)
	assert.NotNil(t, settings.Custom)
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		expectedConfig *Config
		expectError    bool
	}{
		{
			name: "valid config",
			configContent: `{
				"inputDir": "custom_input",
				"outputFiles": ["custom.md"],
				"excludeFiles": ["*.tmp"],
				"header": "Custom Header\n",
				"footer": "Custom Footer\n"
			}`,
			expectedConfig: &Config{
				InputDir:     "custom_input",
				OutputFiles:  []string{"custom.md"},
				ExcludeFiles: []string{"*.tmp"},
				Header:       "Custom Header\n",
				Footer:       "Custom Footer\n",
			},
			expectError: false,
		},
		{
			name:          "invalid json",
			configContent: `{invalid json}`,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.json")

			if tt.configContent != "" {
				err := os.WriteFile(configPath, []byte(tt.configContent), 0644)
				require.NoError(t, err)
			}

			config, err := LoadConfig(configPath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedConfig != nil {
					assert.Equal(t, tt.expectedConfig.InputDir, config.InputDir)
					assert.Equal(t, tt.expectedConfig.OutputFiles, config.OutputFiles)
					assert.Equal(t, tt.expectedConfig.ExcludeFiles, config.ExcludeFiles)
					assert.Equal(t, tt.expectedConfig.Header, config.Header)
					assert.Equal(t, tt.expectedConfig.Footer, config.Footer)
				}
			}
		})
	}
}

func TestLoadConfigNonExistentFile(t *testing.T) {
	config, err := LoadConfig("non_existent_file.json")

	require.NoError(t, err)
	// Should return default config when file doesn't exist
	defaultConfig := DefaultConfig()
	assert.Equal(t, defaultConfig.InputDir, config.InputDir)
	assert.Equal(t, defaultConfig.OutputFiles, config.OutputFiles)
}

func TestLoadSettings(t *testing.T) {
	tests := []struct {
		name             string
		settingsContent  string
		expectedSettings *Settings
		expectError      bool
	}{
		{
			name: "valid settings",
			settingsContent: `[app]
language = "ja"

[claude]
generate = false
path = "custom/path"
file_name = "custom.md"

[cline]
generate = true
path = ""
file_name = ""

[custom.mytool]
generate = true
path = "tools"
file_name = "mytool.md"`,
			expectedSettings: &Settings{
				App: AppSettings{
					Language: "ja",
				},
				Claude: AIToolSettings{
					Generate: false,
					Path:     "custom/path",
					FileName: "custom.md",
				},
				Cline: AIToolSettings{
					Generate: true,
					Path:     "",
					FileName: ".clinerules", // Should be set to default
				},
				Custom: map[string]AIToolSettings{
					"mytool": {
						Generate: true,
						Path:     "tools",
						FileName: "mytool.md",
					},
				},
			},
			expectError: false,
		},
		{
			name:            "invalid toml",
			settingsContent: `[invalid toml`,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary settings file
			tempDir := t.TempDir()
			settingsPath := filepath.Join(tempDir, "settings.toml")

			if tt.settingsContent != "" {
				err := os.WriteFile(settingsPath, []byte(tt.settingsContent), 0644)
				require.NoError(t, err)
			}

			settings, err := LoadSettings(settingsPath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedSettings != nil {
					assert.Equal(t, tt.expectedSettings.App.Language, settings.App.Language)
					assert.Equal(t, tt.expectedSettings.Claude.Generate, settings.Claude.Generate)
					assert.Equal(t, tt.expectedSettings.Claude.Path, settings.Claude.Path)
					assert.Equal(t, tt.expectedSettings.Claude.FileName, settings.Claude.FileName)
					assert.Equal(t, tt.expectedSettings.Cline.Generate, settings.Cline.Generate)
					assert.Equal(t, tt.expectedSettings.Cline.FileName, settings.Cline.FileName)

					for toolName, expectedTool := range tt.expectedSettings.Custom {
						actualTool, exists := settings.Custom[toolName]
						assert.True(t, exists, "custom tool %s should exist", toolName)
						assert.Equal(t, expectedTool.Generate, actualTool.Generate)
						assert.Equal(t, expectedTool.Path, actualTool.Path)
						assert.Equal(t, expectedTool.FileName, actualTool.FileName)
					}
				}
			}
		})
	}
}

func TestLoadSettingsNonExistentFile(t *testing.T) {
	settings, err := LoadSettings("non_existent_settings.toml")

	require.NoError(t, err)
	// Should return default settings when file doesn't exist
	defaultSettings := DefaultSettings()
	assert.Equal(t, defaultSettings.App.Language, settings.App.Language)
	assert.Equal(t, defaultSettings.Claude.Generate, settings.Claude.Generate)
}

func TestLoadConfigWithSettings(t *testing.T) {
	tempDir := t.TempDir()

	// Create config file
	configContent := `{
		"inputDir": "test_input",
		"outputFiles": ["test.md"]
	}`
	configPath := filepath.Join(tempDir, "config.json")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Create settings file
	settingsContent := `[app]
language = "en"

[claude]
generate = true
path = "output"`
	settingsPath := filepath.Join(tempDir, "settings.toml")
	err = os.WriteFile(settingsPath, []byte(settingsContent), 0644)
	require.NoError(t, err)

	config, err := LoadConfigWithSettings(configPath, settingsPath)
	require.NoError(t, err)

	// Check config values
	assert.Equal(t, "test_input", config.InputDir)
	assert.Equal(t, []string{"test.md"}, config.OutputFiles)

	// Check settings values
	assert.NotNil(t, config.Settings)
	assert.Equal(t, "en", config.Settings.App.Language)
	assert.True(t, config.Settings.Claude.Generate)
	assert.Equal(t, "output", config.Settings.Claude.Path)
}

func TestConfigSave(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "subdir", "config.json")

	config := &Config{
		InputDir:     "custom_input",
		OutputFiles:  []string{"custom.md"},
		ExcludeFiles: []string{"*.tmp"},
		Header:       "Test Header",
		Footer:       "Test Footer",
	}

	err := config.Save(configPath)
	require.NoError(t, err)

	// Check if file exists and directory was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Load and verify content
	loadedConfig, err := LoadConfig(configPath)
	require.NoError(t, err)

	assert.Equal(t, config.InputDir, loadedConfig.InputDir)
	assert.Equal(t, config.OutputFiles, loadedConfig.OutputFiles)
	assert.Equal(t, config.ExcludeFiles, loadedConfig.ExcludeFiles)
	assert.Equal(t, config.Header, loadedConfig.Header)
	assert.Equal(t, config.Footer, loadedConfig.Footer)
}

func TestSettingsDefaultValues(t *testing.T) {
	tempDir := t.TempDir()

	// Create minimal settings file without file_name fields
	settingsContent := `[claude]
generate = true

[cline]
generate = false`

	settingsPath := filepath.Join(tempDir, "settings.toml")
	err := os.WriteFile(settingsPath, []byte(settingsContent), 0644)
	require.NoError(t, err)

	settings, err := LoadSettings(settingsPath)
	require.NoError(t, err)

	// Check that default values are applied
	assert.Equal(t, "CLAUDE.md", settings.Claude.FileName)
	assert.Equal(t, ".clinerules", settings.Cline.FileName)
}
