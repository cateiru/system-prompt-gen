package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cateiru/system-prompt-gen/internal/config"
	"github.com/cateiru/system-prompt-gen/internal/generator"
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

	s.WriteString(titleStyle.Render("System Prompt Generator"))
	s.WriteString("\n\n")

	switch m.state {
	case stateLoading:
		s.WriteString("📁 プロンプトファイルを収集中...\n")
		s.WriteString("🔄 処理中です...")

	case stateSuccess:
		s.WriteString(infoStyle.Render(fmt.Sprintf(
			"✅ %d個のプロンプトファイルを発見しました\n\n"+
				"📂 入力ディレクトリ: %s\n"+
				"📄 出力ファイル: %s\n\n"+
				"[Enter] 生成実行  [q] 終了",
			len(m.files),
			m.config.InputDir,
			strings.Join(m.config.OutputFiles, ", "),
		)))
		s.WriteString("\n\n")

		s.WriteString("📋 検出されたファイル:\n")
		for _, file := range m.files {
			s.WriteString(fmt.Sprintf("  • %s\n", file.Filename))
		}

	case stateError:
		s.WriteString(errorStyle.Render("❌ エラーが発生しました"))
		s.WriteString("\n\n")
		s.WriteString(fmt.Sprintf("詳細: %v\n\n", m.err))
		s.WriteString("[r] 再試行  [q] 終了")
	}

	return s.String()
}

func RunInteractive(cfg *config.Config) error {
	p := tea.NewProgram(initialModel(cfg))
	_, err := p.Run()
	return err
}