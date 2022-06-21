package constants

import (
	"strings"
	"time"
)

const (
	WanderSSHHost = "WANDER_SSH_HOST"
	WanderSSHPort = "WANDER_SSH_PORT"
)

var LogoString = strings.Join([]string{
	"█ █ █ █▀█ █▄ █ █▀▄ █▀▀ █▀█",
	"▀▄▀▄▀ █▀█ █ ▀█ █▄▀ ██▄ █▀▄",
}, "\n")

const ToastDuration = time.Second * 5

const SaveDialogPlaceholder = "Output file name (path optional)"
