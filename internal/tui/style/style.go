package style

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/wander/internal/tui/constants"
)

const (
	black     = lipgloss.Color("#000000")
	blue      = lipgloss.Color("6")
	greenblue = lipgloss.Color("#00A095")
	pink      = lipgloss.Color("#E760FC")
	darkred   = lipgloss.Color("#FF0000")
	darkgreen = lipgloss.Color("#00FF00")
	grey      = lipgloss.Color("#737373")
	red       = lipgloss.Color("#FF5353")
	yellow    = lipgloss.Color("#DBBD70")
)

type Styles struct {
	Regular                       lipgloss.Style
	Bold                          lipgloss.Style
	Logo                          lipgloss.Style
	ClusterUrl                    lipgloss.Style
	KeyHelp                       lipgloss.Style
	KeyHelpKey                    lipgloss.Style
	KeyHelpDescription            lipgloss.Style
	Header                        lipgloss.Style
	FilterPrefix                  lipgloss.Style
	FilterEditing                 lipgloss.Style
	FilterApplied                 lipgloss.Style
	JobRowPending                 lipgloss.Style
	JobRowDead                    lipgloss.Style
	StatBad                       lipgloss.Style
	PseudoPrompt                  lipgloss.Style
	Viewport                      lipgloss.Style
	ViewportHeaderStyle           lipgloss.Style
	ViewportSelectedRowStyle      lipgloss.Style
	ViewportHighlightStyle        lipgloss.Style
	ViewportSpecialHighlightStyle lipgloss.Style
	ViewportFooterStyle           lipgloss.Style
	SaveDialogPromptStyle         lipgloss.Style
	SaveDialogPlaceholderStyle    lipgloss.Style
	SaveDialogTextStyle           lipgloss.Style
	StdOut                        lipgloss.Style
	StdErr                        lipgloss.Style
	SuccessToast                  lipgloss.Style
	ErrorToast                    lipgloss.Style
}

func NewStyles(renderer *lipgloss.Renderer) Styles {
	var regular = renderer.NewStyle()
	var bold = regular.Copy().Bold(true)
	var logo = regular.Copy().Padding(0, 0).Foreground(yellow)
	var clusterUrl = bold.Copy()
	var keyHelp = regular.Copy().Padding(0, 1)
	var keyHelpKey = regular.Copy().Foreground(blue).Bold(true)
	var keyHelpDescription = regular.Copy()
	var header = regular.Copy().Padding(0, 1).Border(lipgloss.RoundedBorder(), true)
	var filterPrefix = regular.Copy().Padding(0, 3).Border(lipgloss.NormalBorder(), true)
	var filterEditing = regular.Copy().Foreground(black).Background(blue)
	var filterApplied = regular.Copy().Foreground(black).Background(greenblue)
	var jobRowPending = regular.Copy().Foreground(yellow)
	var jobRowDead = regular.Copy().Foreground(red)
	var statBad = regular.Copy().Foreground(black).Background(red)
	var pseudoPrompt = regular.Copy().Background(blue)
	var viewport = regular.Copy()
	var viewportHeaderStyle = bold.Copy()
	var viewportSelectedRowStyle = regular.Copy().Foreground(black).Background(blue)
	var viewportHighlightStyle = regular.Copy().Foreground(black).Background(pink)
	var viewportSpecialHighlightStyle = regular.Copy().Foreground(black).Background(yellow)
	var viewportFooterStyle = regular.Copy().Foreground(grey)
	var saveDialogPromptStyle = regular.Copy().Background(darkred).Foreground(black)
	var saveDialogPlaceholderStyle = regular.Copy().Background(darkred).Foreground(black)
	var saveDialogTextStyle = regular.Copy().Background(darkred).Foreground(black)
	var stdOut = regular.Copy().UnsetForeground()
	var stdErr = regular.Copy().Foreground(red)
	var successToast = bold.Copy().PaddingLeft(1).Foreground(black).Background(darkgreen)
	var errorToast = bold.Copy().PaddingLeft(1).Foreground(black).Background(darkred)

	return Styles{
		Regular:                       regular,
		Bold:                          bold,
		Logo:                          logo,
		ClusterUrl:                    clusterUrl,
		KeyHelp:                       keyHelp,
		KeyHelpKey:                    keyHelpKey,
		KeyHelpDescription:            keyHelpDescription,
		Header:                        header,
		FilterPrefix:                  filterPrefix,
		FilterEditing:                 filterEditing,
		FilterApplied:                 filterApplied,
		JobRowPending:                 jobRowPending,
		JobRowDead:                    jobRowDead,
		StatBad:                       statBad,
		PseudoPrompt:                  pseudoPrompt,
		Viewport:                      viewport,
		ViewportHeaderStyle:           viewportHeaderStyle,
		ViewportSelectedRowStyle:      viewportSelectedRowStyle,
		ViewportHighlightStyle:        viewportHighlightStyle,
		ViewportSpecialHighlightStyle: viewportSpecialHighlightStyle,
		ViewportFooterStyle:           viewportFooterStyle,
		SaveDialogPromptStyle:         saveDialogPromptStyle,
		SaveDialogPlaceholderStyle:    saveDialogPlaceholderStyle,
		SaveDialogTextStyle:           saveDialogTextStyle,
		StdOut:                        stdOut,
		StdErr:                        stdErr,
		SuccessToast:                  successToast,
		ErrorToast:                    errorToast,
	}
}

func GetTableStyles(s Styles) map[string]lipgloss.Style {
	return map[string]lipgloss.Style{
		constants.TablePadding + "pending" + constants.TablePadding: s.JobRowPending,
		constants.TablePadding + "dead" + constants.TablePadding:    s.JobRowDead,
	}
}
