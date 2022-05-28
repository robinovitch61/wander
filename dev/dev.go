package dev

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"os"
)

var debugSet = os.Getenv("WANDER_DEBUG")

func Debug(msg string) {
	if debugSet != "" {
		f, err := tea.LogToFile("wander.log", "")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		log.Printf("%q", msg)
		defer f.Close()
	}
}

func DebugMsg(source string, msg tea.Msg) {
	msgTypeString := fmt.Sprintf("%T", msg)
	// too many of these
	if msgTypeString != "textinput.blinkMsg" {
		Debug(source + " " + msgTypeString)
	}
}
