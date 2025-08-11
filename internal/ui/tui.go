package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cateiru/system-prompt-gen/internal/config"
	"github.com/cateiru/system-prompt-gen/internal/generator"
	"github.com/cateiru/system-prompt-gen/internal/i18n"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			Width(50)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5733"))
)

type state int

const (
	stateLoading state = iota
	stateSuccess
	stateError
)

type model struct {
	config    *config.Config
	generator *generator.Generator
	files     []generator.PromptFile
	state     state
	err       error
	content   string
}

type generateMsg struct {
	files []generator.PromptFile
	err   error
}

func initialModel(cfg *config.Config) model {
	return model{
		config:    cfg,
		generator: generator.New(cfg),
		state:     stateLoading,
	}
}

func (m model) Init() tea.Cmd {
	return generatePrompts(m.generator)
}

func generatePrompts(gen *generator.Generator) tea.Cmd {
	return func() tea.Msg {
		files, err := gen.CollectPromptFiles()
		return generateMsg{files: files, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter", " ":
			if m.state == stateSuccess {
				if err := m.generator.WriteOutputFiles(m.content); err != nil {
					m.state = stateError
					m.err = err
					return m, nil
				}
				return m, tea.Quit
			}
		case "r":
			if m.state == stateError {
				m.state = stateLoading
				return m, generatePrompts(m.generator)
			}
		}

	case generateMsg:
		if msg.err != nil {
			m.state = stateError
			m.err = msg.err
		} else {
			m.files = msg.files
			m.content = m.generator.GeneratePrompt(msg.files)
			m.state = stateSuccess
		}
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(i18n.T("app_name")))
	s.WriteString("\n\n")

	switch m.state {
	case stateLoading:
		s.WriteString(i18n.T("collecting_files") + "\n")
		s.WriteString(i18n.T("processing"))

	case stateSuccess:
		// 出力ファイル一覧の取得
		outputFiles := m.generator.GetGeneratedTargets()

		s.WriteString(infoStyle.Render(i18n.T("files_found", map[string]interface{}{
			"Count":       len(m.files),
			"InputDir":    m.config.InputDir,
			"OutputFiles": strings.Join(outputFiles, ", "),
		})))
		s.WriteString("\n\n")

		s.WriteString(i18n.T("detected_files") + "\n")
		for _, file := range m.files {
			s.WriteString(fmt.Sprintf("  • %s\n", file.Filename))
		}

	case stateError:
		s.WriteString(errorStyle.Render(i18n.T("error_occurred")))
		s.WriteString("\n\n")
		s.WriteString(i18n.T("error_details", map[string]interface{}{
			"Error": m.err,
		}))
	}

	return s.String()
}

func RunInteractive(cfg *config.Config) error {
	p := tea.NewProgram(initialModel(cfg))
	_, err := p.Run()
	return err
}
