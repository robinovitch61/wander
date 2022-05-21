package style

import "github.com/charmbracelet/lipgloss"

var (
	Bold               = lipgloss.NewStyle().Bold(true)
	Logo               = lipgloss.NewStyle().MarginBottom(1).Foreground(lipgloss.Color("#dbbd70"))
	ClusterUrl         = lipgloss.NewStyle().Padding(0, 0, 0, 1)
	KeyHelp            = lipgloss.NewStyle().Padding(0, 2)
	KeyHelpKey         = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	KeyHelpDescription = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	Header             = lipgloss.NewStyle().Padding(0, 1).Border(lipgloss.RoundedBorder(), true)
	Viewport           = lipgloss.NewStyle().Background(lipgloss.Color("#000000"))
	StdOut             = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	StdErr             = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5353"))
	SuccessToast       = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#00FF00"))
	ErrorToast         = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#FF0000"))
)
