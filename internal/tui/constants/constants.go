package constants

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/wander/internal/tui/style"
	"strings"
	"time"
)

var LogoString = strings.Join([]string{
	"█ █ █ █▀█ █▄ █ █▀▄ █▀▀ █▀█",
	"▀▄▀▄▀ █▀█ █ ▀█ █▄▀ ██▄ █▀▄",
}, "\n")

const ToastDuration = time.Second * 5

const SaveDialogPlaceholder = "Output file name (path optional)"

const ExecWebSocketClosed = "> connection closed <"

const ExecWebSocketHeartbeatDuration = time.Second * 10

const TablePadding = "    "

var JobsViewportConditionalStyle = map[string]lipgloss.Style{
	TablePadding + "pending" + TablePadding: style.JobRowPending,
	TablePadding + "dead" + TablePadding:    style.JobRowDead,
}

var AllocationsViewportConditionalStyle = JobsViewportConditionalStyle
