package style

import "github.com/charmbracelet/lipgloss"

var (
	Bold                       = lipgloss.NewStyle().Bold(true)
	Logo                       = lipgloss.NewStyle().MarginBottom(1).Padding(0).Foreground(lipgloss.Color("#dbbd70"))
	ClusterUrl                 = lipgloss.NewStyle()
	KeyHelp                    = lipgloss.NewStyle().Padding(0, 2)
	KeyHelpKey                 = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	KeyHelpDescription         = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	Header                     = lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.RoundedBorder(), true)
	Viewport                   = lipgloss.NewStyle().Background(lipgloss.Color("#000000"))
	ViewportHeaderStyle        = lipgloss.NewStyle().Bold(true)
	ViewportCursorRowStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("6"))
	ViewportHighlightStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#e760fc"))
	ViewportFooterStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#737373"))
	SaveDialogPromptStyle      = lipgloss.NewStyle().Background(lipgloss.Color("#FF0000")).Foreground(lipgloss.Color("#000000"))
	SaveDialogPlaceholderStyle = lipgloss.NewStyle().Background(lipgloss.Color("#FF0000")).Foreground(lipgloss.Color("#000000"))
	SaveDialogTextStyle        = lipgloss.NewStyle().Background(lipgloss.Color("#FF0000")).Foreground(lipgloss.Color("#000000"))
	StdOut                     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	StdErr                     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5353"))
	SuccessToast               = lipgloss.NewStyle().Bold(true).PaddingLeft(1).Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#00FF00"))
	ErrorToast                 = lipgloss.NewStyle().Bold(true).PaddingLeft(1).Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#FF0000"))
)
