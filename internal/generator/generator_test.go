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

func TestNew(t *testing.T) {
	settings := config.TestSettings(t)
	gen := New(settings)

	assert.NotNil(t, gen)
	assert.Equal(t, settings, gen.settings)
}

func TestCollectPromptFiles(t *testing.T) {
	i18n.TestSetupI18n(t)

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
	i18n.TestSetupI18n(t)

	settings := config.TestSettings(t)

	gen := New(settings)
	files, err := gen.CollectPromptFiles()

	require.NoError(t, err)
	assert.Len(t, files, 0)
}

func TestCollectPromptFilesNonExistentDirectory(t *testing.T) {
	i18n.TestSetupI18n(t)

	appSettings := config.AppSettings{
		InputDir: "/non/existent/directory",
		Header:   "",
		Footer:   "",
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
	i18n.TestSetupI18n(t)

	tempDir := t.TempDir()

	settings := &config.Settings{
		App: config.AppSettings{
			InputDir:  filepath.Join(tempDir, "input"),
			OutputDir: tempDir,
		},
		Tools: map[string]config.AIToolSettings{
			"claude": config.AIToolSettings{
				Generate: true,
				AIToolPaths: config.AIToolPaths{
					FileName: "CLAUDE.md",
				},
			},
			"mytool": config.AIToolSettings{
				Generate: true,
				AIToolPaths: config.AIToolPaths{
					FileName: "mytool.md",
				},
			},
		},
	}

	gen := New(settings)

	content := "Test content for TOML mode"
	err := gen.WriteOutputFiles(content)

	require.NoError(t, err)

	// Check Claude file was created
	claudeFile := filepath.Join(tempDir, "CLAUDE.md")
	testutil.AssertFileExists(t, claudeFile)
	claudeContent := testutil.ReadTestFile(t, claudeFile)
	assert.Equal(t, content, claudeContent)

	// Check custom tool file was created
	customFile := filepath.Join(tempDir, "mytool.md")
	testutil.AssertFileExists(t, customFile)
	customContent := testutil.ReadTestFile(t, customFile)
	assert.Equal(t, content, customContent)
}

func TestWriteOutputFiles_TOMLModeWithEmptyPath(t *testing.T) {
	i18n.TestSetupI18n(t)

	tempDir := t.TempDir()

	// Change working directory to temp dir for this test
	originalWd, wd_err := os.Getwd()
	require.NoError(t, wd_err)

	chdir_err := os.Chdir(tempDir)
	require.NoError(t, chdir_err)

	t.Cleanup(func() {
		err := os.Chdir(originalWd)
		require.NoError(t, err)
	})

	settings := &config.Settings{
		App: config.AppSettings{
			OutputDir: tempDir,
		},
		Tools: map[string]config.AIToolSettings{
			"claude": {
				Generate: true,
				AIToolPaths: config.AIToolPaths{
					FileName: "CLAUDE.md",
				},
			},
			"cline": {
				Generate: true,
				AIToolPaths: config.AIToolPaths{
					FileName: ".clinerules",
				},
			},
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
	tmpDir := t.TempDir()
	settings := &config.Settings{
		App: config.AppSettings{
			OutputDir: tmpDir,
		},
		Tools: map[string]config.AIToolSettings{
			"claude": {
				Generate: true,
				AIToolPaths: config.AIToolPaths{
					FileName: "CLAUDE.md",
				},
			},
			"tool1": {
				Generate: true,
				AIToolPaths: config.AIToolPaths{
					DirName:  "custom",
					FileName: "tool1.md",
				},
			},
		},
	}
	gen := New(settings)

	targets := gen.GetGeneratedTargets()

	expected := []string{
		filepath.Join(tmpDir, "CLAUDE.md"),
		filepath.Join(tmpDir, "custom", "tool1.md"),
	}

	assert.ElementsMatch(t, expected, targets)
}

func TestRun_Success(t *testing.T) {
	i18n.TestSetupI18n(t)

	settings := config.TestSettings(t)

	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "test.md"), "# Test\nTest content\n")

	gen := New(settings)
	err := gen.Run()

	require.NoError(t, err)

	outputFile := path.Join(settings.App.OutputDir, "CLAUDE.md")

	// Check output file was created
	testutil.AssertFileExists(t, outputFile)
	content := testutil.ReadTestFile(t, outputFile)

	assert.Contains(t, content, "Test Header")
	assert.Contains(t, content, "Test content")
	assert.Contains(t, content, "Test Footer")
}

func TestRun_NoFiles(t *testing.T) {
	i18n.TestSetupI18n(t)

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
	i18n.TestSetupI18n(t)

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
	i18n.TestSetupI18n(t)

	tempDir := t.TempDir()

	settings := &config.Settings{
		App: config.AppSettings{
			OutputDir: tempDir,
		},
		Tools: map[string]config.AIToolSettings{
			"claude": {
				Generate: true,
				AIToolPaths: config.AIToolPaths{
					DirName:  "deep/nested/directory",
					FileName: "CLAUDE.md",
				},
			},
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

func TestCollectPromptFilesForTool(t *testing.T) {
	i18n.TestSetupI18n(t)

	settings := config.TestSettings(t)

	// Create test files
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "01_first.md"), "# First\nContent of first file\n")
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "02_second.md"), "# Second\nContent of second file\n")
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "003_excluded.md"), "# Excluded\nThis should be excluded\n")
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "temp_file.md"), "# Temp\nTemporary file\n")

	gen := New(settings)

	// Test tool with exclude patterns
	toolSettings := config.AIToolSettings{
		Generate: true,
		Exclude:  []string{"003_*.md", "temp*.md"},
		AIToolPaths: config.AIToolPaths{
			FileName: "tool.md",
		},
	}

	files, err := gen.CollectPromptFilesForTool("test_tool", toolSettings)

	require.NoError(t, err)
	assert.Len(t, files, 2) // Should find 2 .md files (excluding the patterns)

	// Check files are sorted by filename and excluded files are not present
	fileNames := []string{"01_first.md", "02_second.md"}
	for i, file := range files {
		assert.Equal(t, fileNames[i], file.Filename)
		assert.NotEmpty(t, file.Content)
		assert.NotEmpty(t, file.Path)
	}

	// Check specific content
	assert.Contains(t, files[0].Content, "Content of first file")
	assert.Contains(t, files[1].Content, "Content of second file")

	// Verify excluded files are not present
	for _, file := range files {
		assert.NotEqual(t, "003_excluded.md", file.Filename)
		assert.NotEqual(t, "temp_file.md", file.Filename)
	}
}

func TestCollectPromptFilesForToolWithInclude(t *testing.T) {
	i18n.TestSetupI18n(t)

	settings := config.TestSettings(t)

	// Create test files
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "001_first.md"), "# First\nContent of first file\n")
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "002_second.md"), "# Second\nContent of second file\n")
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "003_third.md"), "# Third\nContent of third file\n")
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "public_doc.md"), "# Public\nPublic documentation\n")

	gen := New(settings)

	// Test tool with include patterns
	toolSettings := config.AIToolSettings{
		Generate: true,
		Include:  []string{"001_*.md", "002_*.md"},
		AIToolPaths: config.AIToolPaths{
			FileName: "tool.md",
		},
	}

	files, err := gen.CollectPromptFilesForTool("test_tool", toolSettings)

	require.NoError(t, err)
	assert.Len(t, files, 2) // Should find 2 .md files matching include patterns

	// Check files are sorted by filename and only included files are present
	fileNames := []string{"001_first.md", "002_second.md"}
	for i, file := range files {
		assert.Equal(t, fileNames[i], file.Filename)
		assert.NotEmpty(t, file.Content)
		assert.NotEmpty(t, file.Path)
	}

	// Verify non-included files are not present
	for _, file := range files {
		assert.NotEqual(t, "003_third.md", file.Filename)
		assert.NotEqual(t, "public_doc.md", file.Filename)
	}
}

func TestCollectPromptFilesForToolWithIncludeAndExclude(t *testing.T) {
	i18n.TestSetupI18n(t)

	settings := config.TestSettings(t)

	// Create test files
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "public_doc.md"), "# Public\nPublic documentation\n")
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "public_secret.md"), "# Secret\nSecret information\n")
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "private_doc.md"), "# Private\nPrivate documentation\n")

	gen := New(settings)

	// Test tool with both include and exclude patterns
	// Include all public_*.md files but exclude public_secret.md
	toolSettings := config.AIToolSettings{
		Generate: true,
		Include:  []string{"public_*.md"},
		Exclude:  []string{"public_secret.md"},
		AIToolPaths: config.AIToolPaths{
			FileName: "tool.md",
		},
	}

	files, err := gen.CollectPromptFilesForTool("test_tool", toolSettings)

	require.NoError(t, err)
	assert.Len(t, files, 1) // Should find only public_doc.md

	// Check correct file is included
	assert.Equal(t, "public_doc.md", files[0].Filename)
	assert.Contains(t, files[0].Content, "Public documentation")

	// Verify excluded and non-included files are not present
	for _, file := range files {
		assert.NotEqual(t, "public_secret.md", file.Filename)
		assert.NotEqual(t, "private_doc.md", file.Filename)
	}
}

func TestCollectPromptFilesForToolIncludeUndefined(t *testing.T) {
	i18n.TestSetupI18n(t)

	settings := config.TestSettings(t)

	// Create test files
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "001_first.md"), "# First\nContent of first file\n")
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "002_second.md"), "# Second\nContent of second file\n")
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "003_third.md"), "# Third\nContent of third file\n")

	gen := New(settings)

	// Test tool without include patterns (should include all files)
	toolSettings := config.AIToolSettings{
		Generate: true,
		Include:  []string{}, // Empty include patterns
		AIToolPaths: config.AIToolPaths{
			FileName: "tool.md",
		},
	}

	files, err := gen.CollectPromptFilesForTool("test_tool", toolSettings)

	require.NoError(t, err)
	assert.Len(t, files, 3) // Should find all .md files

	// Check files are sorted by filename
	fileNames := []string{"001_first.md", "002_second.md", "003_third.md"}
	for i, file := range files {
		assert.Equal(t, fileNames[i], file.Filename)
		assert.NotEmpty(t, file.Content)
		assert.NotEmpty(t, file.Path)
	}
}

func TestCollectPromptFilesForToolNoExclude(t *testing.T) {
	i18n.TestSetupI18n(t)

	settings := config.TestSettings(t)

	// Create test files
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "01_first.md"), "# First\nContent of first file\n")
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "02_second.md"), "# Second\nContent of second file\n")

	gen := New(settings)

	// Test tool without exclude patterns
	toolSettings := config.AIToolSettings{
		Generate: true,
		Exclude:  []string{}, // No exclude patterns
		AIToolPaths: config.AIToolPaths{
			FileName: "tool.md",
		},
	}

	files, err := gen.CollectPromptFilesForTool("test_tool", toolSettings)

	require.NoError(t, err)
	assert.Len(t, files, 2) // Should find all .md files

	// Check files are sorted by filename
	fileNames := []string{"01_first.md", "02_second.md"}
	for i, file := range files {
		assert.Equal(t, fileNames[i], file.Filename)
		assert.NotEmpty(t, file.Content)
		assert.NotEmpty(t, file.Path)
	}
}

func TestWriteOutputFilesWithExcludes(t *testing.T) {
	i18n.TestSetupI18n(t)

	tempDir := t.TempDir()

	settings := &config.Settings{
		App: config.AppSettings{
			InputDir:  filepath.Join(tempDir, "input"),
			OutputDir: tempDir,
		},
		Tools: map[string]config.AIToolSettings{
			"claude": config.AIToolSettings{
				Generate: true,
				Exclude:  []string{"003_*.md"},
				AIToolPaths: config.AIToolPaths{
					FileName: "CLAUDE.md",
				},
			},
			"cline": config.AIToolSettings{
				Generate: true,
				Exclude:  []string{"001_*.md"},
				AIToolPaths: config.AIToolPaths{
					FileName: ".clinerules",
				},
			},
		},
	}

	// Create input directory and files
	err := os.MkdirAll(settings.App.InputDir, 0755)
	require.NoError(t, err)

	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "001_first.md"), "# First\nContent of first file\n")
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "002_second.md"), "# Second\nContent of second file\n")
	testutil.CreateTestFile(t, filepath.Join(settings.App.InputDir, "003_third.md"), "# Third\nContent of third file\n")

	gen := New(settings)

	err = gen.WriteOutputFilesWithExcludes()
	require.NoError(t, err)

	// Check Claude file (should exclude 003_*.md)
	claudeFile := filepath.Join(tempDir, "CLAUDE.md")
	testutil.AssertFileExists(t, claudeFile)
	claudeContent := testutil.ReadTestFile(t, claudeFile)
	assert.Contains(t, claudeContent, "Content of first file")
	assert.Contains(t, claudeContent, "Content of second file")
	assert.NotContains(t, claudeContent, "Content of third file")

	// Check Cline file (should exclude 001_*.md)
	clineFile := filepath.Join(tempDir, ".clinerules")
	testutil.AssertFileExists(t, clineFile)
	clineContent := testutil.ReadTestFile(t, clineFile)
	assert.NotContains(t, clineContent, "Content of first file")
	assert.Contains(t, clineContent, "Content of second file")
	assert.Contains(t, clineContent, "Content of third file")
}
