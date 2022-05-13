package style

import "github.com/charmbracelet/lipgloss"

var (
	Bold               = lipgloss.NewStyle().Bold(true)
	KeyHelp            = lipgloss.NewStyle().Padding(0, 5)
	KeyHelpKey         = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	KeyHelpDescription = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	EditingText        = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("6"))
	Header             = lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.RoundedBorder(), true)
	Viewport           = lipgloss.NewStyle().Background(lipgloss.Color("#000000"))
	ViewportCursorRow  = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("6"))
	ViewportHighlight  = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#e760fc"))
	TableHeaderStyle   = lipgloss.NewStyle().Bold(true)
)
