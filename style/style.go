package style

import "github.com/charmbracelet/lipgloss"

var (
	KeyHelpStyle            = lipgloss.NewStyle().Padding(0, 5)
	KeyHelpKeyStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	KeyHelpDescriptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	EditingTextStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("6"))
	HeaderStyle             = lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.RoundedBorder(), true)
	ViewportStyle           = lipgloss.NewStyle().Background(lipgloss.Color("#000000"))
	ViewportCursorRowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("6"))
	ViewportHighlightStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#e760fc"))
)
