package generator

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cateiru/system-prompt-gen/internal/config"
	"github.com/cateiru/system-prompt-gen/internal/i18n"
	"github.com/cateiru/system-prompt-gen/internal/testutil"
)

func setupI18n() {
	// Initialize i18n for testing - ignore errors
	i18n.Initialize("en")
}

func TestNew(t *testing.T) {
	settings := config.TestSettings(t)
	gen := New(settings)

	assert.NotNil(t, gen)
	assert.Equal(t, settings, gen.settings)
}

func TestCollectPromptFiles(t *testing.T) {
	setupI18n()

	settings := config.TestSettings(t)

	// Create test files
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "01_first.md"), "# First\nContent of first file\n")
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "02_second.md"), "# Second\nContent of second file\n")
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "not_markdown.txt"), "This is not markdown")

	// Create subdirectory with markdown file
	subDir := filepath.Join(settings.App.InputDir, "subdir")
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err)
	testutil.CreateTestFile(t, filepath.Join(subDir, "04_subdir.md"), "# Subdir\nContent from subdirectory\n")

	gen := New(settings)
	files, err := gen.CollectPromptFiles()

	require.NoError(t, err)
	assert.Len(t, files, 3) // Should find 3 .md files (excluding the excluded one and non-markdown file)

	// Check files are sorted by filename
	fileNames := []string{"01_first.md", "02_second.md", "04_subdir.md"}
	for i, file := range files {
		assert.Equal(t, fileNames[i], file.Filename)
		assert.NotEmpty(t, file.Content)
		assert.NotEmpty(t, file.Path)
	}

	// Check specific content
	assert.Contains(t, files[0].Content, "Content of first file")
	assert.Contains(t, files[1].Content, "Content of second file")
	assert.Contains(t, files[2].Content, "Content from subdirectory")
}

func TestCollectPromptFilesEmptyDirectory(t *testing.T) {
	setupI18n()
	settings := config.TestSettings(t)

	gen := New(settings)
	files, err := gen.CollectPromptFiles()

	require.NoError(t, err)
	assert.Len(t, files, 0)
}

func TestCollectPromptFilesNonExistentDirectory(t *testing.T) {
	setupI18n()

	appSettings := config.AppSettings{
		InputDir: "/non/existent/directory",
		Header:   "",
		Footer:   "",
		Language: "",
	}

	settings := config.TestSettings(t, appSettings)

	gen := New(settings)
	files, err := gen.CollectPromptFiles()

	assert.Error(t, err)
	assert.Nil(t, files)
}

func TestGeneratePrompt(t *testing.T) {
	appSettings := config.AppSettings{
		Header: "# System Prompt Header\n\n",
		Footer: "\n---\nFooter content\n",
	}
	settings := config.TestSettings(t, appSettings)
	gen := New(settings)

	files := []PromptFile{
		{
			Path:     "/path/to/01_first.md",
			Filename: "01_first.md",
			Content:  "Content of first file",
		},
		{
			Path:     "/path/to/02_second.md",
			Filename: "02_second.md",
			Content:  "Content of second file\n",
		},
		{
			Path:     "/path/to/03_third.md",
			Filename: "03_third.md",
			Content:  "Content without trailing newline",
		},
	}

	result := gen.GeneratePrompt(files)

	expected := "# System Prompt Header\n\n" +
		"# 01_first\n\n" +
		"Content of first file\n\n" +
		"# 02_second\n\n" +
		"Content of second file\n\n" +
		"# 03_third\n\n" +
		"Content without trailing newline\n\n" +
		"\n---\nFooter content\n"

	assert.Equal(t, expected, result)
}

func TestGeneratePromptWithEmptyFiles(t *testing.T) {
	appSettings := config.AppSettings{
		Header: "Header\n",
		Footer: "Footer\n",
	}
	settings := config.TestSettings(t, appSettings)
	gen := New(settings)

	files := []PromptFile{}
	result := gen.GeneratePrompt(files)

	expected := "Header\nFooter\n"
	assert.Equal(t, expected, result)
}

func TestWriteOutputFiles_TOMLMode(t *testing.T) {
	setupI18n()
	tempDir := t.TempDir()

	settings := &config.Settings{
		Claude: config.AIToolSettings{
			Generate: true,
			Path:     filepath.Join(tempDir, "claude"),
			FileName: "CLAUDE.md",
		},
		Cline: config.AIToolSettings{
			Generate: false, // Should not generate
			Path:     tempDir,
			FileName: ".clinerules",
		},
		Custom: map[string]config.AIToolSettings{
			"mytool": {
				Generate: true,
				Path:     filepath.Join(tempDir, "custom"),
				FileName: "mytool.md",
			},
		},
	}
	gen := New(settings)

	content := "Test content for TOML mode"
	err := gen.WriteOutputFiles(content)

	require.NoError(t, err)

	// Check Claude file was created
	claudeFile := filepath.Join(tempDir, "claude", "CLAUDE.md")
	testutil.AssertFileExists(t, claudeFile)
	claudeContent := testutil.ReadTestFile(t, claudeFile)
	assert.Equal(t, content, claudeContent)

	// Check Cline file was NOT created (generate = false)
	clineFile := filepath.Join(tempDir, ".clinerules")
	testutil.AssertFileNotExists(t, clineFile)

	// Check custom tool file was created
	customFile := filepath.Join(tempDir, "custom", "mytool.md")
	testutil.AssertFileExists(t, customFile)
	customContent := testutil.ReadTestFile(t, customFile)
	assert.Equal(t, content, customContent)
}

func TestWriteOutputFiles_TOMLModeWithEmptyPath(t *testing.T) {
	setupI18n()
	tempDir := t.TempDir()

	// Change working directory to temp dir for this test
	originalWd, wd_err := os.Getwd()
	require.NoError(t, wd_err)

	chdir_err := os.Chdir(tempDir)
	require.NoError(t, chdir_err)

	t.Cleanup(func() {
		os.Chdir(originalWd)
	})

	settings := &config.Settings{
		Claude: config.AIToolSettings{
			Generate: true,
			Path:     "", // Empty path should default to current directory
			FileName: "CLAUDE.md",
		},
		Cline: config.AIToolSettings{
			Generate: true,
			Path:     "", // Empty path should default to current directory
			FileName: ".clinerules",
		},
	}

	gen := New(settings)

	content := "Test content with empty path"
	err := gen.WriteOutputFiles(content)

	require.NoError(t, err)

	// Files should be created in current directory (tempDir)
	claudeFile := filepath.Join(tempDir, "CLAUDE.md")
	clineFile := filepath.Join(tempDir, ".clinerules")

	testutil.AssertFileExists(t, claudeFile)
	testutil.AssertFileExists(t, clineFile)

	claudeContent := testutil.ReadTestFile(t, claudeFile)
	clineContent := testutil.ReadTestFile(t, clineFile)

	assert.Equal(t, content, claudeContent)
	assert.Equal(t, content, clineContent)
}

func TestGetGeneratedTargets_TOMLMode(t *testing.T) {
	settings := &config.Settings{
		Claude: config.AIToolSettings{
			Generate: true,
			Path:     "claude",
			FileName: "CLAUDE.md",
		},
		Cline: config.AIToolSettings{
			Generate: false, // Should not appear in targets
			Path:     "",
			FileName: ".clinerules",
		},
		Custom: map[string]config.AIToolSettings{
			"tool1": {
				Generate: true,
				Path:     "custom/tool1",
				FileName: "tool1.md",
			},
			"tool2": {
				Generate: false, // Should not appear in targets
				Path:     "custom/tool2",
				FileName: "tool2.md",
			},
		},
	}
	gen := New(settings)

	targets := gen.GetGeneratedTargets()

	expected := []string{
		filepath.Join("claude", "CLAUDE.md"),
		filepath.Join("custom/tool1", "tool1.md"),
	}

	assert.ElementsMatch(t, expected, targets)
}

func TestRun_Success(t *testing.T) {
	setupI18n()

	settings := config.TestSettings(t)

	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "test.md"), "# Test\nTest content\n")

	gen := New(settings)
	err := gen.Run()

	require.NoError(t, err)

	outputFile := path.Join(settings.Custom["test"].Path, settings.Custom["test"].FileName)

	// Check output file was created
	testutil.AssertFileExists(t, outputFile)
	content := testutil.ReadTestFile(t, outputFile)

	assert.Contains(t, content, "Test Header")
	assert.Contains(t, content, "Test content")
	assert.Contains(t, content, "Test Footer")
}

func TestRun_NoFiles(t *testing.T) {
	setupI18n()

	tmpDir := t.TempDir()

	appSettings := config.AppSettings{
		Header:   "",
		Footer:   "",
		InputDir: tmpDir,
	}

	settings := config.TestSettings(t, appSettings)

	gen := New(settings)
	run_err := gen.Run()

	assert.Error(t, run_err)
	assert.Contains(t, strings.ToLower(run_err.Error()), "no")
}

func TestRun_InvalidInputDirectory(t *testing.T) {
	setupI18n()

	appSettings := config.AppSettings{
		Header:   "",
		Footer:   "",
		InputDir: "/non/existent/directory",
	}

	settings := config.TestSettings(t, appSettings)

	gen := New(settings)
	err := gen.Run()

	assert.Error(t, err)
}

func TestWriteOutputFiles_DirectoryCreation(t *testing.T) {
	setupI18n()
	tempDir := t.TempDir()

	settings := &config.Settings{
		Claude: config.AIToolSettings{
			Generate: true,
			Path:     filepath.Join(tempDir, "deep", "nested", "directory"),
			FileName: "CLAUDE.md",
		},
	}

	gen := New(settings)

	content := "Test content for directory creation"
	err := gen.WriteOutputFiles(content)

	require.NoError(t, err)

	// Check that nested directories were created
	outputFile := filepath.Join(tempDir, "deep", "nested", "directory", "CLAUDE.md")
	testutil.AssertFileExists(t, outputFile)

	actualContent := testutil.ReadTestFile(t, outputFile)
	assert.Equal(t, content, actualContent)
}
