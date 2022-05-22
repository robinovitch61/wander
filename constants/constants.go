package constants

import (
	"strings"
	"time"
)

const (
	NomadTokenEnvVariable = "NOMAD_TOKEN"
	NomadUrlEnvVariable   = "NOMAD_ADDR"
)

var Logo = strings.Join([]string{
	"█ █ █ █▀█ █▄ █ █▀▄ █▀▀ █▀█",
	"▀▄▀▄▀ █▀█ █ ▀█ █▄▀ ██▄ █▀▄",
}, "\n")

const ToastDuration = time.Second * 5
