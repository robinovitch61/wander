package constants

import (
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
