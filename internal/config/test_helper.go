package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSettings(t *testing.T, appSettings ...AppSettings) *Settings {
	inputTmpDir := t.TempDir()

	settings, err := DefaultSettings(inputTmpDir)
	require.NoError(t, err)

	outputTmpDir := t.TempDir()

	if len(appSettings) > 0 {
		settings.App = appSettings[0]
	} else {

		settings.App = AppSettings{
			Language: "en",
			Header:   "Test Header\n",
			Footer:   "Test Footer\n",
			InputDir: inputTmpDir,
		}
	}

	settings.App.OutputDir = outputTmpDir
	settings.Tools["test"] = AIToolSettings{
		Generate: true,
		AIToolPaths: AIToolPaths{
			FileName: "test.md",
		},
	}

	return settings
}
