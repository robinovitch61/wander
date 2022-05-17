package style

import "github.com/charmbracelet/lipgloss"

var (
	Bold               = lipgloss.NewStyle().Bold(true)
	Logo               = lipgloss.NewStyle().Padding(1).Foreground(lipgloss.Color("#dbbd70"))
	KeyHelp            = lipgloss.NewStyle().Padding(0, 2)
	KeyHelpKey         = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	KeyHelpDescription = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	Header             = lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.RoundedBorder(), true)
	Viewport           = lipgloss.NewStyle().Background(lipgloss.Color("#000000"))
	ViewportCursorRow  = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("6"))
	ViewportHighlight  = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#e760fc"))
	TableHeaderStyle   = lipgloss.NewStyle().Bold(true)
)
