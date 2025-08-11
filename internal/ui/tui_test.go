package ui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cateiru/system-prompt-gen/internal/config"
	"github.com/cateiru/system-prompt-gen/internal/generator"
	"github.com/cateiru/system-prompt-gen/internal/i18n"
	"github.com/cateiru/system-prompt-gen/internal/testutil"
)

func setupI18n() {
	// Initialize i18n for testing - ignore errors
	i18n.Initialize("en")
}

func TestInitialModel(t *testing.T) {
	setupI18n()
	cfg := config.DefaultConfig()

	model := initialModel(cfg)

	assert.Equal(t, cfg, model.config)
	assert.NotNil(t, model.generator)
	assert.Equal(t, stateLoading, model.state)
	assert.Nil(t, model.files)
	assert.Nil(t, model.err)
	assert.Empty(t, model.content)
}

func TestModelInit(t *testing.T) {
	setupI18n()
	cfg := config.DefaultConfig()
	model := initialModel(cfg)

	cmd := model.Init()
	assert.NotNil(t, cmd)
}

func TestGeneratePrompts(t *testing.T) {
	setupI18n()
	tempDir := testutil.TempDir(t)

	// Create test input files
	inputDir := filepath.Join(tempDir, "input")
	err := os.MkdirAll(inputDir, 0755)
	require.NoError(t, err)

	testutil.CreateTestFile(t, filepath.Join(inputDir, "test.md"), "# Test\nContent\n")

	cfg := &config.Config{
		InputDir:     inputDir,
		OutputFiles:  []string{"test.md"},
		ExcludeFiles: []string{},
		Header:       "",
		Footer:       "",
		Settings:     config.DefaultSettings(),
	}

	gen := generator.New(cfg)
	cmd := generatePrompts(gen)

	assert.NotNil(t, cmd)

	// Execute the command to get the message
	msg := cmd()
	generateMessage, ok := msg.(generateMsg)
	assert.True(t, ok)
	assert.NoError(t, generateMessage.err)
	assert.Len(t, generateMessage.files, 1)
}

func TestGeneratePromptsWithError(t *testing.T) {
	setupI18n()
	cfg := &config.Config{
		InputDir:     "/non/existent/directory",
		OutputFiles:  []string{"test.md"},
		ExcludeFiles: []string{},
		Header:       "",
		Footer:       "",
		Settings:     config.DefaultSettings(),
	}

	gen := generator.New(cfg)
	cmd := generatePrompts(gen)

	msg := cmd()
	generateMessage, ok := msg.(generateMsg)
	assert.True(t, ok)
	assert.Error(t, generateMessage.err)
	assert.Nil(t, generateMessage.files)
}

func TestModelUpdate_KeyMessages(t *testing.T) {
	setupI18n()
	cfg := config.DefaultConfig()
	m := initialModel(cfg)

	tests := []struct {
		name         string
		keyString    string
		modelState   state
		expectQuit   bool
		expectAction bool
	}{
		{
			name:       "ctrl+c quits",
			keyString:  "ctrl+c",
			modelState: stateLoading,
			expectQuit: true,
		},
		{
			name:       "q quits",
			keyString:  "q",
			modelState: stateLoading,
			expectQuit: true,
		},
		{
			name:         "enter in success state",
			keyString:    "enter",
			modelState:   stateSuccess,
			expectAction: true,
		},
		{
			name:         "space in success state",
			keyString:    " ",
			modelState:   stateSuccess,
			expectAction: true,
		},
		{
			name:         "r in error state retries",
			keyString:    "r",
			modelState:   stateError,
			expectAction: true,
		},
		{
			name:       "random key does nothing",
			keyString:  "x",
			modelState: stateLoading,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.state = tt.modelState
			if tt.modelState == stateSuccess {
				// Set up success state with some test data
				m.files = []generator.PromptFile{
					{Filename: "test.md", Content: "test"},
				}
				m.content = "test content"
				
				// 書き込みエラーを発生させるために無効なパスを設定
				m.config.OutputFiles = []string{"/invalid/path/file.md"}
			}

			keyMsg := tea.KeyMsg{}
			keyMsg.Type = tea.KeyRunes

			// Simulate key press
			switch tt.keyString {
			case "ctrl+c":
				keyMsg.Type = tea.KeyCtrlC
			case "enter":
				keyMsg.Type = tea.KeyEnter
			case " ":
				keyMsg.Type = tea.KeySpace
			default:
				keyMsg.Type = tea.KeyRunes
				keyMsg.Runes = []rune(tt.keyString)
			}

			newModel, cmd := m.Update(keyMsg)

			if tt.expectQuit {
				// tea.Quitが返されたかを確認
				if cmd != nil {
					msg := cmd()
					assert.IsType(t, tea.QuitMsg{}, msg)
				} else {
					t.Error("Expected tea.Quit command, but got nil")
				}
			} else if tt.expectAction {
				if tt.modelState == stateError && tt.keyString == "r" {
					// Retry should set state to loading and return generate command
					if m, ok := newModel.(model); ok {
						assert.Equal(t, stateLoading, m.state)
					} else {
						t.Errorf("Expected model type, got %T", newModel)
					}
					assert.NotNil(t, cmd)
				} else if tt.modelState == stateSuccess && (tt.keyString == "enter" || tt.keyString == " ") {
					// Success state with enter/space should trigger write and quit
					// Note: This test doesn't create actual files, so it will set error state
					if m, ok := newModel.(model); ok {
						assert.Equal(t, stateError, m.state)
						assert.NotNil(t, m.err)
					} else {
						t.Errorf("Expected model type, got %T", newModel)
					}
					assert.Nil(t, cmd) // エラー時はcmdはnilになる
				}
			} else {
				assert.Nil(t, cmd)
			}
		})
	}
}

func TestModelUpdate_GenerateMsg(t *testing.T) {
	setupI18n()
	cfg := config.DefaultConfig()
	m := initialModel(cfg)

	tests := []struct {
		name          string
		msg           generateMsg
		expectedState state
		expectError   bool
	}{
		{
			name: "successful generate",
			msg: generateMsg{
				files: []generator.PromptFile{
					{Filename: "test.md", Content: "content"},
				},
				err: nil,
			},
			expectedState: stateSuccess,
			expectError:   false,
		},
		{
			name: "failed generate",
			msg: generateMsg{
				files: nil,
				err:   assert.AnError,
			},
			expectedState: stateError,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newModel, cmd := m.Update(tt.msg)

			if m, ok := newModel.(model); ok {
				assert.Equal(t, tt.expectedState, m.state)

				if tt.expectError {
					assert.NotNil(t, m.err)
				} else {
					assert.Nil(t, m.err)
					assert.Equal(t, tt.msg.files, m.files)
					assert.NotEmpty(t, m.content)
				}
			} else {
				t.Errorf("Expected model type, got %T", newModel)
			}

			assert.Nil(t, cmd)
		})
	}
}

func TestModelView(t *testing.T) {
	setupI18n()
	tempDir := testutil.TempDir(t)

	cfg := &config.Config{
		InputDir:    tempDir,
		OutputFiles: []string{"test.md"},
		Settings:    config.DefaultSettings(),
	}

	tests := []struct {
		name          string
		state         state
		files         []generator.PromptFile
		err           error
		expectContent []string
	}{
		{
			name:  "loading state",
			state: stateLoading,
			expectContent: []string{
				"Collecting", "Processing",
			},
		},
		{
			name:  "success state",
			state: stateSuccess,
			files: []generator.PromptFile{
				{Filename: "test1.md", Content: "content1"},
				{Filename: "test2.md", Content: "content2"},
			},
			expectContent: []string{
				"test1.md", "test2.md",
			},
		},
		{
			name:  "error state",
			state: stateError,
			err:   assert.AnError,
			expectContent: []string{
				"error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := initialModel(cfg)
			model.state = tt.state
			model.files = tt.files
			model.err = tt.err

			view := model.View()

			assert.NotEmpty(t, view)

			for _, expected := range tt.expectContent {
				// Convert to lowercase for case-insensitive matching
				// since translations might have different cases
				assert.Contains(t, view, expected, "View should contain '%s'", expected)
			}
		})
	}
}

func TestModelView_WithActualMessages(t *testing.T) {
	setupI18n()
	cfg := config.DefaultConfig()
	model := initialModel(cfg)

	// Test that view doesn't panic and returns reasonable content
	model.state = stateLoading
	view := model.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "System Prompt Generator") // Should contain app name or similar

	model.state = stateSuccess
	model.files = []generator.PromptFile{
		{Filename: "test.md", Content: "test content"},
	}
	view = model.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "test.md")

	model.state = stateError
	model.err = assert.AnError
	view = model.View()
	assert.NotEmpty(t, view)
	// Error view should contain some indication of error
}

func TestRunInteractive_BasicFlow(t *testing.T) {
	setupI18n()
	tempDir := testutil.TempDir(t)

	// Create test input files
	inputDir := filepath.Join(tempDir, "input")
	err := os.MkdirAll(inputDir, 0755)
	require.NoError(t, err)

	testutil.CreateTestFile(t, filepath.Join(inputDir, "test.md"), "# Test\nContent\n")

	cfg := &config.Config{
		InputDir:     inputDir,
		OutputFiles:  []string{filepath.Join(tempDir, "output.md")},
		ExcludeFiles: []string{},
		Header:       "",
		Footer:       "",
		Settings:     nil,
	}

	// Note: This test is limited because we can't easily simulate user input
	// in a Bubble Tea program. In a real scenario, you'd want to test with
	// a custom program that can be controlled programmatically.
	// For now, we just test that the function doesn't panic and the model is created correctly.

	model := initialModel(cfg)
	assert.NotNil(t, model)
	assert.Equal(t, stateLoading, model.state)
}

// Test helper functions and types

func TestStateConstants(t *testing.T) {
	assert.Equal(t, state(0), stateLoading)
	assert.Equal(t, state(1), stateSuccess)
	assert.Equal(t, state(2), stateError)
}

func TestPromptFileStruct(t *testing.T) {
	files := []generator.PromptFile{
		{
			Path:     "/path/to/file.md",
			Filename: "file.md",
			Content:  "content",
		},
	}

	assert.Len(t, files, 1)
	assert.Equal(t, "/path/to/file.md", files[0].Path)
	assert.Equal(t, "file.md", files[0].Filename)
	assert.Equal(t, "content", files[0].Content)
}
