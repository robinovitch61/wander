package dev

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"os"
)

var debugSet = os.Getenv("WANDER_DEBUG")
var debugPath = os.Getenv("WANDER_DEBUG_PATH")

// dev
func Debug(msg string) {
	if debugPath == "" {
		debugPath = "wander.log"
	}
	if debugSet != "" {
		f, err := tea.LogToFile(debugPath, "")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		log.Printf("%q", msg)
		defer f.Close()
	}
}
