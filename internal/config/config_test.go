package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultSettings(t *testing.T) {
	tmpDir := t.TempDir()
	settings, err := DefaultSettings(tmpDir)
	require.NoError(t, err)

	// Language field removed from AppSettings
	assert.Equal(t, filepath.Join(tmpDir, ".system_prompt"), settings.App.InputDir)
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
header = "header"
footer = "footer"

[tools.claude]
generate = false
dir_name = "custom/path"
file_name = "custom.md"

[tools.cline]
generate = true

[tools.mytool]
generate = true
dir_name = "tools"
file_name = "mytool.md"`,
			expectedSettings: &Settings{
				App: AppSettings{
					Header:   "header",
					Footer:   "footer",
				},
				Tools: map[string]AIToolSettings{
					"claude": {
						Generate: false,
					},
					"cline": {
						Generate: true,
						AIToolPaths: AIToolPaths{
							DirName:  "",
							FileName: ".clinerules",
						},
					},
					"mytool": {
						Generate: true,
						AIToolPaths: AIToolPaths{
							DirName:  "tools",
							FileName: "mytool.md",
						},
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

			fmt.Println(settings)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedSettings != nil {
					assert.Equal(t, tt.expectedSettings.App.Header, settings.App.Header)
					assert.Equal(t, tt.expectedSettings.App.Footer, settings.App.Footer)
					assert.NotEqual(t, "", settings.App.InputDir)
					assert.NotEqual(t, "", settings.App.OutputDir)

					for name, tool := range tt.expectedSettings.Tools {
						if tool.Generate {
							assert.Equal(t, tool.DirName, settings.Tools[name].DirName)
							assert.Equal(t, tool.FileName, settings.Tools[name].FileName)
						} else {
							_, ok := settings.Tools[name]
							assert.False(t, ok)
						}
					}
				}
			}
		})
	}
}

func TestSettingsDefaultValues(t *testing.T) {
	tempDir := t.TempDir()

	// Create minimal settings file without file_name fields
	settingsContent := `[tools.claude]
generate = true

[tools.cline]
generate = false`

	settingsPath := filepath.Join(tempDir, "settings.toml")
	err := os.WriteFile(settingsPath, []byte(settingsContent), 0644)
	require.NoError(t, err)

	settings, err := LoadSettings(settingsPath)
	require.NoError(t, err)

	assert.Equal(t, FileName("CLAUDE.md"), settings.Tools["claude"].FileName)
	assert.True(t, settings.Tools["claude"].Generate)

	_, ok := settings.Tools["cline"]
	assert.False(t, ok)
}
