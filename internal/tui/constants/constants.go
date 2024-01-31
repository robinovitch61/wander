package constants

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/wander/internal/tui/style"
	"strings"
	"time"
)

var LogoString = strings.Join([]string{
	"Wander",
}, "\n")

const ToastDuration = time.Second * 5

const SaveDialogPlaceholder = "Output file name (path optional)"

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

// DefaultAllocEventJQQuery is a single line as this shows up verbatim in `wander --help`
const DefaultAllocEventJQQuery = `.Index as $index | .Events[] | .Type as $type | .Payload.Allocation | .DeploymentStatus.Healthy as $healthy | .ClientStatus as $clientStatus | .Name as $allocName | (.TaskStates // {"":{"Events": [{}]}}) | to_entries[] | .key as $k | .value.Events[] | {"0:Index": $index, "1:AllocName": $allocName, "2:TaskName": $k, "3:Type": $type, "4:Time": ((.Time // 0) / 1000000000 | todate), "5:Msg": .DisplayMessage, "6:Healthy": $healthy, "7:ClientStatus": $clientStatus}`

const ConfirmationKey = "Yes"
