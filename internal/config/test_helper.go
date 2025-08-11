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

	settings.Claude.Path = outputTmpDir
	settings.Cline.Path = outputTmpDir

	settings.Custom = map[string]AIToolSettings{
		"test": {
			Generate: true,
			Path:     outputTmpDir,
			FileName: "test.md",
		},
	}

	return settings
}
