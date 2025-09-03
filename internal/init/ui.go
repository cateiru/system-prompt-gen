package init

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cateiru/system-prompt-gen/internal/config"
	"github.com/cateiru/system-prompt-gen/internal/i18n"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	listStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			Width(80)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	unselectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5733"))
)

type uiState int

const (
	stateOverwriteConfirm uiState = iota
	stateFileSelection
	stateToolSelection
	stateConfirmation
	stateProcessing
	stateSuccess
	stateError
)

type initModel struct {
	state         uiState
	initState     *InitState
	cursor        int
	fileSelection map[int]bool
	toolSelection map[int]bool
	allTools      []string
	err           error
}

// runInteractiveInit はインタラクティブな初期化UIを実行する
func runInteractiveInit(initState *InitState) error {
	// DefaultKnownToolFileNamesからツール名を取得してソート
	var allTools []string
	for toolName := range config.DefaultKnownToolFileNames {
		allTools = append(allTools, toolName)
	}
	sort.Strings(allTools)

	model := initModel{
		initState:     initState,
		fileSelection: make(map[int]bool),
		toolSelection: make(map[int]bool),
		allTools:      allTools,
	}

	// 初期状態を設定
	if initState.OverwriteConfirmed {
		if len(initState.ExistingFiles) > 0 {
			model.state = stateFileSelection
		} else {
			model.state = stateToolSelection
		}
	} else {
		model.state = stateOverwriteConfirm
	}

	p := tea.NewProgram(model)
	_, err := p.Run()
	return err
}

func (m initModel) Init() tea.Cmd {
	return nil
}

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			maxCursor := m.getMaxCursor()
			if m.cursor < maxCursor {
				m.cursor++
			}

		case " ":
			m.toggleSelection()

		case "enter":
			return m.handleEnter()
		}
	}

	return m, nil
}

func (m initModel) getMaxCursor() int {
	switch m.state {
	case stateOverwriteConfirm:
		return 1 // Yes/No
	case stateFileSelection:
		return len(m.initState.ExistingFiles) - 1
	case stateToolSelection:
		return len(m.allTools) - 1
	case stateConfirmation:
		return 1 // Proceed/Cancel
	default:
		return 0
	}
}

func (m initModel) toggleSelection() {
	switch m.state {
	case stateFileSelection:
		m.fileSelection[m.cursor] = !m.fileSelection[m.cursor]
	case stateToolSelection:
		m.toolSelection[m.cursor] = !m.toolSelection[m.cursor]
	}
}

func (m initModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateOverwriteConfirm:
		if m.cursor == 0 { // Yes
			m.initState.OverwriteConfirmed = true
			if len(m.initState.ExistingFiles) > 0 {
				m.state = stateFileSelection
			} else {
				m.state = stateToolSelection
			}
		} else { // No
			return m, tea.Quit
		}

	case stateFileSelection:
		// 選択されたファイルを収集
		var selectedFiles []ExistingFile
		for i, selected := range m.fileSelection {
			if selected && i < len(m.initState.ExistingFiles) {
				selectedFiles = append(selectedFiles, m.initState.ExistingFiles[i])
			}
		}
		m.initState.SelectedFiles = selectedFiles
		m.state = stateToolSelection
		m.cursor = 0

	case stateToolSelection:
		// 選択されたツールを収集
		var selectedTools []string
		for i, selected := range m.toolSelection {
			if selected && i < len(m.allTools) {
				selectedTools = append(selectedTools, m.allTools[i])
			}
		}
		m.initState.SelectedTools = selectedTools
		m.state = stateConfirmation
		m.cursor = 0

	case stateConfirmation:
		if m.cursor == 0 { // Proceed
			return m.processInit()
		} else { // Cancel
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m initModel) processInit() (tea.Model, tea.Cmd) {
	m.state = stateProcessing

	// ディレクトリ作成
	if err := m.initState.CreateSystemPromptDir(); err != nil {
		m.state = stateError
		m.err = err
		return m, nil
	}

	// ファイル作成
	if err := m.initState.WriteDefaultFile(); err != nil {
		m.state = stateError
		m.err = err
		return m, nil
	}

	if err := m.initState.WriteSettingsFile(); err != nil {
		m.state = stateError
		m.err = err
		return m, nil
	}

	m.state = stateSuccess
	return m, tea.Quit
}

func (m initModel) View() string {
	switch m.state {
	case stateOverwriteConfirm:
		return m.renderOverwriteConfirm()
	case stateFileSelection:
		return m.renderFileSelection()
	case stateToolSelection:
		return m.renderToolSelection()
	case stateConfirmation:
		return m.renderConfirmation()
	case stateProcessing:
		return m.renderProcessing()
	case stateSuccess:
		return m.renderSuccess()
	case stateError:
		return m.renderError()
	default:
		return "Unknown state"
	}
}

func (m initModel) renderOverwriteConfirm() string {
	options := []string{
		fmt.Sprintf("%s\n", i18n.T("init_overwrite_message")),
		fmt.Sprintf("%s %s", m.getCursor(0), i18n.T("yes")),
		fmt.Sprintf("%s %s", m.getCursor(1), i18n.T("no")),
	}

	content := strings.Join(options, "\n")

	return fmt.Sprintf("%s\n\n%s\n",
		listStyle.Render(content),
		i18n.T("init_navigation_help"),
	)
}

func (m initModel) renderFileSelection() string {
	if len(m.initState.ExistingFiles) == 0 {
		return fmt.Sprintf("%s\n", i18n.T("init_no_files_found"))
	}

	options := []string{
		fmt.Sprintf("%s\n", i18n.T("init_file_selection_message")),
	}

	for i, file := range m.initState.ExistingFiles {
		selected := ""
		if m.fileSelection[i] {
			selected = "✓"
		} else {
			selected = " "
		}
		options = append(options, fmt.Sprintf("%s [%s] %s (%s)", m.getCursor(i), selected, file.Path, file.ToolName))
	}

	content := strings.Join(options, "\n")

	return fmt.Sprintf("%s\n\n%s\n",
		listStyle.Render(content),
		i18n.T("init_selection_help"),
	)
}

func (m initModel) renderToolSelection() string {
	options := []string{
		fmt.Sprintf("%s\n", i18n.T("init_tool_selection_message")),
	}

	for i, tool := range m.allTools {
		selected := ""
		if m.toolSelection[i] {
			selected = "✓"
		} else {
			selected = " "
		}
		options = append(options, fmt.Sprintf("%s [%s] %s", m.getCursor(i), selected, tool))
	}

	content := strings.Join(options, "\n")

	return fmt.Sprintf("%s\n\n%s\n",
		listStyle.Render(content),
		i18n.T("init_selection_help"),
	)
}

func (m initModel) renderConfirmation() string {
	var details []string

	// 選択されたファイル
	if len(m.initState.SelectedFiles) > 0 {
		details = append(details, i18n.T("init_selected_files")+":")
		for _, file := range m.initState.SelectedFiles {
			details = append(details, fmt.Sprintf("\t◯ %s", file.Path))
		}
	} else {
		details = append(details, i18n.T("init_no_files_selected"))
	}

	details = append(details, "")

	// 選択されたツール
	if len(m.initState.SelectedTools) > 0 {
		details = append(details, i18n.T("init_selected_tools")+":")
		for _, tool := range m.initState.SelectedTools {
			details = append(details, fmt.Sprintf("\t● %s", tool))
		}
	} else {
		details = append(details, i18n.T("init_no_tools_selected"))
	}

	options := []string{
		fmt.Sprintf("%s\n", i18n.T("init_confirmation_message")),
		strings.Join(details, "\n"),
		"\n\n",
		fmt.Sprintf("%s %s", m.getCursor(0), i18n.T("proceed")),
		fmt.Sprintf("%s %s", m.getCursor(1), i18n.T("cancel")),
	}

	return fmt.Sprintf("%s\n\n%s\n",
		listStyle.Render(strings.Join(options, "\n")),
		i18n.T("init_navigation_help"),
	)
}

func (m initModel) renderProcessing() string {
	title := titleStyle.Render(i18n.T("init_processing_title"))
	return fmt.Sprintf("%s\n\n%s\n", title, i18n.T("init_processing_message"))
}

func (m initModel) renderSuccess() string {
	title := successStyle.Render(i18n.T("init_success_title"))
	return fmt.Sprintf("%s\n\n%s\n", title, i18n.T("init_success_message"))
}

func (m initModel) renderError() string {
	title := errorStyle.Render(i18n.T("init_error_title"))
	return fmt.Sprintf("%s\n\n%s: %v\n", title, i18n.T("init_error_message"), m.err)
}

func (m initModel) getCursor(index int) string {
	if m.cursor == index {
		return selectedStyle.Render("►")
	}
	return unselectedStyle.Render(" ")
}
