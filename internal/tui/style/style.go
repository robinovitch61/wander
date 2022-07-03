package style

import "github.com/charmbracelet/lipgloss"

var (
	Regular                    = lipgloss.NewStyle()
	Bold                       = Regular.Copy().Bold(true)
	Logo                       = Regular.Copy().Padding(0, 1).Foreground(lipgloss.Color("#dbbd70"))
	ClusterUrl                 = Bold.Copy()
	KeyHelp                    = Regular.Copy().Padding(0, 2)
	KeyHelpKey                 = Regular.Copy().Foreground(lipgloss.Color("6")).Bold(true)
	KeyHelpDescription         = Regular.Copy().Foreground(lipgloss.Color("7"))
	Header                     = Regular.Copy().Padding(0, 1).Border(lipgloss.RoundedBorder(), true)
	PseudoPrompt               = lipgloss.NewStyle().Background(lipgloss.Color("3"))
	Viewport                   = Regular.Copy().Background(lipgloss.Color("#000000"))
	ViewportHeaderStyle        = Bold.Copy()
	ViewportSelectedRowStyle   = Regular.Copy().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("6"))
	ViewportHighlightStyle     = Regular.Copy().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#e760fc"))
	ViewportFooterStyle        = Regular.Copy().Foreground(lipgloss.Color("#737373"))
	SaveDialogPromptStyle      = Regular.Copy().Background(lipgloss.Color("#FF0000")).Foreground(lipgloss.Color("#000000"))
	SaveDialogPlaceholderStyle = Regular.Copy().Background(lipgloss.Color("#FF0000")).Foreground(lipgloss.Color("#000000"))
	SaveDialogTextStyle        = Regular.Copy().Background(lipgloss.Color("#FF0000")).Foreground(lipgloss.Color("#000000"))
	StdOut                     = Regular.Copy().Foreground(lipgloss.Color("#FFFFFF"))
	StdErr                     = Regular.Copy().Foreground(lipgloss.Color("#FF5353"))
	SuccessToast               = Bold.Copy().PaddingLeft(1).Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#00FF00"))
	ErrorToast                 = Bold.Copy().PaddingLeft(1).Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#FF0000"))
)
