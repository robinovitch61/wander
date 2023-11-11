package constants

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/wander/internal/tui/style"
	"strings"
	"time"
)

const NoVersionString = "built from source"

var LogoString = strings.Join([]string{
	"█ █ █ █▀█ █▄ █ █▀▄ █▀▀ █▀█",
	"▀▄▀▄▀ █▀█ █ ▀█ █▄▀ ██▄ █▀▄",
}, "\n")

const ToastDuration = time.Second * 5

const SaveDialogPlaceholder = "Output file name (path optional)"

const ExecWebSocketClosed = "> connection closed <"

const ExecWebSocketHeartbeatDuration = time.Second * 10

const TableSeparator = "|【=◈︿◈=】|"

const TablePadding = "   "

var JobsTableStatusStyles = map[string]lipgloss.Style{
	TablePadding + "pending" + TablePadding: style.JobRowPending,
	TablePadding + "dead" + TablePadding:    style.JobRowDead,
}

var TasksTableStatusStyles = JobsTableStatusStyles

const DefaultPageInput = "/bin/sh"

// DefaultEventJQQuery is a single line as this shows up verbatim in `wander --help`
const DefaultEventJQQuery = `.Events[] | {"1:Index": .Index, "2:Topic": .Topic, "3:Type": .Type, "4:Name": .Payload | (.Job // .Allocation // .Deployment // .Evaluation) | (.JobID // .ID), "5:ID": .Payload | (.Job.ID // (.Allocation // .Deployment // .Evaluation).ID[:8])}`
