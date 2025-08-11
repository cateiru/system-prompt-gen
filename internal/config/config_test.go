package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultSettings(t *testing.T) {
	settings, err := DefaultSettings()
	require.NoError(t, err)

	assert.Equal(t, "", settings.App.Language)
	assert.True(t, settings.Claude.Generate)
	assert.Equal(t, "", settings.Claude.Path)
	assert.Equal(t, "CLAUDE.md", settings.Claude.FileName)
	assert.True(t, settings.Cline.Generate)
	assert.Equal(t, "", settings.Cline.Path)
	assert.Equal(t, ".clinerules", settings.Cline.FileName)
	assert.NotNil(t, settings.Custom)
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
	defaultSettings, err := DefaultSettings()
	require.NoError(t, err)
	assert.Equal(t, defaultSettings.App.Language, settings.App.Language)
	assert.Equal(t, defaultSettings.Claude.Generate, settings.Claude.Generate)
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
